package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/tailnet"
	log "github.com/sirupsen/logrus"
)

// GetTailscale returns the Tailscale configuration and live node status.
func (h *Handler) GetTailscale(c *gin.Context) {
	h.mu.Lock()
	tsCfg := h.cfg.Tailscale
	h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"config": tsCfg,
		"status": tailnet.Default().Status(c.Request.Context(), tsCfg.Enable),
	})
}

// PutTailscale updates the Tailscale configuration, applies it and persists it.
// Omitted fields keep their current values; an empty auth-key keeps the stored key.
func (h *Handler) PutTailscale(c *gin.Context) {
	var body struct {
		Enable    *bool   `json:"enable"`
		Hostname  *string `json:"hostname"`
		AuthKey   *string `json:"auth-key"`
		StateDir  *string `json:"state-dir"`
		Ephemeral *bool   `json:"ephemeral"`
		HTTPS     *bool   `json:"https"`
		Funnel    *bool   `json:"funnel"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	next := h.cfg.Tailscale
	if body.Enable != nil {
		next.Enable = *body.Enable
	}
	if body.Hostname != nil {
		next.Hostname = strings.TrimSpace(*body.Hostname)
	}
	if body.AuthKey != nil && strings.TrimSpace(*body.AuthKey) != "" {
		next.AuthKey = strings.TrimSpace(*body.AuthKey)
	}
	if body.StateDir != nil {
		next.StateDir = strings.TrimSpace(*body.StateDir)
	}
	if body.Ephemeral != nil {
		next.Ephemeral = *body.Ephemeral
	}
	if body.HTTPS != nil {
		next.HTTPS = *body.HTTPS
	}
	if body.Funnel != nil {
		next.Funnel = *body.Funnel
	}
	h.applyTailscaleLocked(c, next)
}

// StartTailscale enables and starts the embedded Tailscale node.
func (h *Handler) StartTailscale(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	next := h.cfg.Tailscale
	next.Enable = true
	h.applyTailscaleLocked(c, next)
}

// StopTailscale disables and stops the embedded Tailscale node.
func (h *Handler) StopTailscale(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	next := h.cfg.Tailscale
	next.Enable = false
	h.applyTailscaleLocked(c, next)
}

// applyTailscaleLocked applies the new config to the runtime and persists it. Caller holds h.mu.
func (h *Handler) applyTailscaleLocked(c *gin.Context, next config.TailscaleConfig) {
	if errApply := tailnet.Default().ApplyConfig(c.Request.Context(), next); errApply != nil {
		log.WithError(errApply).Warn("failed to apply tailscale configuration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": errApply.Error()})
		return
	}
	h.cfg.Tailscale = next
	h.persistLocked(c)
}
