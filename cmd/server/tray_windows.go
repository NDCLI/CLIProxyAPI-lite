//go:build windows

package main

import (
	"bufio"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/browser"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/cmd"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/pluginhost"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

//go:embed tray.ps1
var trayPowerShellScript string

//go:embed tray.ico
var trayIcon []byte

func acquireTrayInstance() (func(), bool, error) {
	name, err := windows.UTF16PtrFromString(`Local\CLIProxyAPI-Tray`)
	if err != nil {
		return nil, false, fmt.Errorf("encode tray lock name: %w", err)
	}
	handle, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		_ = windows.CloseHandle(handle)
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("create tray instance mutex: %w", err)
	}
	return func() { _ = windows.CloseHandle(handle) }, true, nil
}

func runTrayMode(cfg *config.Config, configFilePath, localPassword string, host *pluginhost.Host, serverOptions ...api.ServerOption) error {
	controller, events, waitController, err := startTrayController()
	if err != nil {
		return err
	}
	controllerWaited := false
	defer func() {
		if !controllerWaited {
			_ = controller.Process.Kill()
			_ = waitController()
		}
	}()

	select {
	case event, ok := <-events:
		if !ok {
			controllerWaited = true
			return fmt.Errorf("tray controller exited before creating the tray icon: %w", waitController())
		}
		if event == "already-running" {
			controllerWaited = true
			_ = waitController()
			return nil
		}
		if event != "ready" {
			return fmt.Errorf("unexpected tray controller event %q", event)
		}
	case <-time.After(15 * time.Second):
		return fmt.Errorf("tray controller did not become ready")
	}

	cancel, done := cmd.StartServiceBackgroundWithPluginHost(cfg, configFilePath, localPassword, host, serverOptions...)
	serviceStopped := false
	stopService := func() {
		if serviceStopped {
			return
		}
		cancel()
		<-done
		serviceStopped = true
	}
	defer stopService()

	if err := waitForTrayServer(cfg, done); err != nil {
		return err
	}
	log.Infof("server is running in the system tray; management panel: %s", trayManagementURL(cfg))

	for {
		select {
		case event, ok := <-events:
			if !ok {
				controllerWaited = true
				stopService()
				return fmt.Errorf("tray controller exited: %w", waitController())
			}
			switch event {
			case "open":
				if err := browser.OpenURL(trayManagementURL(cfg)); err != nil {
					log.Errorf("failed to open management panel: %v", err)
				}
			case "quit":
				stopService()
				controllerWaited = true
				if err := waitController(); err != nil {
					return fmt.Errorf("tray controller exited: %w", err)
				}
				log.Info("server stopped from the system tray")
				return nil
			}
		case <-done:
			serviceStopped = true
			controllerWaited = true
			_ = controller.Process.Kill()
			_ = waitController()
			return fmt.Errorf("proxy service stopped unexpectedly")
		}
	}
}

func startTrayController() (*exec.Cmd, <-chan string, func() error, error) {
	powershell := filepath.Join(strings.TrimSpace(os.Getenv("SystemRoot")), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if _, err := exec.LookPath(powershell); err != nil {
		powershell, err = exec.LookPath("powershell.exe")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("find Windows PowerShell: %w", err)
		}
	}

	command := exec.Command(powershell, "-NoLogo", "-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", encodePowerShell(trayPowerShellScript))
	command.Env = append(os.Environ(), "CLIPROXY_TRAY_ICON="+base64.StdEncoding.EncodeToString(trayIcon))
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	command.Stderr = log.StandardLogger().Out
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open tray controller output: %w", err)
	}
	if err := command.Start(); err != nil {
		return nil, nil, nil, fmt.Errorf("start tray controller: %w", err)
	}

	events := make(chan string, 4)
	go func() {
		defer close(events)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			events <- strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			log.Errorf("read tray controller output: %v", err)
		}
	}()
	return command, events, command.Wait, nil
}

func encodePowerShell(script string) string {
	encodedText := utf16.Encode([]rune(script))
	encodedBytes := make([]byte, len(encodedText)*2)
	for i, value := range encodedText {
		binary.LittleEndian.PutUint16(encodedBytes[i*2:], value)
	}
	return base64.StdEncoding.EncodeToString(encodedBytes)
}

func waitForTrayServer(cfg *config.Config, done <-chan struct{}) error {
	if cfg.Port <= 0 {
		return fmt.Errorf("invalid server port %d", cfg.Port)
	}
	address := net.JoinHostPort(trayBindHost(cfg.Host), strconv.Itoa(cfg.Port))
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()
	for {
		if conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond); err == nil {
			_ = conn.Close()
			select {
			case <-done:
				return fmt.Errorf("proxy service exited before becoming ready")
			default:
				return nil
			}
		}
		select {
		case <-done:
			return fmt.Errorf("proxy service exited before becoming ready")
		case <-timeout.C:
			return fmt.Errorf("proxy service did not listen on %s within 30 seconds", address)
		case <-ticker.C:
		}
	}
}

func trayBindHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" || host == "0.0.0.0" || host == "::" || host == "[::]" {
		return "127.0.0.1"
	}
	return strings.Trim(host, "[]")
}

func trayManagementURL(cfg *config.Config) string {
	scheme := "http"
	if cfg.TLS.Enable {
		scheme = "https"
	}
	return scheme + "://" + net.JoinHostPort(trayBindHost(cfg.Host), strconv.Itoa(cfg.Port)) + "/management.html"
}
