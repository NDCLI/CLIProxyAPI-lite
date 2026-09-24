package management

import (
	"net/http"
	"runtime"

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
		},
	})
}
