package management

import (
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

type providerItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Provider string `json:"provider"`
	AuthType string `json:"auth_type,omitempty"`
	Enabled  bool   `json:"enabled"`
	Status   string `json:"status"`
	Success  int64  `json:"success"`
	Failed   int64  `json:"failed"`
}

type providerModelItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
	Type        string `json:"type,omitempty"`
	OwnedBy     string `json:"owned_by,omitempty"`
}

func providerDisplayLabel(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	if label := strings.TrimSpace(auth.Label); label != "" {
		return label
	}
	if name := strings.TrimSpace(auth.FileName); name != "" {
		return filepath.Base(name)
	}
	return strings.TrimSpace(auth.Provider)
}

func providerItemFromAuth(auth *coreauth.Auth) providerItem {
	enabled := auth != nil && !auth.Disabled && auth.Status != coreauth.StatusDisabled
	item := providerItem{
		ID:       strings.TrimSpace(auth.ID),
		Label:    providerDisplayLabel(auth),
		Provider: strings.TrimSpace(auth.Provider),
		AuthType: auth.AuthKind(),
		Enabled:  enabled,
		Status:   string(auth.Status),
		Success:  auth.Success,
		Failed:   auth.Failed,
	}
	if item.Status == "" {
		item.Status = string(coreauth.StatusUnknown)
	}
	return item
}

func (h *Handler) providerAuths() []*coreauth.Auth {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	manager := h.authManager
	h.mu.Unlock()
	if manager == nil {
		return nil
	}
	return manager.List()
}

func (h *Handler) findProvider(id string) (*coreauth.Auth, bool) {
	for _, auth := range h.providerAuths() {
		if auth != nil && strings.TrimSpace(auth.ID) == id {
			return auth, true
		}
	}
	return nil, false
}

// GetProviders returns normalized, secret-free provider credential metadata.
func (h *Handler) GetProviders(c *gin.Context) {
	items := make([]providerItem, 0)
	for _, auth := range h.providerAuths() {
		if auth == nil || strings.TrimSpace(auth.ID) == "" {
			continue
		}
		items = append(items, providerItemFromAuth(auth))
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Provider != items[j].Provider {
			return items[i].Provider < items[j].Provider
		}
		return items[i].ID < items[j].ID
	})
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items, "next_cursor": nil})
}

// GetProvider returns one normalized provider credential by its stable ID.
func (h *Handler) GetProvider(c *gin.Context) {
	auth, ok := h.findProvider(strings.TrimSpace(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "provider_not_found", "message": "Provider not found"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": providerItemFromAuth(auth)})
}

// GetProviderModels returns models currently registered for one provider credential.
func (h *Handler) GetProviderModels(c *gin.Context) {
	auth, ok := h.findProvider(strings.TrimSpace(c.Param("id")))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "provider_not_found", "message": "Provider not found"}})
		return
	}
	models := registry.GetGlobalRegistry().GetModelsForClient(auth.ID)
	items := make([]providerModelItem, 0, len(models))
	for _, model := range models {
		if model == nil || strings.TrimSpace(model.ID) == "" {
			continue
		}
		items = append(items, providerModelItem{ID: model.ID, DisplayName: model.DisplayName, Type: model.Type, OwnedBy: model.OwnedBy})
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items, "next_cursor": nil})
}
