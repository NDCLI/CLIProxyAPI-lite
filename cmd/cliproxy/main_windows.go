//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

func main() {
	if err := launchServer(); err != nil {
		text, _ := windows.UTF16PtrFromString(err.Error())
		caption, _ := windows.UTF16PtrFromString("Lumina")
		_, _ = windows.MessageBox(0, text, caption, windows.MB_OK|windows.MB_ICONERROR)
	}
}

func launchServer() error {
	launcherPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find Lumina launcher: %w", err)
	}
	installDir := filepath.Dir(launcherPath)
	serverPath := filepath.Join(installDir, "cli-proxy-api.exe")
	if _, err := os.Stat(serverPath); err != nil {
		return fmt.Errorf("find server binary %q: %w", serverPath, err)
	}
	dataDir := os.Getenv("LUMINA_DATA_DIR")
	if dataDir == "" {
		dataDir = installDir
	}
	configPath := filepath.Join(dataDir, "config.yaml")
	args := []string{"--tray", "--config", configPath}
	args = append(args, os.Args[1:]...)
	server := exec.Command(serverPath, args...)
	server.Dir = dataDir
	server.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008 | 0x00000200,
		HideWindow:    true,
	}
	if err := server.Start(); err != nil {
		return fmt.Errorf("start Lumina server: %w", err)
	}
	_ = server.Process.Release()
	return nil
}
