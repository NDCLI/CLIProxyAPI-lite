package management

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	coreexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
)

func (h *Handler) TestProviderModel(c *gin.Context) {
	h.mu.Lock()
	manager := h.authManager
	h.mu.Unlock()
	if manager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "provider_unavailable", "message": "Provider execution is unavailable"}})
		return
	}
	auth, ok := h.findProvider(strings.TrimSpace(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "provider_not_found", "message": "Provider not found"}})
		return
	}
	if auth.Disabled || auth.Status == coreauth.StatusDisabled {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "provider_disabled", "message": "Enable this provider before testing a model"}})
		return
	}
	var body struct {
		Model string `json:"model"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil || strings.TrimSpace(body.Model) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "model_required", "message": "A model is required"}})
		return
	}
	model := strings.TrimSpace(body.Model)
	registered := false
	for _, info := range registry.GetGlobalRegistry().GetModelsForClient(auth.ID) {
		if info != nil && info.ID == model {
			registered = true
			break
		}
	}
	if !registered {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "model_not_registered", "message": "Choose a model listed for this provider"}})
		return
	}
	payload, errMarshal := json.Marshal(map[string]any{
		"model":      model,
		"messages":   []map[string]string{{"role": "user", "content": "Reply with OK."}},
		"max_tokens": 32,
	})
	if errMarshal != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "model_test_failed", "message": "Could not prepare the test request"}})
		return
	}
	start := time.Now()
	_, errExecute := manager.Execute(c.Request.Context(), []string{auth.Provider}, coreexecutor.Request{
		Model:   model,
		Payload: payload,
		Format:  sdktranslator.FormatOpenAI,
	}, coreexecutor.Options{Metadata: map[string]any{
		coreexecutor.PinnedAuthMetadataKey:     auth.ID,
		coreexecutor.RequestedModelMetadataKey: model,
		coreexecutor.GenerateMetadataKey:       false,
	}})
	if errExecute != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "model": model, "latency_ms": time.Since(start).Milliseconds(), "error": "The provider could not complete this model test"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "model": model, "latency_ms": time.Since(start).Milliseconds()})
}
