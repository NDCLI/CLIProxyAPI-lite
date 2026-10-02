package management

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

type compatibleProviderItem struct {
	ID         string                            `json:"id"`
	Name       string                            `json:"name"`
	BaseURL    string                            `json:"base_url"`
	Prefix     string                            `json:"prefix,omitempty"`
	Enabled    bool                              `json:"enabled"`
	APIKeyMask []string                          `json:"api_key_masks"`
	Models     []config.OpenAICompatibilityModel `json:"models"`
}

type compatibleProviderMutation struct {
	Name         string                             `json:"name"`
	BaseURL      string                             `json:"base_url"`
	Prefix       *string                            `json:"prefix"`
	APIKey       string                             `json:"api_key"`
	APIKeyAction string                             `json:"api_key_action"`
	APIKeyIndex  *int                               `json:"api_key_index"`
	Disabled     *bool                              `json:"disabled"`
	Models       *[]config.OpenAICompatibilityModel `json:"models"`
}

func mutateCompatibleProviderAPIKey(entry *config.OpenAICompatibility, action string, index *int, key string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	key = strings.TrimSpace(key)
	if action == "" && key != "" && index == nil {
		action = "append"
	}
	switch action {
	case "":
		if index != nil || key != "" {
			return errors.New("invalid API key operation")
		}
		return nil
	case "append":
		if index != nil || key == "" {
			return errors.New("invalid API key operation")
		}
		entry.APIKeyEntries = append(entry.APIKeyEntries, config.OpenAICompatibilityAPIKey{APIKey: key})
	case "replace":
		if index == nil || *index < 0 || *index >= len(entry.APIKeyEntries) || key == "" {
			return errors.New("invalid API key operation")
		}
		entry.APIKeyEntries[*index].APIKey = key
	case "delete":
		if index == nil || *index < 0 || *index >= len(entry.APIKeyEntries) || key != "" {
			return errors.New("invalid API key operation")
		}
		entry.APIKeyEntries = append(entry.APIKeyEntries[:*index], entry.APIKeyEntries[*index+1:]...)
	default:
		return errors.New("invalid API key operation")
	}
	return nil
}

func compatibleProviderID(name string) string {
	return endpointKeyID("compatible-provider:" + strings.ToLower(strings.TrimSpace(name)))
}

func safeCompatibleProviderURL(value string) string {
	parsed, errParse := url.Parse(value)
	if errParse != nil {
		return "configured URL"
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func compatibleProviderFromConfig(item config.OpenAICompatibility) compatibleProviderItem {
	keys := make([]string, 0, len(item.APIKeyEntries))
	for _, entry := range item.APIKeyEntries {
		keys = append(keys, maskEndpointKey(entry.APIKey))
	}
	models := append([]config.OpenAICompatibilityModel(nil), item.Models...)
	return compatibleProviderItem{
		ID: compatibleProviderID(item.Name), Name: item.Name, BaseURL: safeCompatibleProviderURL(item.BaseURL),
		Prefix: item.Prefix, Enabled: !item.Disabled, APIKeyMask: keys, Models: models,
	}
}

func validCompatibleProviderURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != "" && parsed.User == nil
}

func validateCompatibleProviderModels(c *gin.Context, models []config.OpenAICompatibilityModel) bool {
	aliases := make(map[string]struct{}, len(models))
	for index := range models {
		models[index].Name = strings.TrimSpace(models[index].Name)
		models[index].Alias = strings.TrimSpace(models[index].Alias)
		if models[index].Name == "" || models[index].Alias == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_provider_model", "message": "Each model requires a name and alias"}})
			return false
		}
		key := strings.ToLower(models[index].Alias)
		if _, exists := aliases[key]; exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "duplicate_provider_model", "message": "Model aliases must be unique"}})
			return false
		}
		aliases[key] = struct{}{}
	}
	return true
}

func (h *Handler) GetCompatibleProviders(c *gin.Context) {
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	items := make([]compatibleProviderItem, 0, len(h.cfg.OpenAICompatibility))
	for _, item := range h.cfg.OpenAICompatibility {
		items = append(items, compatibleProviderFromConfig(item))
	}
	h.mu.Unlock()
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items, "next_cursor": nil})
}

func (h *Handler) DiscoverCompatibleProviderModels(c *gin.Context) {
	var body struct {
		ID      string `json:"id"`
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil || !validCompatibleProviderURL(body.BaseURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid provider URL required"})
		return
	}
	baseURL := strings.TrimRight(strings.TrimSpace(body.BaseURL), "/")
	apiKey := strings.TrimSpace(body.APIKey)
	if apiKey == "" && body.ID != "" {
		h.mu.Lock()
		if h.cfg != nil {
			for _, entry := range h.cfg.OpenAICompatibility {
				if compatibleProviderID(entry.Name) == body.ID && strings.TrimRight(entry.BaseURL, "/") == baseURL && len(entry.APIKeyEntries) > 0 {
					apiKey = entry.APIKeyEntries[0].APIKey
					break
				}
			}
		}
		h.mu.Unlock()
	}
	models, errModels := fetchCompatibleProviderModels(c.Request.Context(), baseURL, apiKey)
	if errModels != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": errModels.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": models})
}

func (h *Handler) TestCompatibleProviderAPIKey(c *gin.Context) {
	var body struct {
		ID          string `json:"id"`
		BaseURL     string `json:"base_url"`
		APIKey      string `json:"api_key"`
		APIKeyIndex *int   `json:"api_key_index"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil || !validCompatibleProviderURL(body.BaseURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid provider URL required"})
		return
	}
	baseURL := strings.TrimRight(strings.TrimSpace(body.BaseURL), "/")
	apiKey := strings.TrimSpace(body.APIKey)
	if body.APIKeyIndex != nil {
		if body.ID == "" || apiKey != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "select a saved API key or provide a new API key"})
			return
		}
		h.mu.Lock()
		if h.cfg != nil {
			for _, entry := range h.cfg.OpenAICompatibility {
				if compatibleProviderID(entry.Name) != body.ID {
					continue
				}
				if strings.TrimRight(strings.TrimSpace(entry.BaseURL), "/") != baseURL {
					h.mu.Unlock()
					c.JSON(http.StatusConflict, gin.H{"error": "save the provider URL before testing this key"})
					return
				}
				if *body.APIKeyIndex < 0 || *body.APIKeyIndex >= len(entry.APIKeyEntries) {
					h.mu.Unlock()
					c.JSON(http.StatusNotFound, gin.H{"error": "saved API key not found"})
					return
				}
				apiKey = strings.TrimSpace(entry.APIKeyEntries[*body.APIKeyIndex].APIKey)
				break
			}
		}
		h.mu.Unlock()
		if apiKey == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider or saved API key not found"})
			return
		}
	}
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key required"})
		return
	}
	start := time.Now()
	models, errModels := fetchCompatibleProviderModels(c.Request.Context(), baseURL, apiKey)
	if errModels != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "latency_ms": time.Since(start).Milliseconds(), "error": "API key or provider connection failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "latency_ms": time.Since(start).Milliseconds(), "model_count": len(models)})
}

func fetchCompatibleProviderModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	request, errRequest := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/models", nil)
	if errRequest != nil {
		return nil, errors.New("invalid provider URL")
	}
	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, errDo := client.Do(request)
	if errDo != nil {
		return nil, errors.New("provider model request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned HTTP %s", http.StatusText(response.StatusCode))
	}
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if errDecode := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&result); errDecode != nil {
		return nil, errors.New("provider returned an invalid model list")
	}
	models := make([]string, 0, len(result.Data))
	for _, model := range result.Data {
		if id := strings.TrimSpace(model.ID); id != "" && !strings.ContainsAny(id, "\r\n\x00|") {
			models = append(models, id)
		}
	}
	return models, nil
}

func (h *Handler) PostCompatibleProvider(c *gin.Context) {
	var body compatibleProviderMutation
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid provider configuration"}})
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.BaseURL = strings.TrimSpace(body.BaseURL)
	if body.Name == "" || !validCompatibleProviderURL(body.BaseURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_provider", "message": "A name and valid HTTP(S) base URL are required"}})
		return
	}
	models := []config.OpenAICompatibilityModel{}
	if body.Models != nil {
		models = append(models, (*body.Models)...)
	}
	if !validateCompatibleProviderModels(c, models) {
		return
	}
	prefix := ""
	if body.Prefix != nil {
		prefix = strings.TrimSpace(*body.Prefix)
	}
	entry := config.OpenAICompatibility{Name: body.Name, BaseURL: body.BaseURL, Prefix: prefix, Models: models}
	if strings.TrimSpace(body.APIKey) != "" {
		entry.APIKeyEntries = []config.OpenAICompatibilityAPIKey{{APIKey: strings.TrimSpace(body.APIKey)}}
	}
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	for _, existing := range h.cfg.OpenAICompatibility {
		if strings.EqualFold(strings.TrimSpace(existing.Name), entry.Name) {
			h.mu.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "provider_exists", "message": "A provider with this name already exists"}})
			return
		}
	}
	h.cfg.OpenAICompatibility = append(h.cfg.OpenAICompatibility, entry)
	h.cfg.SanitizeOpenAICompatibility()
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.JSON(http.StatusCreated, gin.H{"schema_version": 1, "item": compatibleProviderFromConfig(entry)})
}

func (h *Handler) PatchCompatibleProvider(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var body compatibleProviderMutation
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid provider configuration"}})
		return
	}
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	index := -1
	for candidate, item := range h.cfg.OpenAICompatibility {
		if compatibleProviderID(item.Name) == id {
			index = candidate
			break
		}
	}
	if index < 0 {
		h.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "provider_not_found", "message": "Provider not found"}})
		return
	}
	entry := h.cfg.OpenAICompatibility[index]
	entry.APIKeyEntries = append([]config.OpenAICompatibilityAPIKey(nil), entry.APIKeyEntries...)
	if body.BaseURL != "" {
		if !validCompatibleProviderURL(body.BaseURL) {
			h.mu.Unlock()
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_provider_url", "message": "A valid HTTP(S) base URL is required"}})
			return
		}
		entry.BaseURL = strings.TrimSpace(body.BaseURL)
	}
	if body.Prefix != nil {
		entry.Prefix = strings.TrimSpace(*body.Prefix)
	}
	if body.Disabled != nil {
		entry.Disabled = *body.Disabled
	}
	if body.Models != nil {
		models := append([]config.OpenAICompatibilityModel(nil), (*body.Models)...)
		if !validateCompatibleProviderModels(c, models) {
			h.mu.Unlock()
			return
		}
		entry.Models = models
	}
	if err := mutateCompatibleProviderAPIKey(&entry, body.APIKeyAction, body.APIKeyIndex, body.APIKey); err != nil {
		h.mu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_api_key_operation", "message": "Select a saved API key to replace or delete, or choose to add a new key"}})
		return
	}
	h.cfg.OpenAICompatibility[index] = entry
	h.cfg.SanitizeOpenAICompatibility()
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": compatibleProviderFromConfig(entry)})
}

func (h *Handler) DeleteCompatibleProvider(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	h.mu.Lock()
	if h.cfg == nil {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "config_unavailable", "message": "Configuration unavailable"}})
		return
	}
	index := -1
	for candidate, item := range h.cfg.OpenAICompatibility {
		if compatibleProviderID(item.Name) == id {
			index = candidate
			break
		}
	}
	if index < 0 {
		h.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "provider_not_found", "message": "Provider not found"}})
		return
	}
	h.cfg.OpenAICompatibility = append(h.cfg.OpenAICompatibility[:index], h.cfg.OpenAICompatibility[index+1:]...)
	h.cfg.SanitizeOpenAICompatibility()
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.Status(http.StatusNoContent)
}
