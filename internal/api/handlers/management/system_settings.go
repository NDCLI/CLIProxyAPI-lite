package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type systemSettings struct {
	Host                   string `json:"host"`
	Port                   int    `json:"port"`
	LoggingToFile          bool   `json:"logging_to_file"`
	UsageStatisticsEnabled bool   `json:"usage_statistics_enabled"`
	RoutingStrategy        string `json:"routing_strategy"`
	RequestRetry           int    `json:"request_retry"`
	MaxRetryCredentials    int    `json:"max_retry_credentials"`
	MaxRetryInterval       int    `json:"max_retry_interval"`
}

func (h *Handler) currentSystemSettingsLocked() systemSettings {
	return systemSettings{
		Host: h.cfg.Host, Port: h.cfg.Port, LoggingToFile: h.cfg.LoggingToFile, UsageStatisticsEnabled: h.cfg.UsageStatisticsEnabled,
		RoutingStrategy: h.cfg.Routing.Strategy, RequestRetry: h.cfg.RequestRetry, MaxRetryCredentials: h.cfg.MaxRetryCredentials, MaxRetryInterval: h.cfg.MaxRetryInterval,
	}
}

func (h *Handler) GetSystemSettings(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cfg == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": h.currentSystemSettingsLocked()})
}

func (h *Handler) PatchSystemSettings(c *gin.Context) {
	var body struct {
		LoggingToFile          *bool   `json:"logging_to_file"`
		UsageStatisticsEnabled *bool   `json:"usage_statistics_enabled"`
		RoutingStrategy        *string `json:"routing_strategy"`
		RequestRetry           *int    `json:"request_retry"`
		MaxRetryCredentials    *int    `json:"max_retry_credentials"`
		MaxRetryInterval       *int    `json:"max_retry_interval"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid request body"}})
		return
	}
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	if body.LoggingToFile != nil {
		h.cfg.LoggingToFile = *body.LoggingToFile
	}
	if body.UsageStatisticsEnabled != nil {
		h.cfg.UsageStatisticsEnabled = *body.UsageStatisticsEnabled
	}
	if body.RoutingStrategy != nil {
		strategy := strings.TrimSpace(*body.RoutingStrategy)
		if strategy == "" {
			h.mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_routing_strategy", "message": "Routing strategy is required"}})
			return
		}
		h.cfg.Routing.Strategy = strategy
	}
	if body.RequestRetry != nil {
		if *body.RequestRetry < 0 {
			h.mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_retry", "message": "Request retry must not be negative"}})
			return
		}
		h.cfg.RequestRetry = *body.RequestRetry
	}
	if body.MaxRetryCredentials != nil {
		if *body.MaxRetryCredentials < 0 {
			h.mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_retry", "message": "Max retry credentials must not be negative"}})
			return
		}
		h.cfg.MaxRetryCredentials = *body.MaxRetryCredentials
	}
	if body.MaxRetryInterval != nil {
		if *body.MaxRetryInterval < 0 {
			h.mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_retry", "message": "Max retry interval must not be negative"}})
			return
		}
		h.cfg.MaxRetryInterval = *body.MaxRetryInterval
	}
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	item := h.currentSystemSettingsLocked()
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": item})
}
