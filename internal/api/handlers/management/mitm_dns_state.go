package management

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func (m *mitmRuntime) dnsStatePath() string {
	return filepath.Join(m.baseDir, "certs", "mitm-dns-state.json")
}

func (m *mitmRuntime) loadDNSState() {
	data, errRead := os.ReadFile(m.dnsStatePath())
	if errRead != nil {
		return
	}
	var state map[string]bool
	if json.Unmarshal(data, &state) != nil {
		return
	}
	for tool, enabled := range state {
		if enabled && len(mitmToolHosts[tool]) > 0 {
			m.dnsDesired[tool] = true
		}
	}
}

func (m *mitmRuntime) saveDNSStateLocked() error {
	path := m.dnsStatePath()
	if len(m.dnsDesired) == 0 {
		if errRemove := os.Remove(path); errRemove != nil && !errors.Is(errRemove, os.ErrNotExist) {
			return errRemove
		}
		return nil
	}
	if errMkdir := os.MkdirAll(filepath.Dir(path), 0o700); errMkdir != nil {
		return errMkdir
	}
	data, errMarshal := json.MarshalIndent(m.dnsDesired, "", "  ")
	if errMarshal != nil {
		return errMarshal
	}
	tmp, errCreate := os.CreateTemp(filepath.Dir(path), ".mitm-dns-state-*")
	if errCreate != nil {
		return errCreate
	}
	defer os.Remove(tmp.Name())
	if _, errWrite := tmp.Write(data); errWrite != nil {
		_ = tmp.Close()
		return errWrite
	}
	if errClose := tmp.Close(); errClose != nil {
		return errClose
	}
	return os.Rename(tmp.Name(), path)
}

func (m *mitmRuntime) setDNSPreference(tool string, enabled bool) error {
	if len(mitmToolHosts[tool]) == 0 {
		return fmt.Errorf("unsupported MITM tool %q", tool)
	}
	m.dnsMu.Lock()
	defer m.dnsMu.Unlock()
	previous, existed := m.dnsDesired[tool]
	if enabled {
		m.dnsDesired[tool] = true
	} else {
		delete(m.dnsDesired, tool)
	}
	if errSave := m.saveDNSStateLocked(); errSave != nil {
		if existed {
			m.dnsDesired[tool] = previous
		} else {
			delete(m.dnsDesired, tool)
		}
		return errSave
	}
	return nil
}

func (m *mitmRuntime) setDNS(tool string, enabled bool) error {
	m.dnsOpMu.Lock()
	defer m.dnsOpMu.Unlock()
	m.dnsMu.RLock()
	previous := m.dnsDesired[tool]
	m.dnsMu.RUnlock()
	if errPreference := m.setDNSPreference(tool, enabled); errPreference != nil {
		return errPreference
	}
	if errSet := setMITMDNS(tool, enabled); errSet != nil {
		if errRestore := m.setDNSPreference(tool, previous); errRestore != nil {
			return errors.Join(errSet, fmt.Errorf("restore saved DNS preference: %w", errRestore))
		}
		return errSet
	}
	m.setDNSWarning("")
	return nil
}

func (m *mitmRuntime) restoreDNSPreferences() error {
	m.dnsOpMu.Lock()
	defer m.dnsOpMu.Unlock()
	m.dnsMu.RLock()
	state := make(map[string]bool, len(m.dnsDesired))
	for tool, enabled := range m.dnsDesired {
		state[tool] = enabled
	}
	m.dnsMu.RUnlock()

	var restoreErrors []error
	for tool, enabled := range state {
		if enabled {
			if errSet := setMITMDNS(tool, true); errSet != nil {
				restoreErrors = append(restoreErrors, fmt.Errorf("%s: %w", tool, errSet))
			}
		}
	}
	errRestore := errors.Join(restoreErrors...)
	if errRestore != nil {
		m.setDNSWarning(errRestore.Error())
		return errRestore
	}
	m.setDNSWarning("")
	return nil
}

func (m *mitmRuntime) disableDNSPreferences() error {
	m.dnsOpMu.Lock()
	defer m.dnsOpMu.Unlock()
	m.dnsMu.Lock()
	previous := m.dnsDesired
	m.dnsDesired = make(map[string]bool)
	if errSave := m.saveDNSStateLocked(); errSave != nil {
		m.dnsDesired = previous
		m.dnsMu.Unlock()
		return errSave
	}
	m.dnsMu.Unlock()
	if errClear := cleanupMITMDNSHosts(); errClear != nil {
		m.setDNSWarning(errClear.Error())
		return errClear
	}
	m.setDNSWarning("")
	return nil
}

func (m *mitmRuntime) cleanupDNSForShutdown(ctx context.Context) error {
	m.dnsOpMu.Lock()
	defer m.dnsOpMu.Unlock()
	errDNS := cleanupMITMDNSHosts()
	errStop := m.stop(ctx)
	if errDNS != nil {
		m.setDNSWarning(errDNS.Error())
	} else {
		m.setDNSWarning("")
	}
	return errors.Join(errDNS, errStop)
}

func (m *mitmRuntime) recoverStaleDNS() error {
	m.dnsOpMu.Lock()
	defer m.dnsOpMu.Unlock()
	if runtime.GOOS != "windows" {
		return nil
	}
	conn, errDial := net.DialTimeout("tcp", mitmListenAddress, 250*time.Millisecond)
	if errDial == nil {
		_ = conn.Close()
		return nil
	}
	if errClear := cleanupMITMDNSHosts(); errClear != nil {
		m.setDNSWarning(errClear.Error())
		return errClear
	}
	m.setDNSWarning("")
	return nil
}

func cleanupMITMDNSHosts() error {
	if runtime.GOOS != "windows" {
		return nil
	}
	path := windowsHostsPath()
	data, errRead := os.ReadFile(path)
	if errors.Is(errRead, os.ErrNotExist) {
		return nil
	}
	if errRead != nil {
		return errRead
	}
	updated, changed, errCleanup := cleanupMITMDNSHostsContent(string(data))
	if errCleanup != nil || !changed {
		return errCleanup
	}
	if !currentProcessIsAdmin() {
		return fmt.Errorf("administrator privileges required to clean stale MITM DNS redirects")
	}
	if errWrite := os.WriteFile(path, []byte(updated), 0o644); errWrite != nil {
		return errWrite
	}
	_ = hiddenMITMCommand("ipconfig", "/flushdns").Run()
	return nil
}

func cleanupMITMDNSHostsContent(content string) (string, bool, error) {
	known := make(map[string]struct{})
	for _, hosts := range mitmToolHosts {
		for _, host := range hosts {
			known[strings.ToLower(host)] = struct{}{}
		}
	}
	owned := false
	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, "#", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[1]) != "CLIProxyAPI-lite MITM" {
			continue
		}
		fields := strings.Fields(parts[0])
		for _, host := range fields[1:] {
			if _, ok := known[strings.ToLower(host)]; ok {
				owned = true
				break
			}
		}
		if owned {
			break
		}
	}
	if !owned {
		return content, false, nil
	}
	updated := content
	for _, hosts := range mitmToolHosts {
		next, errUpdate := updateMITMHosts(updated, hosts, false)
		if errUpdate != nil {
			return content, false, errUpdate
		}
		updated = next
	}
	return updated, updated != content, nil
}

func (m *mitmRuntime) dnsStateSnapshot() (map[string]bool, map[string]bool, string) {
	actual := mitmDNSStatus()
	pending := make(map[string]bool)
	m.dnsMu.RLock()
	for tool, enabled := range m.dnsDesired {
		if enabled && !actual[tool] {
			pending[tool] = true
		}
	}
	warning := m.dnsWarning
	m.dnsMu.RUnlock()
	return actual, pending, warning
}

func (m *mitmRuntime) setDNSWarning(warning string) {
	m.dnsMu.Lock()
	m.dnsWarning = warning
	m.dnsMu.Unlock()
}

func (h *Handler) RecoverStaleMITMDNS() error {
	if h == nil || h.mitm == nil {
		return nil
	}
	return h.mitm.recoverStaleDNS()
}

func (h *Handler) ShutdownMITM(ctx context.Context) error {
	if h == nil || h.mitm == nil {
		return nil
	}
	errShutdown := h.mitm.cleanupDNSForShutdown(ctx)
	certPath, _ := h.mitmCAPaths()
	errTrust := clearMITMNodeCA(certPath)
	return errors.Join(errShutdown, errTrust)
}
