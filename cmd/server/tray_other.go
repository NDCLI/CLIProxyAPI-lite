//go:build !windows

package main

import (
	"fmt"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/pluginhost"
)

func acquireTrayInstance() (func(), bool, error) {
	return func() {}, true, nil
}

func runTrayMode(_ *config.Config, _ string, _ string, _ *pluginhost.Host, _ ...api.ServerOption) error {
	return fmt.Errorf("--tray is supported only on Windows")
}
