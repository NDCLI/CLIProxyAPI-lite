package management

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/proxyutil"
)

type proxyPoolRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	ProxyURL  string    `json:"proxy_url"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type proxyPoolStore struct {
	path string
	mu   sync.Mutex
}

func newProxyPoolStore(configPath string) *proxyPoolStore {
	return &proxyPoolStore{path: filepath.Join(filepath.Dir(configPath), "proxy-pools.json")}
}

func (s *proxyPoolStore) list() ([]proxyPoolRecord, error) {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return []proxyPoolRecord{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *proxyPoolStore) loadLocked() ([]proxyPoolRecord, error) {
	body, errRead := os.ReadFile(s.path)
	if errors.Is(errRead, os.ErrNotExist) {
		return []proxyPoolRecord{}, nil
	}
	if errRead != nil {
		return nil, errRead
	}
	var items []proxyPoolRecord
	if errDecode := json.Unmarshal(body, &items); errDecode != nil {
		return nil, fmt.Errorf("decode proxy pools: %w", errDecode)
	}
	return items, nil
}

func (s *proxyPoolStore) save(item proxyPoolRecord) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return errors.New("proxy pool store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items, errLoad := s.loadLocked()
	if errLoad != nil {
		return errLoad
	}
	for index := range items {
		if items[index].ID == item.ID {
			items[index] = item
			return s.persistLocked(items)
		}
	}
	items = append(items, item)
	return s.persistLocked(items)
}

func (s *proxyPoolStore) delete(id string) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return errors.New("proxy pool store unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items, errLoad := s.loadLocked()
	if errLoad != nil {
		return errLoad
	}
	for index := range items {
		if items[index].ID == id {
			items = append(items[:index], items[index+1:]...)
			return s.persistLocked(items)
		}
	}
	return os.ErrNotExist
}

func (s *proxyPoolStore) persistLocked(items []proxyPoolRecord) error {
	if errDir := os.MkdirAll(filepath.Dir(s.path), 0o700); errDir != nil {
		return errDir
	}
	body, errMarshal := json.MarshalIndent(items, "", "  ")
	if errMarshal != nil {
		return errMarshal
	}
	return os.WriteFile(s.path, body, 0o600)
}

type proxyPoolItem struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	ProxyURLMasked     string    `json:"proxy_url_masked"`
	ProxyURLConfigured bool      `json:"proxy_url_configured"`
	IsActive           bool      `json:"is_active"`
	BoundCredentials   int       `json:"bound_credentials"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func maskProxyURL(raw string) string {
	parsed, errParse := url.Parse(strings.TrimSpace(raw))
	if errParse != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func proxyPoolIDForAuth(auth *coreauth.Auth) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	id, _ := auth.Metadata["proxy_pool_id"].(string)
	return strings.TrimSpace(id)
}

func (h *Handler) proxyPoolItems() ([]proxyPoolItem, error) {
	items, errList := h.proxyPools.list()
	if errList != nil {
		return nil, errList
	}
	counts := make(map[string]int)
	for _, auth := range h.providerAuths() {
		if id := proxyPoolIDForAuth(auth); id != "" {
			counts[id]++
		}
	}
	result := make([]proxyPoolItem, 0, len(items))
	for _, item := range items {
		result = append(result, proxyPoolItem{
			ID: item.ID, Name: item.Name, ProxyURLMasked: maskProxyURL(item.ProxyURL), ProxyURLConfigured: item.ProxyURL != "",
			IsActive: item.IsActive, BoundCredentials: counts[item.ID], CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		})
	}
	return result, nil
}

func (h *Handler) GetProxyPools(c *gin.Context) {
	items, errList := h.proxyPoolItems()
	if errList != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_read_failed", "message": "Could not read proxy pools"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items})
}

func validateProxyPoolURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	setting, errParse := proxyutil.Parse(raw)
	if errParse != nil || setting.Mode != proxyutil.ModeProxy {
		return "", errors.New("proxy_url must be a valid HTTP, HTTPS, or SOCKS5 proxy URL")
	}
	return raw, nil
}

func newProxyPoolID() (string, error) {
	var value [16]byte
	if _, errRead := rand.Read(value[:]); errRead != nil {
		return "", errRead
	}
	return hex.EncodeToString(value[:]), nil
}

func (h *Handler) PostProxyPool(c *gin.Context) {
	var body struct {
		Name     string `json:"name"`
		ProxyURL string `json:"proxy_url"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid proxy pool"}})
		return
	}
	name := strings.TrimSpace(body.Name)
	proxyURL, errProxy := validateProxyPoolURL(body.ProxyURL)
	if name == "" || errProxy != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_proxy_pool", "message": "Name and a valid proxy URL are required"}})
		return
	}
	id, errID := newProxyPoolID()
	if errID != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_create_failed", "message": "Could not create proxy pool"}})
		return
	}
	now := time.Now().UTC()
	item := proxyPoolRecord{ID: id, Name: name, ProxyURL: proxyURL, IsActive: true, CreatedAt: now, UpdatedAt: now}
	if errSave := h.proxyPools.save(item); errSave != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_create_failed", "message": "Could not save proxy pool"}})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"schema_version": 1, "item": proxyPoolItem{ID: item.ID, Name: item.Name, ProxyURLMasked: maskProxyURL(item.ProxyURL), ProxyURLConfigured: true, IsActive: true, CreatedAt: now, UpdatedAt: now}})
}

func (h *Handler) PatchProxyPool(c *gin.Context) {
	var body struct {
		Name     *string `json:"name"`
		ProxyURL *string `json:"proxy_url"`
		IsActive *bool   `json:"is_active"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_body", "message": "Invalid proxy pool update"}})
		return
	}
	items, errList := h.proxyPools.list()
	if errList != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_read_failed", "message": "Could not read proxy pools"}})
		return
	}
	id := strings.TrimSpace(c.Param("id"))
	var item proxyPoolRecord
	found := false
	for _, existing := range items {
		if existing.ID == id {
			item, found = existing, true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "proxy_pool_not_found", "message": "Proxy pool not found"}})
		return
	}
	if body.Name != nil {
		item.Name = strings.TrimSpace(*body.Name)
		if item.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_proxy_pool", "message": "Name is required"}})
			return
		}
	}
	if body.ProxyURL != nil {
		proxyURL, errProxy := validateProxyPoolURL(*body.ProxyURL)
		if errProxy != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_proxy_pool", "message": "A valid proxy URL is required"}})
			return
		}
		item.ProxyURL = proxyURL
	}
	if body.IsActive != nil {
		item.IsActive = *body.IsActive
	}
	item.UpdatedAt = time.Now().UTC()
	if errSave := h.proxyPools.save(item); errSave != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_update_failed", "message": "Could not save proxy pool"}})
		return
	}
	if errSync := h.syncProxyPoolCredentials(c.Request.Context(), item); errSync != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_runtime_update_failed", "message": "Pool saved but credentials could not all be updated"}})
		return
	}
	updated, _ := h.proxyPoolItems()
	for _, publicItem := range updated {
		if publicItem.ID == item.ID {
			c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": publicItem})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_update_failed", "message": "Could not read updated proxy pool"}})
}

func (h *Handler) DeleteProxyPool(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	for _, auth := range h.providerAuths() {
		if proxyPoolIDForAuth(auth) == id {
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "proxy_pool_in_use", "message": "Unassign provider credentials before deleting this pool"}})
			return
		}
	}
	if errDelete := h.proxyPools.delete(id); errDelete != nil {
		status := http.StatusInternalServerError
		if errors.Is(errDelete, os.ErrNotExist) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": gin.H{"code": "proxy_pool_delete_failed", "message": "Could not delete proxy pool"}})
		return
	}
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

func proxyPoolOriginalURL(auth *coreauth.Auth) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	value, _ := auth.Metadata["proxy_pool_original_url"].(string)
	return strings.TrimSpace(value)
}

func (h *Handler) updateCredentialProxy(ctx context.Context, auth *coreauth.Auth, poolID, proxyURL, originalURL string) error {
	if auth == nil || coreauth.IsPluginVirtualAuth(auth) {
		return errors.New("credential does not support proxy pools")
	}
	updated := auth.Clone()
	if updated.Metadata == nil {
		updated.Metadata = make(map[string]any)
	}
	updated.ProxyURL = proxyURL
	updated.Metadata["proxy_url"] = proxyURL
	if poolID == "" {
		delete(updated.Metadata, "proxy_pool_id")
		delete(updated.Metadata, "proxy_pool_original_url")
	} else {
		updated.Metadata["proxy_pool_id"] = poolID
		updated.Metadata["proxy_pool_original_url"] = originalURL
	}
	updated.UpdatedAt = time.Now().UTC()
	stored, errUpdate := h.authManager.Update(ctx, updated)
	if errUpdate != nil {
		return errUpdate
	}
	if stored == nil {
		stored = updated
	}
	return h.invokePostAuthPersistHooks(ctx, []*coreauth.Auth{stored})
}

func (h *Handler) syncProxyPoolCredentials(ctx context.Context, pool proxyPoolRecord) error {
	h.authStatusMu.Lock()
	defer h.authStatusMu.Unlock()
	for _, auth := range h.providerAuths() {
		if proxyPoolIDForAuth(auth) != pool.ID {
			continue
		}
		targetURL := proxyPoolOriginalURL(auth)
		if pool.IsActive {
			targetURL = pool.ProxyURL
		}
		if auth.ProxyURL == targetURL {
			continue
		}
		if errUpdate := h.updateCredentialProxy(ctx, auth, pool.ID, targetURL, proxyPoolOriginalURL(auth)); errUpdate != nil {
			return errUpdate
		}
	}
	return nil
}

func (h *Handler) assignProxyPool(c *gin.Context, auth *coreauth.Auth, poolID string) {
	if h.authManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{"code": "auth_manager_unavailable", "message": "Credential manager unavailable"}})
		return
	}
	poolID = strings.TrimSpace(poolID)
	var pool proxyPoolRecord
	if poolID != "" {
		items, errList := h.proxyPools.list()
		if errList != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_read_failed", "message": "Could not read proxy pools"}})
			return
		}
		found := false
		for _, item := range items {
			if item.ID == poolID {
				pool, found = item, true
				break
			}
		}
		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "proxy_pool_not_found", "message": "Proxy pool not found"}})
			return
		}
	}
	if auth == nil || coreauth.IsPluginVirtualAuth(auth) {
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "proxy_pool_assignment_unsupported", "message": "This credential cannot use a proxy pool"}})
		return
	}
	h.authStatusMu.Lock()
	defer h.authStatusMu.Unlock()
	currentID := proxyPoolIDForAuth(auth)
	originalURL := auth.ProxyURL
	if currentID != "" {
		originalURL = proxyPoolOriginalURL(auth)
	}
	proxyURL := originalURL
	if poolID != "" && pool.IsActive {
		proxyURL = pool.ProxyURL
	}
	if errUpdate := h.updateCredentialProxy(c.Request.Context(), auth, poolID, proxyURL, originalURL); errUpdate != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "proxy_pool_assignment_failed", "message": "Could not update credential proxy"}})
		return
	}
	updated, _ := h.findProvider(auth.ID)
	if updated == nil {
		updated = auth
	}
	item := providerItemFromAuth(updated)
	item.ProxyPoolID = proxyPoolIDForAuth(updated)
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": item})
}
