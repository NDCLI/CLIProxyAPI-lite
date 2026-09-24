package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func proxyPoolRequest(t *testing.T, handler func(*gin.Context), method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	for _, prefix := range []string{"/proxy-pools/", "/providers/"} {
		if index := strings.LastIndex(path, prefix); index >= 0 {
			id := path[index+len(prefix):]
			id = strings.TrimSuffix(id, "/test")
			ctx.Params = gin.Params{{Key: "id", Value: id}}
			break
		}
	}
	handler(ctx)
	return response
}

func TestProbeProxyPoolUsesProxyAndReportsHTTPStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var requestMethod, requestHost string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod = r.Method
		requestHost = r.URL.Host
		w.WriteHeader(http.StatusNoContent)
	}))
	defer proxy.Close()

	ok, status, _, reason := probeProxyPool(context.Background(), proxy.URL, "http://probe.example/")
	if !ok || status != http.StatusNoContent || reason != "" {
		t.Fatalf("probeProxyPool() = (%v, %d, %q), want success", ok, status, reason)
	}
	if requestMethod != http.MethodHead || requestHost != "probe.example" {
		t.Fatalf("proxy received %s %s, want HEAD probe.example", requestMethod, requestHost)
	}
}

func TestTestProxyPoolPersistsFailureWithoutDisablingPoolOrLeakingSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	now := time.Now().UTC()
	h := NewHandler(&config.Config{AuthDir: root}, filepath.Join(root, "config.yaml"), nil)
	pool := proxyPoolRecord{ID: "test-pool", Name: "Local", ProxyURL: "http://proxy-user:proxy-secret@127.0.0.1:1", IsActive: true, TestStatus: "unknown", CreatedAt: now, UpdatedAt: now}
	if errSave := h.proxyPools.save(pool); errSave != nil {
		t.Fatalf("save test pool: %v", errSave)
	}

	response := proxyPoolRequest(t, h.TestProxyPool, http.MethodPost, "/v0/management/proxy-pools/test-pool/test", "")
	if response.Code != http.StatusOK {
		t.Fatalf("test status = %d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "proxy-secret") || !strings.Contains(response.Body.String(), `"error_code":"proxy_unreachable"`) {
		t.Fatalf("unsafe or unexpected test response: %s", response.Body.String())
	}
	items, errList := h.proxyPools.list()
	if errList != nil || len(items) != 1 {
		t.Fatalf("pool list = %#v error=%v", items, errList)
	}
	if items[0].TestStatus != "error" || items[0].LastTestedAt == nil || !items[0].IsActive {
		t.Fatalf("test result = %+v, want recorded error without changing active state", items[0])
	}
}

func TestRecordProxyPoolTestResultPreservesConcurrentChanges(t *testing.T) {
	store := newProxyPoolStore(filepath.Join(t.TempDir(), "config.yaml"))
	now := time.Now().UTC()
	original := proxyPoolRecord{ID: "pool", Name: "before", ProxyURL: "http://proxy.example:8080", IsActive: true, CreatedAt: now, UpdatedAt: now}
	if errSave := store.save(original); errSave != nil {
		t.Fatalf("save initial pool: %v", errSave)
	}
	changed := original
	changed.Name = "renamed"
	changed.IsActive = false
	changed.UpdatedAt = now.Add(time.Minute)
	if errSave := store.save(changed); errSave != nil {
		t.Fatalf("save concurrent update: %v", errSave)
	}

	testedAt := now.Add(2 * time.Minute)
	found, unchanged, errRecord := store.recordTestResult(original.ID, original.ProxyURL, "active", 14, testedAt)
	if errRecord != nil || !found || !unchanged {
		t.Fatalf("recordTestResult() = (%v, %v, %v), want (true, true, nil)", found, unchanged, errRecord)
	}
	items, errList := store.list()
	if errList != nil || len(items) != 1 {
		t.Fatalf("pool list = %#v error=%v", items, errList)
	}
	if items[0].Name != "renamed" || items[0].IsActive || items[0].TestStatus != "active" || items[0].TestLatencyMS != 14 {
		t.Fatalf("recorded pool = %+v, want concurrent fields and test result preserved", items[0])
	}

	found, unchanged, errRecord = store.recordTestResult(original.ID, "http://new-proxy.example:8080", "active", 1, testedAt)
	if errRecord != nil || !found || unchanged {
		t.Fatalf("stale result = (%v, %v, %v), want (true, false, nil)", found, unchanged, errRecord)
	}
	items, errList = store.list()
	if errList != nil || items[0].ProxyURL != original.ProxyURL || items[0].TestLatencyMS != 14 {
		t.Fatalf("stale test mutated pool: items=%+v error=%v", items, errList)
	}
}

func TestProxyPoolAssignmentUpdatesRuntimeAndKeepsSecretsMasked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := coreauth.NewManager(&memoryAuthStore{}, nil, nil)
	auth := &coreauth.Auth{
		ID: "account.json", FileName: "account.json", Provider: "claude", ProxyURL: "http://old.proxy:8080",
		Metadata: map[string]any{"type": "claude", "proxy_url": "http://old.proxy:8080"},
	}
	if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
		t.Fatalf("register auth: %v", errRegister)
	}
	h := NewHandler(&config.Config{AuthDir: t.TempDir()}, filepath.Join(t.TempDir(), "config.yaml"), manager)

	create := proxyPoolRequest(t, h.PostProxyPool, http.MethodPost, "/v0/management/proxy-pools", `{"name":"Office","proxy_url":"http://proxy-user:proxy-pass@localhost:3128"}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		Item proxyPoolItem `json:"item"`
	}
	if errDecode := json.Unmarshal(create.Body.Bytes(), &created); errDecode != nil {
		t.Fatalf("decode create response: %v", errDecode)
	}
	if strings.Contains(create.Body.String(), "proxy-pass") || created.Item.ProxyURLMasked != "http://localhost:3128" {
		t.Fatalf("create response exposed proxy credentials: %s", create.Body.String())
	}

	assign := proxyPoolRequest(t, h.PatchProvider, http.MethodPatch, "/v0/management/providers/account.json", `{"proxy_pool_id":"`+created.Item.ID+`"}`)
	if assign.Code != http.StatusOK {
		t.Fatalf("assign status = %d body=%s", assign.Code, assign.Body.String())
	}
	current, ok := manager.GetByID(auth.ID)
	if !ok || current.ProxyURL != "http://proxy-user:proxy-pass@localhost:3128" {
		t.Fatalf("assigned proxy = %#v, want pool URL", current)
	}
	if current.Metadata["proxy_pool_id"] != created.Item.ID || current.Metadata["proxy_pool_original_url"] != "http://old.proxy:8080" {
		t.Fatalf("pool metadata = %#v", current.Metadata)
	}

	list := proxyPoolRequest(t, h.GetProxyPools, http.MethodGet, "/v0/management/proxy-pools", "")
	if list.Code != http.StatusOK || strings.Contains(list.Body.String(), "proxy-pass") {
		t.Fatalf("list status/body = %d %s", list.Code, list.Body.String())
	}

	disable := proxyPoolRequest(t, h.PatchProxyPool, http.MethodPatch, "/v0/management/proxy-pools/"+created.Item.ID, `{"is_active":false}`)
	if disable.Code != http.StatusOK {
		t.Fatalf("disable status = %d body=%s", disable.Code, disable.Body.String())
	}
	current, _ = manager.GetByID(auth.ID)
	if current.ProxyURL != "http://old.proxy:8080" {
		t.Fatalf("disabled pool proxy = %q, want restored original", current.ProxyURL)
	}

	removeWhileBound := proxyPoolRequest(t, h.DeleteProxyPool, http.MethodDelete, "/v0/management/proxy-pools/"+created.Item.ID, "")
	if removeWhileBound.Code != http.StatusConflict {
		t.Fatalf("delete bound pool status = %d, want %d", removeWhileBound.Code, http.StatusConflict)
	}

	unassign := proxyPoolRequest(t, h.PatchProvider, http.MethodPatch, "/v0/management/providers/account.json", `{"proxy_pool_id":""}`)
	if unassign.Code != http.StatusOK {
		t.Fatalf("unassign status = %d body=%s", unassign.Code, unassign.Body.String())
	}
	current, _ = manager.GetByID(auth.ID)
	if current.ProxyURL != "http://old.proxy:8080" {
		t.Fatalf("unassigned proxy = %q, want restored original", current.ProxyURL)
	}
	if proxyPoolIDForAuth(current) != "" {
		t.Fatalf("pool assignment remains in metadata: %#v", current.Metadata)
	}

	remove := proxyPoolRequest(t, h.DeleteProxyPool, http.MethodDelete, "/v0/management/proxy-pools/"+created.Item.ID, "")
	if remove.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", remove.Code, remove.Body.String())
	}
}

func TestValidateProxyPoolURLRejectsDirectAndUnsupportedSchemes(t *testing.T) {
	for _, raw := range []string{"", "direct", "ftp://proxy.local:21", "localhost:3128"} {
		if _, errValidate := validateProxyPoolURL(raw); errValidate == nil {
			t.Errorf("validateProxyPoolURL(%q) succeeded, want error", raw)
		}
	}
	if got, errValidate := validateProxyPoolURL("socks5://proxy.local:1080"); errValidate != nil || got != "socks5://proxy.local:1080" {
		t.Fatalf("valid proxy = %q error=%v", got, errValidate)
	}
}
