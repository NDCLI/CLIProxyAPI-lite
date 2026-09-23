package management

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type endpointKeyItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Mask     string `json:"mask"`
	Enabled  bool   `json:"enabled"`
	Success  int64  `json:"success"`
	Failed   int64  `json:"failed"`
	Revision string `json:"revision"`
}

func endpointKeyID(key string) string {
	digest := sha256.Sum256([]byte(key))
	return "key_" + hex.EncodeToString(digest[:8])
}

func maskEndpointKey(key string) string {
	key = strings.TrimSpace(key)
	if len(key) <= 4 {
		return "••••"
	}
	return "••••" + key[len(key)-4:]
}

func generateEndpointKey() (string, error) {
	raw := make([]byte, 24)
	if _, errRead := rand.Read(raw); errRead != nil {
		return "", errRead
	}
	return "sk-" + base64.RawURLEncoding.EncodeToString(raw), nil
}

func (h *Handler) endpointKeyItem(key string, index int) endpointKeyItem {
	item := endpointKeyItem{
		ID:       endpointKeyID(key),
		Label:    "Key " + strconv.Itoa(index+1),
		Mask:     maskEndpointKey(key),
		Enabled:  true,
		Revision: endpointKeyID(key),
	}
	h.mu.Lock()
	manager := h.authManager
	h.mu.Unlock()
	if manager == nil {
		return item
	}
	for _, auth := range manager.List() {
		kind, value := auth.AccountInfo()
		if strings.EqualFold(kind, "api_key") && strings.TrimSpace(value) == key {
			item.Success += auth.Success
			item.Failed += auth.Failed
		}
	}
	return item
}

func (h *Handler) GetEndpointKeys(c *gin.Context) {
	h.mu.Lock()
	keys := append([]string(nil), h.cfg.APIKeys...)
	h.mu.Unlock()
	items := make([]endpointKeyItem, 0, len(keys))
	for index, key := range keys {
		items = append(items, h.endpointKeyItem(key, index))
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items, "next_cursor": nil})
}

func (h *Handler) GetEndpointKeySecret(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	h.mu.Lock()
	defer h.mu.Unlock()
	c.Header("Cache-Control", "no-store")
	for _, key := range h.cfg.APIKeys {
		if endpointKeyID(key) == id {
			c.JSON(http.StatusOK, gin.H{"secret": key})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "endpoint_key_not_found", "message": "Endpoint key not found"}})
}

func (h *Handler) PostEndpointKey(c *gin.Context) {
	var body struct {
		Value string `json:"value"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid request body"}})
		return
	}
	key := strings.TrimSpace(body.Value)
	generated := key == ""
	if generated {
		var errGenerate error
		key, errGenerate = generateEndpointKey()
		if errGenerate != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "key_generation_failed", "message": "Failed to generate endpoint key"}})
			return
		}
	}

	h.mu.Lock()
	for _, existing := range h.cfg.APIKeys {
		if strings.TrimSpace(existing) == key {
			h.mu.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "endpoint_key_exists", "message": "Endpoint key already exists"}})
			return
		}
	}
	h.cfg.APIKeys = append(h.cfg.APIKeys, key)
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	index := len(h.cfg.APIKeys) - 1
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	response := gin.H{"schema_version": 1, "item": h.endpointKeyItem(key, index)}
	if generated {
		response["secret"] = key
	}
	c.JSON(http.StatusCreated, response)
}

func (h *Handler) PatchEndpointKey(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	var body struct {
		Value    string `json:"value"`
		Revision string `json:"revision"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid request body"}})
		return
	}
	key := strings.TrimSpace(body.Value)
	generated := key == ""
	if generated {
		var errGenerate error
		key, errGenerate = generateEndpointKey()
		if errGenerate != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "key_generation_failed", "message": "Failed to generate endpoint key"}})
			return
		}
	}

	h.mu.Lock()
	index := -1
	for candidate, existing := range h.cfg.APIKeys {
		if endpointKeyID(existing) == id {
			index = candidate
			break
		}
	}
	if index < 0 {
		h.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "endpoint_key_not_found", "message": "Endpoint key not found"}})
		return
	}
	if strings.TrimSpace(body.Revision) != endpointKeyID(h.cfg.APIKeys[index]) {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "stale_revision", "message": "Endpoint key changed; refresh and retry"}})
		return
	}
	for candidate, existing := range h.cfg.APIKeys {
		if candidate != index && strings.TrimSpace(existing) == key {
			h.mu.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "endpoint_key_exists", "message": "Endpoint key already exists"}})
			return
		}
	}
	h.cfg.APIKeys[index] = key
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	response := gin.H{"schema_version": 1, "item": h.endpointKeyItem(key, index)}
	if generated {
		response["secret"] = key
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteEndpointKey(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	revision := strings.TrimSpace(c.Query("revision"))
	h.mu.Lock()
	index := -1
	for candidate, key := range h.cfg.APIKeys {
		if endpointKeyID(key) == id {
			index = candidate
			break
		}
	}
	if index < 0 {
		h.mu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "endpoint_key_not_found", "message": "Endpoint key not found"}})
		return
	}
	if revision != endpointKeyID(h.cfg.APIKeys[index]) {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "stale_revision", "message": "Endpoint key changed; refresh and retry"}})
		return
	}
	h.cfg.APIKeys = append(h.cfg.APIKeys[:index], h.cfg.APIKeys[index+1:]...)
	snapshot, ok := h.saveConfigAndSnapshotLocked(c)
	h.mu.Unlock()
	if !ok {
		return
	}
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}
