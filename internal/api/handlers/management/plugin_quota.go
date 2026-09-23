package management

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/misc"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	xcurrency "golang.org/x/text/currency"
)

type credentialQuotaRequest struct {
	AuthIndexSnake  *string `json:"auth_index"`
	AuthIndexCamel  *string `json:"authIndex"`
	AuthIndexPascal *string `json:"AuthIndex"`
	PluginID        string  `json:"plugin_id"`
	Provider        string  `json:"provider"`
}

var errCodexReauthRequired = errors.New("Codex OAuth sign-in required")

func (r credentialQuotaRequest) resolveAuthIndex() string {
	if r.AuthIndexSnake != nil && strings.TrimSpace(*r.AuthIndexSnake) != "" {
		return strings.TrimSpace(*r.AuthIndexSnake)
	}
	if r.AuthIndexCamel != nil && strings.TrimSpace(*r.AuthIndexCamel) != "" {
		return strings.TrimSpace(*r.AuthIndexCamel)
	}
	if r.AuthIndexPascal != nil && strings.TrimSpace(*r.AuthIndexPascal) != "" {
		return strings.TrimSpace(*r.AuthIndexPascal)
	}
	return ""
}

// GetQuotaProviders returns the list of registered quota providers.
func (h *Handler) GetQuotaProviders(c *gin.Context) {
	if h == nil {
		c.JSON(http.StatusOK, gin.H{"providers": []any{}})
		return
	}
	h.mu.Lock()
	host := h.pluginHost
	h.mu.Unlock()
	if host == nil {
		c.JSON(http.StatusOK, gin.H{"providers": []any{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": host.QuotaProviders(c.Request.Context())})
}

// FetchCredentialQuota retrieves normalized quota for a credential via its quota provider or declarative probe.
func (h *Handler) FetchCredentialQuota(c *gin.Context) {
	var body credentialQuotaRequest
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	authIndex := body.resolveAuthIndex()
	if authIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_index is required"})
		return
	}

	auth := h.authByIndex(authIndex)
	if auth == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth not found"})
		return
	}

	h.mu.Lock()
	host := h.pluginHost
	h.mu.Unlock()

	pluginID := strings.TrimSpace(body.PluginID)
	provider := strings.TrimSpace(body.Provider)
	if provider == "" {
		provider = auth.Provider
	}
	if native, handledNative, errNative := h.fetchNativeQuota(c.Request.Context(), auth); handledNative {
		if errNative != nil {
			if errors.Is(errNative, errCodexReauthRequired) {
				c.JSON(http.StatusBadGateway, gin.H{"error": gin.H{"code": "reauth_required", "message": errCodexReauthRequired.Error()}})
				return
			}
			log.WithError(errNative).Warnf("native quota fetch failed for provider %s", provider)
			c.JSON(http.StatusBadGateway, gin.H{"error": errNative.Error()})
			return
		}
		c.JSON(http.StatusOK, native)
		return
	}

	if host != nil {
		req := pluginapi.QuotaFetchRequest{
			AuthIndex:  auth.Index,
			AuthID:     auth.ID,
			Provider:   provider,
			Metadata:   auth.Metadata,
			Attributes: auth.Attributes,
		}
		var quotaResp pluginapi.QuotaFetchResponse
		var handled bool
		var errFetch error
		if pluginID != "" {
			quotaResp, handled, errFetch = host.FetchQuotaByPlugin(c.Request.Context(), pluginID, req)
		} else {
			quotaResp, handled, errFetch = host.FetchQuota(c.Request.Context(), req)
		}
		if handled {
			if errFetch != nil {
				log.WithError(errFetch).Warnf("failed to fetch quota for credential %s", auth.Index)
				c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to fetch quota: %v", errFetch)})
				return
			}
			c.JSON(http.StatusOK, quotaResp)
			return
		}
	}

	// Fallback to declarative quota probe if configured in metadata
	if auth.Metadata != nil {
		if rawProbe, okProbe := auth.Metadata["quota_probe"]; okProbe && rawProbe != nil {
			if probeMap, okMap := rawProbe.(map[string]any); okMap {
				quotaResp, handledProbe, errProbe := h.executeQuotaProbe(c, auth, probeMap)
				if handledProbe {
					if errProbe != nil {
						c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("quota probe failed: %v", errProbe)})
						return
					}
					c.JSON(http.StatusOK, quotaResp)
					return
				}
			}
		}
	}

	// Claude, Codex, and Devin expose quota watermarks on ordinary upstream responses.
	// Return the latest observed snapshot so the tracker remains useful without a plugin.
	if supportsManagementQuota(auth.Provider) {
		c.JSON(http.StatusOK, quotaObservationResponse(auth))
		return
	}

	c.JSON(http.StatusNotImplemented, gin.H{"error": "no quota provider available for credential"})
}

// fetchNativeQuota mirrors 9router's provider-specific usage fetchers. It reads
// quota APIs directly and never sends a generation request, so opening the tracker
// does not consume model quota.
func (h *Handler) fetchNativeQuota(ctx context.Context, auth *coreauth.Auth) (pluginapi.QuotaFetchResponse, bool, error) {
	if auth == nil {
		return pluginapi.QuotaFetchResponse{}, false, nil
	}
	provider := strings.ToLower(strings.TrimSpace(auth.Provider))
	if provider != "codex" && provider != "claude" && provider != "antigravity" {
		return pluginapi.QuotaFetchResponse{}, false, nil
	}
	token, errToken := h.resolveTokenForAuth(ctx, auth, "")
	if errToken != nil || strings.TrimSpace(token) == "" {
		if errToken != nil {
			return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("quota authentication failed: %w", errToken)
		}
		return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("quota authentication token not found")
	}

	client := &http.Client{Transport: h.apiCallTransport(auth, "")}
	requestJSON := func(method, endpoint string, body []byte, headers map[string]string) ([]byte, int, error) {
		var reader io.Reader
		if len(body) > 0 {
			reader = strings.NewReader(string(body))
		}
		req, errReq := http.NewRequestWithContext(ctx, method, endpoint, reader)
		if errReq != nil {
			return nil, 0, errReq
		}
		for key, value := range headers {
			req.Header.Set(key, value)
		}
		resp, errDo := client.Do(req)
		if errDo != nil {
			return nil, 0, errDo
		}
		defer func() { _ = resp.Body.Close() }()
		data, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return nil, resp.StatusCode, errRead
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return data, resp.StatusCode, fmt.Errorf("quota API returned status %d", resp.StatusCode)
		}
		return data, resp.StatusCode, nil
	}

	baseHeaders := map[string]string{"Authorization": "Bearer " + token, "Accept": "application/json"}
	switch provider {
	case "codex":
		if accountID := metadataString(auth, "account_id", "accountId", "chatgptAccountId"); accountID != "" {
			baseHeaders["ChatGPT-Account-ID"] = accountID
		}
		baseHeaders["Originator"] = "codex_cli_rs"
		data, status, errFetch := requestJSON(http.MethodGet, "https://chatgpt.com/backend-api/wham/usage", nil, baseHeaders)
		if status == http.StatusUnauthorized {
			refreshed, errRefresh := h.authManager.ForceRefreshAuth(ctx, auth.ID)
			if errRefresh != nil {
				if codexRefreshRequiresSignIn(errRefresh) {
					return pluginapi.QuotaFetchResponse{}, true, errCodexReauthRequired
				}
				return pluginapi.QuotaFetchResponse{}, true, errors.New("Codex quota token refresh failed; retry later")
			}
			if refreshed == nil || tokenValueForAuth(refreshed) == "" {
				return pluginapi.QuotaFetchResponse{}, true, errCodexReauthRequired
			}
			baseHeaders["Authorization"] = "Bearer " + tokenValueForAuth(refreshed)
			if accountID := metadataString(refreshed, "account_id", "accountId", "chatgptAccountId"); accountID != "" {
				baseHeaders["ChatGPT-Account-ID"] = accountID
			}
			data, status, errFetch = requestJSON(http.MethodGet, "https://chatgpt.com/backend-api/wham/usage", nil, baseHeaders)
			if status == http.StatusUnauthorized {
				return pluginapi.QuotaFetchResponse{}, true, errCodexReauthRequired
			}
		}
		if errFetch != nil {
			return pluginapi.QuotaFetchResponse{}, true, errFetch
		}
		return parseCodexNativeQuota(data), true, nil
	case "claude":
		headers := map[string]string{"Authorization": "Bearer " + token, "Accept": "application/json", "anthropic-beta": "oauth-2025-04-20", "anthropic-version": "2023-06-01"}
		data, _, errFetch := requestJSON(http.MethodGet, "https://api.anthropic.com/api/oauth/usage", nil, headers)
		if errFetch != nil {
			return pluginapi.QuotaFetchResponse{}, true, errFetch
		}
		return parseClaudeNativeQuota(data), true, nil
	case "antigravity":
		project := metadataString(auth, "project_id", "cloudaicompanionProject", "projectId")
		if project == "" {
			loadBody, _ := json.Marshal(map[string]any{"metadata": map[string]string{"ideType": "ANTIGRAVITY"}})
			loadHeaders := map[string]string{"Authorization": "Bearer " + token, "Accept": "application/json", "Content-Type": "application/json", "User-Agent": misc.AntigravityLoadCodeAssistUserAgent("")}
			loadData, _, errLoad := requestJSON(http.MethodPost, "https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist", loadBody, loadHeaders)
			if errLoad == nil {
				project = firstJSONString(loadData, "cloudaicompanionProject", "projectId", "project")
			}
		}
		body := map[string]any{}
		if project != "" {
			body["project"] = project
		}
		requestBody, _ := json.Marshal(body)
		headers := map[string]string{"Authorization": "Bearer " + token, "Accept": "application/json", "Content-Type": "application/json", "User-Agent": misc.AntigravityRequestUserAgent(""), "X-Client-Name": "antigravity"}
		data, _, errFetch := requestJSON(http.MethodPost, "https://daily-cloudcode-pa.googleapis.com/v1internal:fetchAvailableModels", requestBody, headers)
		if errFetch != nil {
			return pluginapi.QuotaFetchResponse{}, true, errFetch
		}
		return parseAntigravityNativeQuota(data, antigravityVisibleModels(auth)), true, nil
	}
	return pluginapi.QuotaFetchResponse{}, false, nil
}

func codexRefreshRequiresSignIn(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "invalid_grant") ||
		strings.Contains(message, "refresh_token_reused") ||
		strings.Contains(message, "refresh_token_revoked") ||
		strings.Contains(message, "token_revoked") ||
		strings.Contains(message, "token refresh failed with status 401")
}

func parseCodexNativeQuota(data []byte) pluginapi.QuotaFetchResponse {
	result := pluginapi.QuotaFetchResponse{}
	result.Subscription = &pluginapi.QuotaSubscription{Plan: gjson.GetBytes(data, "plan_type").String()}
	rateLimit := gjson.GetBytes(data, "rate_limit")
	if !rateLimit.Exists() {
		rateLimit = gjson.GetBytes(data, "rate_limits")
	}
	if !rateLimit.Exists() {
		rateLimit = gjson.GetBytes(data, "rate_limits_by_limit_id.codex")
	}
	for _, item := range []struct {
		keys  []string
		label string
	}{{[]string{"primary_window", "primary"}, "Primary (5h)"}, {[]string{"secondary_window", "secondary"}, "Weekly"}} {
		window := gjson.Result{}
		for _, key := range item.keys {
			window = rateLimit.Get(key)
			if window.Exists() {
				break
			}
		}
		if !window.Exists() {
			continue
		}
		used := window.Get("used_percent").Float()
		result.Groups = append(result.Groups, pluginapi.QuotaGroup{DisplayName: item.label, Buckets: []pluginapi.QuotaBucket{{Window: item.label, RemainingFraction: math.Max(0, math.Min(1, (100-used)/100)), ResetTime: gjsonResetTime(window.Get("reset_at"))}}})
	}
	return result
}

func parseClaudeNativeQuota(data []byte) pluginapi.QuotaFetchResponse {
	result := pluginapi.QuotaFetchResponse{Subscription: &pluginapi.QuotaSubscription{Plan: "Claude Code"}}
	for _, item := range []struct{ key, label string }{{"five_hour", "Session (5h)"}, {"seven_day", "Weekly (7d)"}} {
		window := gjson.GetBytes(data, item.key)
		if !window.Exists() {
			continue
		}
		used := window.Get("utilization").Float()
		result.Groups = append(result.Groups, pluginapi.QuotaGroup{DisplayName: item.label, Buckets: []pluginapi.QuotaBucket{{Window: item.label, RemainingFraction: math.Max(0, math.Min(1, (100-used)/100)), ResetTime: gjsonResetTime(window.Get("resets_at"))}}})
	}
	return result
}

func parseAntigravityNativeQuota(data []byte, allowedNames ...map[string]string) pluginapi.QuotaFetchResponse {
	result := pluginapi.QuotaFetchResponse{}
	allowed := map[string]string(nil)
	if len(allowedNames) > 0 {
		allowed = allowedNames[0]
	}
	seen := make(map[string]struct{})
	gjson.GetBytes(data, "models").ForEach(func(key, value gjson.Result) bool {
		quota := value.Get("quotaInfo")
		if !quota.Exists() {
			return true
		}
		displayName := value.Get("displayName").String()
		if displayName == "" {
			displayName = key.String()
		}
		lowerName := strings.ToLower(displayName)
		if strings.HasPrefix(lowerName, "tab_") || strings.Contains(lowerName, "image") ||
			(!strings.Contains(lowerName, "pro") && !strings.Contains(lowerName, "flash") && !strings.Contains(lowerName, "gpt") && !strings.Contains(lowerName, "claude")) {
			return true
		}
		if len(allowed) > 0 {
			canonical, ok := allowed[lowerName]
			if !ok {
				canonical, ok = allowed[antigravityModelNameKey(displayName)]
			}
			if !ok {
				return true
			}
			displayName = canonical
		}
		if _, exists := seen[displayName]; exists {
			return true
		}
		seen[displayName] = struct{}{}
		remaining := math.Max(0, math.Min(1, quota.Get("remainingFraction").Float()))
		resetTime := quota.Get("resetTime").String()
		result.Groups = append(result.Groups, pluginapi.QuotaGroup{DisplayName: displayName, Buckets: []pluginapi.QuotaBucket{{RemainingFraction: remaining, ResetTime: resetTime}}})
		return true
	})
	if len(allowed) > 0 {
		bucketsByBase := make(map[string]pluginapi.QuotaBucket, len(result.Groups))
		for _, group := range result.Groups {
			if len(group.Buckets) > 0 {
				bucketsByBase[antigravityModelNameKey(group.DisplayName)] = group.Buckets[0]
			}
		}
		added := make(map[string]struct{}, len(result.Groups))
		for _, group := range result.Groups {
			added[group.DisplayName] = struct{}{}
		}
		visibleNames := make(map[string]string, len(allowed))
		for _, name := range allowed {
			visibleNames[strings.ToLower(name)] = name
		}
		for _, name := range visibleNames {
			lowerName := strings.ToLower(name)
			if strings.HasPrefix(lowerName, "tab_") || strings.Contains(lowerName, "image") ||
				(!strings.Contains(lowerName, "pro") && !strings.Contains(lowerName, "flash") && !strings.Contains(lowerName, "gpt") && !strings.Contains(lowerName, "claude")) {
				continue
			}
			if _, exists := added[name]; exists {
				continue
			}
			if bucket, ok := bucketsByBase[antigravityModelNameKey(name)]; ok {
				result.Groups = append(result.Groups, pluginapi.QuotaGroup{DisplayName: name, Buckets: []pluginapi.QuotaBucket{bucket}})
				added[name] = struct{}{}
			}
		}
	}
	sort.Slice(result.Groups, func(i, j int) bool { return result.Groups[i].DisplayName < result.Groups[j].DisplayName })
	return result
}

func antigravityModelNameKey(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, suffix := range []string{" (high)", " (medium)", " (low)"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return name
}

func antigravityVisibleModels(auth *coreauth.Auth) map[string]string {
	if auth == nil {
		return nil
	}
	models := registry.GetGlobalRegistry().GetModelsForClient(auth.ID)
	if len(models) == 0 {
		return nil
	}
	allowed := make(map[string]string, len(models))
	for _, model := range models {
		if model == nil {
			continue
		}
		name := strings.TrimSpace(model.DisplayName)
		if name == "" {
			name = strings.TrimSpace(model.ID)
		}
		if name != "" {
			lowerName := strings.ToLower(strings.TrimSpace(name))
			allowed[lowerName] = name
			if antigravityModelNameKey(name) == lowerName {
				allowed[antigravityModelNameKey(name)] = name
			}
		}
	}
	return allowed
}

func gjsonResetTime(value gjson.Result) string {
	if value.Type == gjson.Number {
		unix := value.Int()
		if unix > 0 && unix < 1000000000000 {
			unix *= 1000
		}
		return time.UnixMilli(unix).UTC().Format(time.RFC3339)
	}
	return value.String()
}

func metadataString(auth *coreauth.Auth, keys ...string) string {
	if auth == nil || auth.Metadata == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := auth.Metadata[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstJSONString(data []byte, keys ...string) string {
	for _, key := range keys {
		if value := gjson.GetBytes(data, key); value.Exists() && value.Type == gjson.String && strings.TrimSpace(value.String()) != "" {
			return strings.TrimSpace(value.String())
		}
	}
	return ""
}

func quotaObservationResponse(auth *coreauth.Auth) gin.H {
	response := gin.H{"groups": []pluginapi.QuotaGroup{}, "summary": []pluginapi.QuotaMetric{}, "signals": map[string]string{}}
	if auth == nil {
		return response
	}
	if strings.EqualFold(strings.TrimSpace(auth.Provider), "antigravity") {
		if hint, okHint := coreauth.GetAntigravityCreditsHint(auth.ID); okHint && hint.Known {
			response["credits_available"] = hint.Available
			response["summary"] = []pluginapi.QuotaMetric{
				{Key: "credit_amount", Label: "AI credits", Value: hint.CreditAmount, Unit: "credits"},
				{Key: "minimum_credit_amount", Label: "Minimum credit", Value: hint.MinCreditAmount, Unit: "credits"},
			}
		}
	}
	signals := make(map[string]string, len(auth.Quota.Signals))
	for key, value := range auth.Quota.Signals {
		signals[key] = value
	}
	response["signals"] = signals
	if !auth.Quota.ObservedAt.IsZero() {
		response["observed_at"] = auth.Quota.ObservedAt
	}
	keys := make([]string, 0, len(signals))
	for key := range signals {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	groups := make([]pluginapi.QuotaGroup, 0)
	for _, key := range keys {
		value := strings.TrimSpace(signals[key])
		lower := strings.ToLower(key)
		percent, okPercent := quotaPercent(value)
		if !okPercent || (!strings.Contains(lower, "used-percent") && !strings.Contains(lower, "remaining-percent")) {
			continue
		}
		remaining := percent
		if strings.Contains(lower, "used-percent") {
			remaining = 1 - percent
		}
		if remaining < 0 {
			remaining = 0
		}
		if remaining > 1 {
			remaining = 1
		}
		groups = append(groups, pluginapi.QuotaGroup{DisplayName: key, Buckets: []pluginapi.QuotaBucket{{RemainingFraction: remaining, Description: value}}})
	}
	response["groups"] = groups
	if len(groups) == 0 && len(signals) > 0 {
		response["summary"] = []pluginapi.QuotaMetric{{Key: "observed_signals", Label: "Observed quota signals", Value: float64(len(signals)), Unit: "headers"}}
	}
	return response
}

func supportsManagementQuota(provider string) bool {
	return coreauth.ProviderSupportsQuotaObservation(provider) || strings.EqualFold(strings.TrimSpace(provider), "antigravity")
}

func quotaPercent(value string) (float64, bool) {
	value = strings.TrimSpace(strings.TrimSuffix(value, "%"))
	parsed, errParse := strconv.ParseFloat(value, 64)
	if errParse != nil {
		return 0, false
	}
	if parsed > 1 {
		parsed /= 100
	}
	return parsed, parsed >= 0 && parsed <= 1
}

// ResetCredentialQuota resets quota or usage for a credential.
func (h *Handler) ResetCredentialQuota(c *gin.Context) {
	var body credentialQuotaRequest
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	authIndex := body.resolveAuthIndex()
	if authIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_index is required"})
		return
	}

	auth := h.authByIndex(authIndex)
	if auth == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth not found"})
		return
	}

	h.mu.Lock()
	host := h.pluginHost
	h.mu.Unlock()

	if host == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "plugin host unavailable"})
		return
	}

	pluginID := strings.TrimSpace(body.PluginID)
	provider := strings.TrimSpace(body.Provider)
	if provider == "" {
		provider = auth.Provider
	}

	req := pluginapi.QuotaResetRequest{
		AuthIndex:  auth.Index,
		AuthID:     auth.ID,
		Provider:   provider,
		Metadata:   auth.Metadata,
		Attributes: auth.Attributes,
	}

	var resetResp pluginapi.QuotaResetResponse
	var handled bool
	var errReset error

	if pluginID != "" {
		if !host.HasQuotaProviderForPlugin(pluginID) {
			c.JSON(http.StatusNotFound, gin.H{"error": "quota provider not found for plugin"})
			return
		}
		resetResp, handled, errReset = host.ResetQuotaByPlugin(c.Request.Context(), pluginID, req)
		if !handled {
			c.JSON(http.StatusNotFound, gin.H{"error": "quota provider not found for plugin"})
			return
		}
	} else {
		if !host.HasQuotaProviderContext(c.Request.Context(), provider) {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "no quota provider available for credential to reset"})
			return
		}
		resetResp, handled, errReset = host.ResetQuota(c.Request.Context(), req)
		if !handled {
			c.JSON(http.StatusBadGateway, gin.H{"error": "quota provider did not handle reset request"})
			return
		}
	}

	if errReset != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("plugin quota reset failed: %v", errReset)})
		return
	}
	if !resetResp.Success {
		msg := resetResp.Message
		if msg == "" {
			msg = "quota reset rejected by provider"
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}

	if h.authManager != nil {
		updated, _, errResetCore := h.authManager.ResetQuota(c.Request.Context(), auth.ID)
		if errResetCore != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to reset routing quota: %v", errResetCore)})
			return
		}
		if updated != nil {
			updated.EnsureIndex()
		}
	}

	resp := gin.H{
		"status":     "ok",
		"auth_index": auth.Index,
	}
	if resetResp.Message != "" {
		resp["message"] = resetResp.Message
	}
	c.JSON(http.StatusOK, resp)
}

// GetPluginQuota handles GET /v0/management/plugins/:id/quota?auth_index=...
func (h *Handler) GetPluginQuota(c *gin.Context) {
	pluginID := strings.TrimSpace(c.Param("id"))
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		authIndex = strings.TrimSpace(c.Query("authIndex"))
	}
	if authIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_index is required"})
		return
	}
	h.fetchQuotaForPlugin(c, pluginID, authIndex)
}

// FetchPluginQuota handles POST /v0/management/plugins/:id/quota
func (h *Handler) FetchPluginQuota(c *gin.Context) {
	pluginID := strings.TrimSpace(c.Param("id"))
	var body credentialQuotaRequest
	if errBind := c.ShouldBindJSON(&body); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	authIndex := body.resolveAuthIndex()
	if authIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_index is required"})
		return
	}
	h.fetchQuotaForPlugin(c, pluginID, authIndex)
}

func (h *Handler) fetchQuotaForPlugin(c *gin.Context, pluginID, authIndex string) {
	auth := h.authByIndex(authIndex)
	if auth == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth not found"})
		return
	}

	h.mu.Lock()
	host := h.pluginHost
	h.mu.Unlock()

	if host == nil || !host.HasQuotaProviderForPlugin(pluginID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "quota provider not found for plugin"})
		return
	}

	quotaResp, handled, errFetch := host.FetchQuotaByPlugin(c.Request.Context(), pluginID, pluginapi.QuotaFetchRequest{
		AuthIndex:  auth.Index,
		AuthID:     auth.ID,
		Provider:   auth.Provider,
		Metadata:   auth.Metadata,
		Attributes: auth.Attributes,
	})
	if !handled {
		c.JSON(http.StatusNotFound, gin.H{"error": "quota provider not found for plugin"})
		return
	}
	if errFetch != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to fetch quota: %v", errFetch)})
		return
	}
	c.JSON(http.StatusOK, quotaResp)
}

// ResetPluginQuota handles DELETE /v0/management/plugins/:id/quota and POST /v0/management/plugins/:id/quota/reset
func (h *Handler) ResetPluginQuota(c *gin.Context) {
	pluginID := strings.TrimSpace(c.Param("id"))
	authIndex := strings.TrimSpace(c.Query("auth_index"))
	if authIndex == "" {
		authIndex = strings.TrimSpace(c.Query("authIndex"))
	}
	if authIndex == "" {
		var body credentialQuotaRequest
		_ = c.ShouldBindJSON(&body)
		authIndex = body.resolveAuthIndex()
	}
	if authIndex == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_index is required"})
		return
	}

	auth := h.authByIndex(authIndex)
	if auth == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth not found"})
		return
	}

	h.mu.Lock()
	host := h.pluginHost
	h.mu.Unlock()

	if host == nil || !host.HasQuotaProviderForPlugin(pluginID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "quota provider not found for plugin"})
		return
	}

	resetResp, handled, errReset := host.ResetQuotaByPlugin(c.Request.Context(), pluginID, pluginapi.QuotaResetRequest{
		AuthIndex:  auth.Index,
		AuthID:     auth.ID,
		Provider:   auth.Provider,
		Metadata:   auth.Metadata,
		Attributes: auth.Attributes,
	})
	if !handled {
		c.JSON(http.StatusNotFound, gin.H{"error": "quota provider not found for plugin"})
		return
	}
	if errReset != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("failed to reset quota: %v", errReset)})
		return
	}
	if !resetResp.Success {
		msg := resetResp.Message
		if msg == "" {
			msg = "quota reset rejected by plugin"
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}

	if h.authManager != nil {
		updated, _, errResetCore := h.authManager.ResetQuota(c.Request.Context(), auth.ID)
		if errResetCore != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to reset routing quota: %v", errResetCore)})
			return
		}
		if updated != nil {
			updated.EnsureIndex()
		}
	}

	resp := gin.H{
		"status":     "ok",
		"auth_index": auth.Index,
	}
	if resetResp.Message != "" {
		resp["message"] = resetResp.Message
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) executeQuotaProbe(c *gin.Context, auth *coreauth.Auth, probe map[string]any) (pluginapi.QuotaFetchResponse, bool, error) {
	urlStr, _ := probe["url"].(string)
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" {
		return pluginapi.QuotaFetchResponse{}, false, nil
	}

	method, _ := probe["method"].(string)
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}

	needsToken := strings.Contains(urlStr, "$TOKEN$")
	rawData, _ := probe["data"].(string)
	if strings.Contains(rawData, "$TOKEN$") {
		needsToken = true
	}
	headers, _ := probe["header"].(map[string]any)
	if headers == nil {
		headers, _ = probe["headers"].(map[string]any)
	}
	for _, v := range headers {
		if strVal, okVal := v.(string); okVal && strings.Contains(strVal, "$TOKEN$") {
			needsToken = true
			break
		}
	}

	var token string
	if needsToken {
		var errToken error
		token, errToken = h.resolveTokenForAuth(c.Request.Context(), auth, "")
		if errToken != nil {
			return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("probe authentication failed: %w", errToken)
		}
		if token == "" {
			return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("probe authentication token not found for credential")
		}
		urlStr = strings.ReplaceAll(urlStr, "$TOKEN$", token)
		rawData = strings.ReplaceAll(rawData, "$TOKEN$", token)
	}

	var reqBody io.Reader
	if rawData != "" {
		reqBody = strings.NewReader(rawData)
	}

	req, errReq := http.NewRequestWithContext(c.Request.Context(), method, urlStr, reqBody)
	if errReq != nil {
		return pluginapi.QuotaFetchResponse{}, false, fmt.Errorf("build probe request: %w", errReq)
	}

	for k, v := range headers {
		if strVal, okVal := v.(string); okVal {
			if needsToken {
				strVal = strings.ReplaceAll(strVal, "$TOKEN$", token)
			}
			req.Header.Set(k, strVal)
		}
	}

	client := &http.Client{
		Transport: h.apiCallTransport(auth, ""),
	}

	resp, errDo := client.Do(req)
	if errDo != nil {
		return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("probe request failed: %w", errDo)
	}
	defer func() {
		if errClose := resp.Body.Close(); errClose != nil {
			log.Errorf("probe response body close error: %v", errClose)
		}
	}()

	respBytes, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("read probe response: %w", errRead)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("probe returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	if !json.Valid(respBytes) {
		return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("upstream probe response is not valid JSON")
	}

	var serverOffsetMs int64
	if dateHeader := resp.Header.Get("Date"); dateHeader != "" {
		if parsedDate, errDate := http.ParseTime(dateHeader); errDate == nil {
			serverOffsetMs = parsedDate.Sub(time.Now()).Milliseconds()
		}
	}

	if mappingRaw, okMapping := probe["mapping"].(map[string]any); okMapping {
		mappedResp, errMap := mapProbeResponse(respBytes, mappingRaw)
		if errMap != nil {
			return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("probe response mapping failed: %w", errMap)
		}
		if mappedResp.ServerTimeOffsetMs == 0 {
			mappedResp.ServerTimeOffsetMs = serverOffsetMs
		}
		return mappedResp, true, nil
	}

	var rawQuota map[string]json.RawMessage
	if errRaw := json.Unmarshal(respBytes, &rawQuota); errRaw == nil {
		for key := range rawQuota {
			if strings.EqualFold(key, "summary") {
				delete(rawQuota, key) // Optional plugin data must not invalidate core quota fields.
			}
		}
		if coreQuotaJSON, errMarshal := json.Marshal(rawQuota); errMarshal == nil {
			var quotaResp pluginapi.QuotaFetchResponse
			if errJSON := json.Unmarshal(coreQuotaJSON, &quotaResp); errJSON == nil {
				hasPlan := quotaResp.Subscription != nil && strings.TrimSpace(quotaResp.Subscription.Plan) != ""
				// Filter groups to retain only buckets with an actual valid numeric remaining fraction in raw JSON
				filteredGroups := make([]pluginapi.QuotaGroup, 0, len(quotaResp.Groups))
				groupsRes := gjson.GetBytes(respBytes, "groups")
				if groupsRes.IsArray() {
					for gIdx, grp := range groupsRes.Array() {
						if gIdx >= len(quotaResp.Groups) {
							break
						}
						bucketsRes := grp.Get("buckets")
						if !bucketsRes.IsArray() {
							continue
						}
						origGroup := quotaResp.Groups[gIdx]
						validBuckets := make([]pluginapi.QuotaBucket, 0, len(origGroup.Buckets))
						for bIdx, bkt := range bucketsRes.Array() {
							if bIdx >= len(origGroup.Buckets) {
								break
							}
							remFrac := bkt.Get("remainingFraction")
							if !remFrac.Exists() {
								remFrac = bkt.Get("remaining_fraction")
							}
							if fracVal, okNum := parseNumericFraction(remFrac); okNum {
								bucket := origGroup.Buckets[bIdx]
								bucket.RemainingFraction = fracVal
								validBuckets = append(validBuckets, bucket)
							}
						}
						if len(validBuckets) > 0 {
							origGroup.Buckets = validBuckets
							filteredGroups = append(filteredGroups, origGroup)
						}
					}
				}
				quotaResp.Groups = filteredGroups
				quotaResp.Summary = filterUsableQuotaSummary(respBytes)
				hasValidBuckets := len(filteredGroups) > 0
				hasValidSummary := len(quotaResp.Summary) > 0
				if hasPlan || hasValidBuckets || hasValidSummary {
					if quotaResp.ServerTimeOffsetMs == 0 {
						quotaResp.ServerTimeOffsetMs = serverOffsetMs
					}
					return quotaResp, true, nil
				}
			}
		}
	}

	return pluginapi.QuotaFetchResponse{}, true, fmt.Errorf("upstream probe response does not match normalized quota shape or declared mapping")
}

func filterUsableQuotaSummary(raw []byte) []pluginapi.QuotaMetric {
	var rawQuota map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawQuota); err != nil {
		return nil
	}
	rawSummary, ok := rawQuota["summary"]
	if !ok {
		for key, value := range rawQuota {
			if strings.EqualFold(key, "summary") {
				rawSummary = value
				break
			}
		}
	}
	summaryResult := gjson.ParseBytes(rawSummary)
	if !summaryResult.IsArray() {
		return nil
	}
	usable := make([]pluginapi.QuotaMetric, 0, len(summaryResult.Array()))
	for _, rawMetric := range summaryResult.Array() {
		keyResult := rawMetric.Get("key")
		labelResult := rawMetric.Get("label")
		key := strings.TrimSpace(keyResult.String())
		label := strings.TrimSpace(labelResult.String())
		value := rawMetric.Get("value")
		if keyResult.Type != gjson.String || labelResult.Type != gjson.String || key == "" || label == "" || value.Type != gjson.Number || math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0) {
			continue
		}
		metric := pluginapi.QuotaMetric{
			Key:   key,
			Label: label,
			Value: value.Float(),
		}
		if unitResult := rawMetric.Get("unit"); unitResult.Type == gjson.String {
			metric.Unit = strings.TrimSpace(unitResult.String())
		}
		if formatResult := rawMetric.Get("format"); formatResult.Type == gjson.String {
			format := strings.TrimSpace(formatResult.String())
			switch format {
			case "number":
				metric.Format = format
			case "currency":
				if currencyResult := rawMetric.Get("currency"); currencyResult.Type == gjson.String {
					code := strings.ToUpper(strings.TrimSpace(currencyResult.String()))
					if _, err := xcurrency.ParseISO(code); err == nil {
						metric.Format = format
						metric.Currency = code
					}
				}
			}
		}
		usable = append(usable, metric)
	}
	return usable
}

func parseNumericFraction(res gjson.Result) (float64, bool) {
	if !res.Exists() {
		return 0, false
	}
	switch res.Type {
	case gjson.Number:
		val := res.Float()
		if math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, false
		}
		return val, true
	case gjson.String:
		s := strings.TrimSpace(res.String())
		if s == "" {
			return 0, false
		}
		val, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
			return 0, false
		}
		return val, true
	default:
		return 0, false
	}
}

func mapProbeResponse(respBytes []byte, mapping map[string]any) (pluginapi.QuotaFetchResponse, error) {
	out := pluginapi.QuotaFetchResponse{}
	if planPath, ok := mapping["plan"].(string); ok && planPath != "" {
		if res := gjson.GetBytes(respBytes, planPath); res.Exists() && strings.TrimSpace(res.String()) != "" {
			if out.Subscription == nil {
				out.Subscription = &pluginapi.QuotaSubscription{}
			}
			out.Subscription.Plan = res.String()
		}
	}
	if tierPath, ok := mapping["tier_name"].(string); ok && tierPath != "" {
		if res := gjson.GetBytes(respBytes, tierPath); res.Exists() && strings.TrimSpace(res.String()) != "" {
			if out.Subscription == nil {
				out.Subscription = &pluginapi.QuotaSubscription{}
			}
			out.Subscription.TierName = res.String()
		}
	} else if tierPath, ok := mapping["tierName"].(string); ok && tierPath != "" {
		if res := gjson.GetBytes(respBytes, tierPath); res.Exists() && strings.TrimSpace(res.String()) != "" {
			if out.Subscription == nil {
				out.Subscription = &pluginapi.QuotaSubscription{}
			}
			out.Subscription.TierName = res.String()
		}
	}
	if tierIDPath, ok := mapping["tier_id"].(string); ok && tierIDPath != "" {
		if res := gjson.GetBytes(respBytes, tierIDPath); res.Exists() && strings.TrimSpace(res.String()) != "" {
			if out.Subscription == nil {
				out.Subscription = &pluginapi.QuotaSubscription{}
			}
			out.Subscription.TierID = res.String()
		}
	} else if tierIDPath, ok := mapping["tierId"].(string); ok && tierIDPath != "" {
		if res := gjson.GetBytes(respBytes, tierIDPath); res.Exists() && strings.TrimSpace(res.String()) != "" {
			if out.Subscription == nil {
				out.Subscription = &pluginapi.QuotaSubscription{}
			}
			out.Subscription.TierID = res.String()
		}
	}

	rawGroups, okGroups := mapping["groups"].([]any)
	if okGroups {
		for _, rg := range rawGroups {
			gm, okGM := rg.(map[string]any)
			if !okGM {
				continue
			}
			group := pluginapi.QuotaGroup{}
			if namePath, okName := gm["display_name"].(string); okName {
				if res := gjson.GetBytes(respBytes, namePath); res.Exists() && strings.TrimSpace(res.String()) != "" {
					group.DisplayName = res.String()
				} else {
					group.DisplayName = namePath
				}
			} else if namePath, okName := gm["displayName"].(string); okName {
				if res := gjson.GetBytes(respBytes, namePath); res.Exists() && strings.TrimSpace(res.String()) != "" {
					group.DisplayName = res.String()
				} else {
					group.DisplayName = namePath
				}
			}

			if bucketsPath, okBP := gm["buckets_path"].(string); okBP && bucketsPath != "" {
				arrayRes := gjson.GetBytes(respBytes, bucketsPath)
				if arrayRes.IsArray() && len(arrayRes.Array()) > 0 {
					winKey, _ := gm["window_key"].(string)
					if winKey == "" {
						winKey = "window"
					}
					remFracKey, _ := gm["remaining_fraction_key"].(string)
					if remFracKey == "" {
						remFracKey = "remaining_fraction"
					}
					remAmtKey, _ := gm["remaining_amount_key"].(string)
					totAmtKey, _ := gm["total_amount_key"].(string)
					resetKey, _ := gm["reset_time_key"].(string)
					if resetKey == "" {
						resetKey = "reset_time"
					}
					descKey, _ := gm["description_key"].(string)
					if descKey == "" {
						descKey = "description"
					}

					for _, item := range arrayRes.Array() {
						var hasFraction bool
						var frac float64
						if remFracKey != "" {
							if val, ok := parseNumericFraction(item.Get(remFracKey)); ok {
								frac = val
								hasFraction = true
							}
						}
						if !hasFraction && remAmtKey != "" && totAmtKey != "" {
							remVal, okRem := parseNumericFraction(item.Get(remAmtKey))
							totVal, okTot := parseNumericFraction(item.Get(totAmtKey))
							if okRem && okTot && totVal > 0 {
								frac = remVal / totVal
								hasFraction = true
							}
						}
						if !hasFraction {
							continue
						}
						bucket := pluginapi.QuotaBucket{
							Window:            item.Get(winKey).String(),
							RemainingFraction: frac,
							ResetTime:         item.Get(resetKey).String(),
							Description:       item.Get(descKey).String(),
						}
						group.Buckets = append(group.Buckets, bucket)
					}
				}
			}

			if rawBuckets, okRB := gm["buckets"].([]any); okRB {
				for _, rb := range rawBuckets {
					bm, okBM := rb.(map[string]any)
					if !okBM {
						continue
					}
					var hasFraction bool
					var frac float64
					if rf, ok := bm["remaining_fraction"].(string); ok && rf != "" {
						if val, okNum := parseNumericFraction(gjson.GetBytes(respBytes, rf)); okNum {
							frac = val
							hasFraction = true
						}
					}
					if !hasFraction {
						if rem, okRem := bm["remaining_amount"].(string); okRem && rem != "" {
							if tot, okTot := bm["total_amount"].(string); okTot && tot != "" {
								remVal, okRemNum := parseNumericFraction(gjson.GetBytes(respBytes, rem))
								totVal, okTotNum := parseNumericFraction(gjson.GetBytes(respBytes, tot))
								if okRemNum && okTotNum && totVal > 0 {
									frac = remVal / totVal
									hasFraction = true
								}
							}
						}
					}
					if !hasFraction {
						continue
					}
					bucket := pluginapi.QuotaBucket{
						RemainingFraction: frac,
					}
					if win, ok := bm["window"].(string); ok {
						if res := gjson.GetBytes(respBytes, win); res.Exists() {
							bucket.Window = res.String()
						} else {
							bucket.Window = win
						}
					}
					if desc, ok := bm["description"].(string); ok {
						if res := gjson.GetBytes(respBytes, desc); res.Exists() {
							bucket.Description = res.String()
						} else {
							bucket.Description = desc
						}
					}
					if reset, ok := bm["reset_time"].(string); ok {
						if res := gjson.GetBytes(respBytes, reset); res.Exists() {
							bucket.ResetTime = res.String()
						} else {
							bucket.ResetTime = reset
						}
					}
					group.Buckets = append(group.Buckets, bucket)
				}
			}
			if len(group.Buckets) > 0 {
				out.Groups = append(out.Groups, group)
			}
		}
	}

	totalBuckets := 0
	for _, g := range out.Groups {
		totalBuckets += len(g.Buckets)
	}
	hasPlan := out.Subscription != nil && strings.TrimSpace(out.Subscription.Plan) != ""
	out.Summary = filterUsableQuotaSummary(respBytes)
	hasSummary := len(out.Summary) > 0
	if totalBuckets == 0 && !hasPlan && !hasSummary {
		return out, fmt.Errorf("response mapping did not match any valid quota fields in upstream response")
	}
	return out, nil
}
