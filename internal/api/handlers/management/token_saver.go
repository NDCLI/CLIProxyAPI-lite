package management

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/tokensaver"
)

func tokenSaverView(value tokensaver.Snapshot) gin.H {
	return gin.H{
		"rtk_enabled":         value.Config.RTKEnabled,
		"headroom_enabled":    value.Config.HeadroomEnabled,
		"headroom_url":        value.Config.HeadroomURL,
		"headroom_timeout_ms": value.Config.HeadroomTimeoutMS,
		"headroom_status":     value.HeadroomStatus,
		"statistics":          value.Statistics,
	}
}

func (h *Handler) tokenSaverSnapshotLocked() tokensaver.Snapshot {
	if h.tokenSaver != nil {
		return h.tokenSaver.Snapshot()
	}
	return tokensaver.NewService(h.cfg.TokenSaver).Snapshot()
}

func (h *Handler) GetTokenSaver(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cfg == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": tokenSaverView(h.tokenSaverSnapshotLocked())})
}

func (h *Handler) PatchTokenSaver(c *gin.Context) {
	var body struct {
		RTKEnabled        *bool   `json:"rtk_enabled"`
		HeadroomEnabled   *bool   `json:"headroom_enabled"`
		HeadroomURL       *string `json:"headroom_url"`
		HeadroomTimeoutMS *int    `json:"headroom_timeout_ms"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid Token Saver settings"}})
		return
	}
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	updated := h.cfg.TokenSaver
	if body.RTKEnabled != nil {
		updated.RTKEnabled = *body.RTKEnabled
	}
	if body.HeadroomEnabled != nil {
		updated.HeadroomEnabled = *body.HeadroomEnabled
	}
	if body.HeadroomURL != nil {
		updated.HeadroomURL = *body.HeadroomURL
	}
	if body.HeadroomTimeoutMS != nil {
		updated.HeadroomTimeoutMS = *body.HeadroomTimeoutMS
	}
	var errNormalize error
	updated, errNormalize = tokensaver.NormalizeConfig(updated)
	if errNormalize != nil {
		h.mu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_token_saver_config", "message": errNormalize.Error()}})
		return
	}
	previous := h.cfg.TokenSaver
	h.cfg.TokenSaver = updated
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	if !ok {
		h.cfg.TokenSaver = previous
		h.mu.Unlock()
		return
	}
	service := h.tokenSaver
	if service != nil {
		service.Configure(updated)
	}
	h.mu.Unlock()
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	view := tokensaver.NewService(updated).Snapshot()
	if service != nil {
		view = service.Snapshot()
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": tokenSaverView(view)})
}

func (h *Handler) TestHeadroom(c *gin.Context) {
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	service := h.tokenSaver
	settings := h.cfg.TokenSaver
	h.mu.Unlock()
	if service == nil {
		service = tokensaver.NewService(settings)
	}
	result := service.CheckHeadroom(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": result})
}
