package management

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) PatchManagementPassword(c *gin.Context) {
	if !requireLocalManagement(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 4096)
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid password request"}})
		return
	}
	if len(body.NewPassword) < 6 || len(body.NewPassword) > 128 || strings.TrimSpace(body.NewPassword) != body.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_password", "message": "Password must be 6-128 characters without leading or trailing spaces"}})
		return
	}

	h.mu.Lock()
	if h.cfg == nil || h.configFilePath == "" || h.cfg.RemoteManagement.SecretKey == "" {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "password_unavailable", "message": "Configured management password is unavailable"}})
		return
	}
	if h.envSecret != "" || h.localPassword != "" {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "password_override", "message": "Remove the runtime management password override before changing the configured password"}})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(h.cfg.RemoteManagement.SecretKey), []byte(body.CurrentPassword)) != nil {
		h.mu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "wrong_current_password", "message": "Current password is incorrect"}})
		return
	}
	if body.CurrentPassword == body.NewPassword {
		h.mu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "password_unchanged", "message": "New password must differ from the current password"}})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.mu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	if err = config.SaveConfigPreserveCommentsUpdateNestedScalar(h.configFilePath, []string{"remote-management", "secret-key"}, string(hash)); err != nil {
		h.mu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save password"})
		return
	}
	h.cfg.RemoteManagement.SecretKey = string(hash)
	snapshot := h.reloadSnapshotConfigLocked()
	h.mu.Unlock()
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
