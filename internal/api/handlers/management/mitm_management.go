package management

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetMITMStatus(c *gin.Context) {
	certPath, _ := h.mitmCAPaths()
	c.JSON(http.StatusOK, h.mitm.status(certPath))
}

func (h *Handler) StartMITM(c *gin.Context) {
	var body struct {
		APIKey   string `json:"api_key"`
		APIKeyID string `json:"api_key_id"`
		BaseURL  string `json:"base_url"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(body.BaseURL) == "" {
		h.mitm.mu.RLock()
		body.BaseURL = h.mitm.gatewayBaseURLLocked()
		h.mitm.mu.RUnlock()
	}
	baseURL, errURL := normalizeCLIBaseURL(body.BaseURL)
	if errURL != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errURL.Error()})
		return
	}
	apiKey, errKey := h.resolveCLIEndpointKey(baseURL, body.APIKeyID, body.APIKey)
	if errKey != nil || apiKey == "" {
		message := "api_key or api_key_id is required"
		if errKey != nil {
			message = errKey.Error()
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}
	certPath, keyPath, errEnsure := h.ensureMITMCA()
	if errEnsure != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errEnsure.Error()})
		return
	}
	if errStart := h.mitm.start(certPath, keyPath, apiKey, strings.TrimSuffix(baseURL, "/v1"), strings.TrimSpace(body.APIKeyID)); errStart != nil {
		c.JSON(http.StatusConflict, gin.H{"error": errStart.Error()})
		return
	}
	if errTrust := setMITMNodeCA(certPath); errTrust != nil {
		_ = h.mitm.stop(c.Request.Context())
		c.JSON(http.StatusInternalServerError, gin.H{"error": errTrust.Error()})
		return
	}
	c.JSON(http.StatusOK, h.mitm.status(certPath))
}

func (h *Handler) StopMITM(c *gin.Context) {
	for tool, enabled := range mitmDNSStatus() {
		if enabled {
			if errDNS := setMITMDNS(tool, false); errDNS != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "stop aborted because DNS cleanup failed: " + errDNS.Error()})
				return
			}
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if errStop := h.mitm.stop(ctx); errStop != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errStop.Error()})
		return
	}
	certPath, _ := h.mitmCAPaths()
	if errTrust := clearMITMNodeCA(certPath); errTrust != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errTrust.Error()})
		return
	}
	c.JSON(http.StatusOK, h.mitm.status(certPath))
}

func (h *Handler) ToggleMITMDNS(c *gin.Context) {
	var body struct {
		Tool    string `json:"tool"`
		Enabled *bool  `json:"enabled"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil || body.Enabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool and enabled are required"})
		return
	}
	body.Tool = strings.ToLower(strings.TrimSpace(body.Tool))
	if *body.Enabled {
		certPath, _ := h.mitmCAPaths()
		status := h.mitm.status(certPath)
		if !status.Running {
			c.JSON(http.StatusConflict, gin.H{"error": "start the MITM server before enabling DNS"})
			return
		}
		if !status.CertTrusted {
			c.JSON(http.StatusConflict, gin.H{"error": "install the Root CA before enabling DNS"})
			return
		}
	}
	if errSet := setMITMDNS(body.Tool, *body.Enabled); errSet != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": errSet.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "dns": mitmDNSStatus()})
}

func (h *Handler) GetMITMMappings(c *gin.Context) {
	tool := strings.ToLower(strings.TrimSpace(c.Query("tool")))
	if tool != "" && len(mitmToolHosts[tool]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported MITM tool"})
		return
	}
	h.mitm.mu.RLock()
	defer h.mitm.mu.RUnlock()
	if tool != "" {
		c.JSON(http.StatusOK, gin.H{"tool": tool, "mappings": h.mitm.mappings[tool]})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mappings": h.mitm.mappings})
}

func (h *Handler) PutMITMMappings(c *gin.Context) {
	var body struct {
		Tool     string            `json:"tool"`
		Mappings map[string]string `json:"mappings"`
	}
	if errBind := c.ShouldBindJSON(&body); errBind != nil || body.Mappings == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tool and mappings are required"})
		return
	}
	body.Tool = strings.ToLower(strings.TrimSpace(body.Tool))
	if len(mitmToolHosts[body.Tool]) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported MITM tool"})
		return
	}
	if !mitmDNSStatus()[body.Tool] {
		c.JSON(http.StatusForbidden, gin.H{"error": "enable DNS for this tool before editing its model mappings"})
		return
	}
	clean := make(map[string]string, len(body.Mappings))
	for source, target := range body.Mappings {
		source, target = strings.TrimSpace(source), strings.TrimSpace(target)
		if strings.ContainsAny(source+target, "\r\n\x00") || len(source) > 256 || len(target) > 512 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model names must not contain control characters and must be within size limits"})
			return
		}
		if source != "" {
			clean[source] = target
		}
	}
	h.mitm.mu.Lock()
	previous, existed := h.mitm.mappings[body.Tool]
	h.mitm.mappings[body.Tool] = clean
	errSave := h.mitm.saveMappingsLocked()
	if errSave != nil {
		if existed {
			h.mitm.mappings[body.Tool] = previous
		} else {
			delete(h.mitm.mappings, body.Tool)
		}
	}
	h.mitm.mu.Unlock()
	if errSave != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errSave.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "tool": body.Tool, "mappings": clean})
}
