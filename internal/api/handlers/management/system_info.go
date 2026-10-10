package management

import (
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/buildinfo"
)

type systemInfoItem struct {
	Version         string `json:"version"`
	Commit          string `json:"commit"`
	BuildDate       string `json:"build_date"`
	GoVersion       string `json:"go_version"`
	OperatingSystem string `json:"operating_system"`
	Architecture    string `json:"architecture"`
	UpdateCommand   string `json:"update_command"`
}

func (h *Handler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"schema_version": 1,
		"item": systemInfoItem{
			Version:         buildinfo.Version,
			Commit:          buildinfo.Commit,
			BuildDate:       buildinfo.BuildDate,
			GoVersion:       runtime.Version(),
			OperatingSystem: runtime.GOOS,
			Architecture:    runtime.GOARCH,
			UpdateCommand:   h.updateCommand(),
		},
	})
}

func (h *Handler) updateCommand() string {
	command := "npm install -g @ndcli/lumina@latest --include=optional"
	configPath := h.configFilePath
	if runtime.GOOS == "windows" {
		launch := "lumina --tray"
		if configPath != "" {
			launch += " --config '" + strings.ReplaceAll(configPath, "'", "''") + "'"
		}
		return command + "\nif ($LASTEXITCODE -eq 0) { " + launch + " }"
	}
	launch := "lumina"
	if configPath != "" {
		launch += " --config '" + strings.ReplaceAll(configPath, "'", "'\"'\"'") + "'"
	}
	return command + " && " + launch
}
