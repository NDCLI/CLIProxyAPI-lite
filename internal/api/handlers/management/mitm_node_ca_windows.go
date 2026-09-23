//go:build windows

package management

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const mitmNodeCAEnv = "NODE_EXTRA_CA_CERTS"

const (
	mitmNodeCAStateEnv    = "CLIPROXYAPI_MITM_NODE_CA_PATH"
	mitmNodeCAPreviousEnv = "CLIPROXYAPI_PREVIOUS_NODE_EXTRA_CA_CERTS"
)

func hiddenMITMCommand(name string, args ...string) *exec.Cmd {
	command := exec.Command(name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command
}

func setMITMNodeCA(certPath string) error {
	absoluteCertPath, errAbs := filepath.Abs(certPath)
	if errAbs != nil {
		return fmt.Errorf("resolve CLIProxy MITM CA path: %w", errAbs)
	}
	certPath = absoluteCertPath
	key, errOpen := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if errOpen != nil {
		return fmt.Errorf("open user environment for MITM CA: %w", errOpen)
	}
	defer key.Close()
	current, _, errRead := key.GetStringValue(mitmNodeCAEnv)
	if errRead != nil && !errors.Is(errRead, registry.ErrNotExist) {
		return fmt.Errorf("read Node CA trust setting: %w", errRead)
	}
	state, _, errState := key.GetStringValue(mitmNodeCAStateEnv)
	if errState == nil && strings.EqualFold(current, state) && strings.EqualFold(state, certPath) {
		return nil
	}
	if current != "" && strings.EqualFold(current, certPath) {
		return nil
	}
	if errState == nil && !strings.EqualFold(current, state) {
		removeMITMNodeCABundle(certPath, state)
		_ = key.DeleteValue(mitmNodeCAStateEnv)
		_ = key.DeleteValue(mitmNodeCAPreviousEnv)
	}

	targetPath := certPath
	if current != "" {
		bundle, errBundle := combineMITMNodeCACertificates(current, certPath)
		if errBundle != nil {
			return errBundle
		}
		file, errCreate := os.CreateTemp(filepath.Dir(certPath), "cliproxy-node-ca-*.pem")
		if errCreate != nil {
			return fmt.Errorf("create combined Node CA bundle: %w", errCreate)
		}
		targetPath = file.Name()
		if errChmod := file.Chmod(0o600); errChmod != nil {
			_ = file.Close()
			_ = os.Remove(targetPath)
			return fmt.Errorf("protect combined Node CA bundle: %w", errChmod)
		}
		if _, errWrite := file.Write(bundle); errWrite != nil {
			_ = file.Close()
			_ = os.Remove(targetPath)
			return fmt.Errorf("write combined Node CA bundle: %w", errWrite)
		}
		if errClose := file.Close(); errClose != nil {
			_ = os.Remove(targetPath)
			return fmt.Errorf("close combined Node CA bundle: %w", errClose)
		}
	}
	if errSet := key.SetStringValue(mitmNodeCAPreviousEnv, current); errSet != nil {
		removeMITMNodeCABundle(certPath, targetPath)
		return fmt.Errorf("preserve existing Node CA trust setting: %w", errSet)
	}
	if errSet := key.SetStringValue(mitmNodeCAStateEnv, targetPath); errSet != nil {
		_ = key.DeleteValue(mitmNodeCAPreviousEnv)
		removeMITMNodeCABundle(certPath, targetPath)
		return fmt.Errorf("record Node MITM CA trust setting: %w", errSet)
	}
	if errSet := key.SetStringValue(mitmNodeCAEnv, targetPath); errSet != nil {
		_ = key.DeleteValue(mitmNodeCAStateEnv)
		_ = key.DeleteValue(mitmNodeCAPreviousEnv)
		removeMITMNodeCABundle(certPath, targetPath)
		return fmt.Errorf("set Node MITM CA trust: %w", errSet)
	}
	broadcastMITMEnvironmentChange()
	return nil
}

func clearMITMNodeCA(certPath string) error {
	absoluteCertPath, errAbs := filepath.Abs(certPath)
	if errAbs != nil {
		return fmt.Errorf("resolve CLIProxy MITM CA path: %w", errAbs)
	}
	certPath = absoluteCertPath
	key, errOpen := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if errOpen != nil {
		return fmt.Errorf("open user environment for MITM CA cleanup: %w", errOpen)
	}
	defer key.Close()
	state, _, errState := key.GetStringValue(mitmNodeCAStateEnv)
	if errState != nil {
		return nil
	}
	current, _, errRead := key.GetStringValue(mitmNodeCAEnv)
	previous, _, errPrevious := key.GetStringValue(mitmNodeCAPreviousEnv)
	if errPrevious != nil {
		previous = ""
	}
	if errRead == nil && strings.EqualFold(current, state) {
		if previous == "" {
			if errDelete := key.DeleteValue(mitmNodeCAEnv); errDelete != nil && !errors.Is(errDelete, registry.ErrNotExist) {
				return fmt.Errorf("remove Node MITM CA trust: %w", errDelete)
			}
		} else if errRestore := key.SetStringValue(mitmNodeCAEnv, previous); errRestore != nil {
			return fmt.Errorf("restore previous Node CA trust setting: %w", errRestore)
		}
	}
	removeMITMNodeCABundle(certPath, state)
	_ = key.DeleteValue(mitmNodeCAStateEnv)
	_ = key.DeleteValue(mitmNodeCAPreviousEnv)
	broadcastMITMEnvironmentChange()
	return nil
}

func broadcastMITMEnvironmentChange() {
	environment, errEnvironment := syscall.UTF16PtrFromString("Environment")
	if errEnvironment != nil {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	_, _, _ = user32.NewProc("SendMessageTimeoutW").Call(
		0xffff, // HWND_BROADCAST
		0x001a, // WM_SETTINGCHANGE
		0,
		uintptr(unsafe.Pointer(environment)),
		0x0002, // SMTO_ABORTIFHUNG
		5000,
		0,
	)
}

func combineMITMNodeCACertificates(existingPath, mitmPath string) ([]byte, error) {
	existing, errRead := os.ReadFile(existingPath)
	if errRead != nil {
		return nil, fmt.Errorf("read existing NODE_EXTRA_CA_CERTS bundle: %w", errRead)
	}
	mitm, errRead := os.ReadFile(mitmPath)
	if errRead != nil {
		return nil, fmt.Errorf("read CLIProxy MITM CA certificate: %w", errRead)
	}

	var bundle bytes.Buffer
	count := 0
	for rest := existing; len(rest) > 0; {
		block, remaining := pem.Decode(rest)
		if block == nil {
			if strings.TrimSpace(string(rest)) != "" {
				return nil, fmt.Errorf("existing NODE_EXTRA_CA_CERTS file is not a valid PEM certificate bundle")
			}
			break
		}
		rest = remaining
		if block.Type != "CERTIFICATE" {
			continue
		}
		if _, errParse := x509.ParseCertificate(block.Bytes); errParse != nil {
			return nil, fmt.Errorf("existing NODE_EXTRA_CA_CERTS contains an invalid certificate: %w", errParse)
		}
		if errWrite := pem.Encode(&bundle, block); errWrite != nil {
			return nil, fmt.Errorf("copy existing Node CA certificate: %w", errWrite)
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("existing NODE_EXTRA_CA_CERTS file contains no PEM certificates")
	}
	block, _ := pem.Decode(mitm)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("CLIProxy MITM CA is not a PEM certificate")
	}
	if _, errParse := x509.ParseCertificate(block.Bytes); errParse != nil {
		return nil, fmt.Errorf("parse CLIProxy MITM CA certificate: %w", errParse)
	}
	if bundle.Len() > 0 && !bytes.HasSuffix(bundle.Bytes(), []byte("\n")) {
		bundle.WriteByte('\n')
	}
	if _, errWrite := bundle.Write(mitm); errWrite != nil {
		return nil, fmt.Errorf("append CLIProxy MITM CA certificate: %w", errWrite)
	}
	return bundle.Bytes(), nil
}

func removeMITMNodeCABundle(certPath, bundlePath string) {
	if bundlePath == "" || strings.EqualFold(bundlePath, certPath) ||
		!strings.EqualFold(filepath.Dir(bundlePath), filepath.Dir(certPath)) ||
		!strings.HasPrefix(strings.ToLower(filepath.Base(bundlePath)), "cliproxy-node-ca-") {
		return
	}
	_ = os.Remove(bundlePath)
}
