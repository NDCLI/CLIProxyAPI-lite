//go:build windows

package management

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const mitmNodeCAEnv = "NODE_EXTRA_CA_CERTS"

func hiddenMITMCommand(name string, args ...string) *exec.Cmd {
	command := exec.Command(name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return command
}

func setMITMNodeCA(certPath string) error {
	key, errOpen := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if errOpen != nil {
		return fmt.Errorf("open user environment for MITM CA: %w", errOpen)
	}
	defer key.Close()
	current, _, errRead := key.GetStringValue(mitmNodeCAEnv)
	if errRead == nil && current != "" && !strings.EqualFold(current, certPath) {
		return fmt.Errorf("NODE_EXTRA_CA_CERTS already points to another certificate; configure Node trust manually before enabling MITM")
	}
	if errSet := key.SetStringValue(mitmNodeCAEnv, certPath); errSet != nil {
		return fmt.Errorf("set Node MITM CA trust: %w", errSet)
	}
	return nil
}

func clearMITMNodeCA(certPath string) error {
	key, errOpen := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if errOpen != nil {
		return fmt.Errorf("open user environment for MITM CA cleanup: %w", errOpen)
	}
	defer key.Close()
	current, _, errRead := key.GetStringValue(mitmNodeCAEnv)
	if errRead != nil || !strings.EqualFold(current, certPath) {
		return nil
	}
	if errDelete := key.DeleteValue(mitmNodeCAEnv); errDelete != nil {
		return fmt.Errorf("remove Node MITM CA trust: %w", errDelete)
	}
	return nil
}
