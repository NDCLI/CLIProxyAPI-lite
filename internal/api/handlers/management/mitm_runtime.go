package management

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
)

const mitmListenAddress = "127.0.0.1:443"

var mitmToolHosts = map[string][]string{
	"antigravity": {"daily-cloudcode-pa.googleapis.com", "cloudcode-pa.googleapis.com"},
	"copilot":     {"api.individual.githubcopilot.com"},
	"kiro":        {"runtime.us-east-1.kiro.dev", "q.us-east-1.amazonaws.com", "codewhisperer.us-east-1.amazonaws.com"},
}

var mitmURLPatterns = map[string][]string{
	"antigravity": {":generateContent", ":streamGenerateContent"},
	"copilot":     {"/chat/completions", "/v1/messages", "/responses"},
	"kiro":        {"/generateAssistantResponse"},
}

type mitmRuntime struct {
	mu          sync.RWMutex
	server      *http.Server
	listener    net.Listener
	gatewayPort int
	baseDir     string
	apiKey      string
	baseURL     string
	apiKeyID    string
	mappings    map[string]map[string]string
	leafCerts   map[string]*tls.Certificate
	rootCert    *x509.Certificate
	rootKey     *rsa.PrivateKey
}

type mitmStatus struct {
	Running          bool            `json:"running"`
	Address          string          `json:"address"`
	CertExists       bool            `json:"cert_exists"`
	CertTrusted      bool            `json:"cert_trusted"`
	IsAdmin          bool            `json:"is_admin"`
	DNS              map[string]bool `json:"dns"`
	BaseURL          string          `json:"base_url"`
	APIKeyID         string          `json:"api_key_id,omitempty"`
	APIKeyConfigured bool            `json:"api_key_configured"`
	Tools            []mitmToolInfo  `json:"tools"`
}

func newMITMRuntime(configFilePath string, gatewayPort int) *mitmRuntime {
	baseDir := "."
	if configFilePath != "" {
		baseDir = filepath.Dir(configFilePath)
	}
	runtime := &mitmRuntime{
		gatewayPort: gatewayPort,
		baseDir:     baseDir,
		mappings:    make(map[string]map[string]string),
		leafCerts:   make(map[string]*tls.Certificate),
	}
	runtime.loadMappings()
	return runtime
}

func (m *mitmRuntime) mappingsPath() string {
	return filepath.Join(m.baseDir, "certs", "mitm-mappings.json")
}

func (m *mitmRuntime) loadMappings() {
	data, errRead := os.ReadFile(m.mappingsPath())
	if errRead == nil {
		var mappings map[string]map[string]string
		if json.Unmarshal(data, &mappings) == nil && mappings != nil {
			m.mappings = mappings
		}
	}
}

func (m *mitmRuntime) saveMappingsLocked() error {
	if errDir := os.MkdirAll(filepath.Dir(m.mappingsPath()), 0o700); errDir != nil {
		return errDir
	}
	data, errMarshal := json.MarshalIndent(m.mappings, "", "  ")
	if errMarshal != nil {
		return errMarshal
	}
	tmp, errCreate := os.CreateTemp(filepath.Dir(m.mappingsPath()), ".mitm-mappings-*")
	if errCreate != nil {
		return errCreate
	}
	defer os.Remove(tmp.Name())
	if _, errWrite := tmp.Write(data); errWrite != nil {
		_ = tmp.Close()
		return errWrite
	}
	if errClose := tmp.Close(); errClose != nil {
		return errClose
	}
	return os.Rename(tmp.Name(), m.mappingsPath())
}

func (m *mitmRuntime) status(certPath string) mitmStatus {
	m.mu.RLock()
	running := m.server != nil && m.listener != nil
	baseURL, apiKeyID, hasKey := m.gatewayBaseURLLocked(), m.apiKeyID, m.apiKey != ""
	m.mu.RUnlock()
	return mitmStatus{
		Running:          running,
		Address:          mitmListenAddress,
		CertExists:       fileExists(certPath),
		CertTrusted:      certificateTrusted(certPath),
		IsAdmin:          currentProcessIsAdmin(),
		DNS:              mitmDNSStatus(),
		BaseURL:          baseURL,
		APIKeyID:         apiKeyID,
		APIKeyConfigured: hasKey,
		Tools:            mitmTools(),
	}
}

func (m *mitmRuntime) start(certPath, keyPath, apiKey, baseURL, apiKeyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.server != nil {
		return fmt.Errorf("MITM server is already running; stop it before changing settings")
	}

	rootCert, rootKey, errLoad := loadMITMCAKeyPair(certPath, keyPath)
	if errLoad != nil {
		return errLoad
	}
	listener, errListen := net.Listen("tcp", mitmListenAddress)
	if errListen != nil {
		return fmt.Errorf("MITM port 443 unavailable: %w", errListen)
	}
	m.rootCert = rootCert
	m.rootKey = rootKey
	m.apiKey = apiKey
	m.baseURL = baseURL
	m.apiKeyID = apiKeyID
	m.leafCerts = make(map[string]*tls.Certificate)
	tlsListener := tls.NewListener(listener, &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: m.getCertificate,
	})
	m.listener = listener
	m.server = &http.Server{
		Handler:           http.HandlerFunc(m.serveHTTP),
		ReadHeaderTimeout: 10 * time.Second,
	}
	server := m.server
	go func() {
		if errServe := server.Serve(tlsListener); errServe != nil && !errors.Is(errServe, http.ErrServerClosed) {
			m.mu.Lock()
			if m.server == server {
				m.server = nil
				m.listener = nil
			}
			m.mu.Unlock()
		}
	}()
	return nil
}

func (m *mitmRuntime) stop(ctx context.Context) error {
	m.mu.Lock()
	server := m.server
	m.server = nil
	m.listener = nil
	m.apiKey = ""
	m.apiKeyID = ""
	m.mu.Unlock()
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}

func (m *mitmRuntime) getCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	host := strings.ToLower(strings.TrimSpace(hello.ServerName))
	if host == "" || mitmToolForHost(host) == "" {
		return nil, fmt.Errorf("unsupported MITM host %q", host)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if cert := m.leafCerts[host]; cert != nil {
		return cert, nil
	}
	leaf, errLeaf := generateMITMLeafCertificate(host, m.rootCert, m.rootKey)
	if errLeaf != nil {
		return nil, errLeaf
	}
	m.leafCerts[host] = leaf
	return leaf, nil
}

func (m *mitmRuntime) serveHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/_mitm_health" {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
		return
	}
	host := strings.ToLower(strings.Split(req.Host, ":")[0])
	tool := mitmToolForHost(host)
	if tool == "" {
		http.Error(w, "unsupported MITM host", http.StatusBadRequest)
		return
	}
	body, errRead := io.ReadAll(http.MaxBytesReader(w, req.Body, 32<<20))
	if errRead != nil {
		http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	if !matchesMITMRequest(tool, req) {
		m.passthrough(w, req, host, body)
		return
	}
	model := extractMITMModel(req.URL.Path, body)
	mapped := m.mappedModel(tool, model)
	if mapped == "" {
		m.passthrough(w, req, host, body)
		return
	}
	if errProxy := m.proxyMapped(w, req, tool, mapped, body); errProxy != nil {
		http.Error(w, errProxy.Error(), http.StatusBadGateway)
	}
}

func (m *mitmRuntime) mappedModel(tool, model string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	model = strings.TrimPrefix(model, "models/")
	if strings.HasPrefix(model, "tab_") || strings.HasPrefix(model, "tab-") {
		return ""
	}
	if exact := strings.TrimSpace(m.mappings[tool][model]); exact != "" {
		return exact
	}
	if tool == "antigravity" {
		if alias := mitmAntigravityAliases[model]; alias != "" {
			model = alias
		}
	}
	return strings.TrimSpace(m.mappings[tool][model])
}

func (m *mitmRuntime) proxyMapped(w http.ResponseWriter, req *http.Request, tool, mapped string, original []byte) error {
	path := "/v1/chat/completions"
	body := original
	stream := strings.Contains(req.URL.Path, ":streamGenerateContent")
	if tool == "copilot" {
		var payload map[string]any
		if errDecode := json.Unmarshal(body, &payload); errDecode != nil {
			return fmt.Errorf("decode Copilot request: %w", errDecode)
		}
		payload["model"] = mapped
		body, _ = json.Marshal(payload)
		if strings.Contains(req.URL.Path, "/v1/messages") {
			path = "/v1/messages"
		} else if strings.Contains(req.URL.Path, "/responses") {
			path = "/v1/responses"
		}
	} else if tool == "kiro" {
		var errConvert error
		body, errConvert = translateKiroMITMRequest(original, mapped)
		if errConvert != nil {
			return errConvert
		}
	} else {
		var envelope map[string]json.RawMessage
		if errDecode := json.Unmarshal(body, &envelope); errDecode != nil {
			return fmt.Errorf("decode Antigravity request: %w", errDecode)
		}
		geminiBody := envelope["request"]
		if len(geminiBody) == 0 {
			geminiBody = body
		}
		body = sdktranslator.TranslateRequest(sdktranslator.FormatGemini, sdktranslator.FormatOpenAI, mapped, geminiBody, stream)
	}

	m.mu.RLock()
	gatewayURL := m.gatewayBaseURLLocked() + path
	apiKey := m.apiKey
	m.mu.RUnlock()
	proxyReq, errRequest := http.NewRequestWithContext(req.Context(), http.MethodPost, gatewayURL, bytes.NewReader(body))
	if errRequest != nil {
		return errRequest
	}
	copyMITMHeaders(proxyReq.Header, req.Header)
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	response, errDo := client.Do(proxyReq)
	if errDo != nil {
		return errDo
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		copyResponse(w, response)
		return nil
	}
	if tool == "kiro" {
		return proxyKiroMITMResponse(w, response, mapped)
	}
	if tool == "copilot" {
		copyResponse(w, response)
		return nil
	}
	return proxyAntigravityResponse(w, response, mapped, original, body, stream)
}

func (m *mitmRuntime) gatewayBaseURLLocked() string {
	if m.baseURL != "" {
		return m.baseURL
	}
	return fmt.Sprintf("http://127.0.0.1:%d", m.gatewayPort)
}

func proxyAntigravityResponse(w http.ResponseWriter, response *http.Response, model string, original, translated []byte, stream bool) error {
	if !stream {
		data, errRead := io.ReadAll(response.Body)
		if errRead != nil {
			return errRead
		}
		converted := sdktranslator.TranslateNonStream(context.Background(), sdktranslator.FormatOpenAI, sdktranslator.FormatGemini, model, original, translated, data, nil)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.StatusCode)
		_, _ = w.Write(converted)
		return nil
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(response.StatusCode)
	flusher, _ := w.(http.Flusher)
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	var state any
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		for _, chunk := range sdktranslator.TranslateStream(context.Background(), sdktranslator.FormatOpenAI, sdktranslator.FormatGemini, model, original, translated, []byte(data), &state) {
			_, _ = fmt.Fprintf(w, "data: %s\r\n\r\n", chunk)
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
	return scanner.Err()
}

func (m *mitmRuntime) passthrough(w http.ResponseWriter, req *http.Request, host string, body []byte) {
	if host == "" {
		http.Error(w, "missing upstream host", http.StatusBadRequest)
		return
	}
	upstreamURL := &url.URL{Scheme: "https", Host: host, Path: req.URL.Path, RawQuery: req.URL.RawQuery}
	upstreamReq, errRequest := http.NewRequestWithContext(req.Context(), req.Method, upstreamURL.String(), bytes.NewReader(body))
	if errRequest != nil {
		http.Error(w, errRequest.Error(), http.StatusBadRequest)
		return
	}
	upstreamReq.Header = req.Header.Clone()
	response, errDo := directMITMClient().Do(upstreamReq)
	if errDo != nil {
		http.Error(w, errDo.Error(), http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	copyResponse(w, response)
}

func directMITMClient() *http.Client {
	resolver := &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "udp", "8.8.8.8:53")
	}}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, errSplit := net.SplitHostPort(address)
			if errSplit != nil {
				return nil, errSplit
			}
			addresses, errLookup := resolver.LookupHost(ctx, host)
			if errLookup != nil || len(addresses) == 0 {
				return nil, fmt.Errorf("resolve upstream %s: %w", host, errLookup)
			}
			return (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(addresses[0], port))
		},
	}
	return &http.Client{Transport: transport}
}

func copyMITMHeaders(dst, src http.Header) {
	for key, values := range src {
		switch strings.ToLower(key) {
		case "host", "content-length", "connection", "transfer-encoding", "authorization", "x-api-key", "cookie", "x-amz-security-token", "x-amz-target", "accept-encoding":
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func copyResponse(w http.ResponseWriter, response *http.Response) {
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func mitmToolForHost(host string) string {
	for tool, hosts := range mitmToolHosts {
		for _, candidate := range hosts {
			if strings.EqualFold(host, candidate) {
				return tool
			}
		}
	}
	return ""
}

func matchesMITMPath(tool, path string) bool {
	for _, pattern := range mitmURLPatterns[tool] {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func matchesMITMRequest(tool string, req *http.Request) bool {
	return matchesMITMPath(tool, req.URL.Path) || (tool == "kiro" && req.Method == http.MethodPost && strings.HasSuffix(req.Header.Get("X-Amz-Target"), ".GenerateAssistantResponse"))
}

func extractMITMModel(path string, body []byte) string {
	if marker := strings.Index(path, "/models/"); marker >= 0 {
		value := path[marker+len("/models/"):]
		if end := strings.IndexAny(value, ":/"); end >= 0 {
			value = value[:end]
		}
		return value
	}
	var payload struct {
		Model             string `json:"model"`
		ConversationState struct {
			CurrentMessage struct {
				UserInputMessage struct {
					ModelID string `json:"modelId"`
				} `json:"userInputMessage"`
			} `json:"currentMessage"`
		} `json:"conversationState"`
		Request          json.RawMessage `json:"request"`
		GenerationConfig struct {
			ThinkingConfig struct {
				ThinkingLevel string `json:"thinkingLevel"`
			} `json:"thinkingConfig"`
		} `json:"generationConfig"`
	}
	_ = json.Unmarshal(body, &payload)
	if payload.ConversationState.CurrentMessage.UserInputMessage.ModelID != "" {
		return payload.ConversationState.CurrentMessage.UserInputMessage.ModelID
	}
	if strings.HasSuffix(payload.Model, "-flash-tiered") {
		level := strings.ToLower(payload.GenerationConfig.ThinkingConfig.ThinkingLevel)
		if len(payload.Request) > 0 {
			var nested struct {
				GenerationConfig struct {
					ThinkingConfig struct {
						ThinkingLevel string `json:"thinkingLevel"`
					} `json:"thinkingConfig"`
				} `json:"generationConfig"`
			}
			_ = json.Unmarshal(payload.Request, &nested)
			if nested.GenerationConfig.ThinkingConfig.ThinkingLevel != "" {
				level = strings.ToLower(nested.GenerationConfig.ThinkingConfig.ThinkingLevel)
			}
		}
		if level != "high" && level != "low" {
			level = "medium"
		}
		return strings.TrimSuffix(payload.Model, "-tiered") + "-" + level
	}
	return payload.Model
}

func loadMITMCAKeyPair(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	certData, errCert := os.ReadFile(certPath)
	if errCert != nil {
		return nil, nil, errCert
	}
	keyData, errKey := os.ReadFile(keyPath)
	if errKey != nil {
		return nil, nil, errKey
	}
	certBlock, _ := pem.Decode(certData)
	keyBlock, _ := pem.Decode(keyData)
	if certBlock == nil || keyBlock == nil {
		return nil, nil, fmt.Errorf("invalid MITM CA key pair")
	}
	cert, errParseCert := x509.ParseCertificate(certBlock.Bytes)
	if errParseCert != nil {
		return nil, nil, errParseCert
	}
	key, errParseKey := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if errParseKey != nil {
		return nil, nil, errParseKey
	}
	return cert, key, nil
}

func generateMITMLeafCertificate(host string, root *x509.Certificate, rootKey *rsa.PrivateKey) (*tls.Certificate, error) {
	key, errKey := rsa.GenerateKey(rand.Reader, 2048)
	if errKey != nil {
		return nil, errKey
	}
	serial, errSerial := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if errSerial != nil {
		return nil, errSerial
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: host, Organization: []string{"CLIProxyAPI-lite"}},
		DNSNames:     []string{host},
		NotBefore:    now.Add(-5 * time.Minute),
		NotAfter:     now.AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, errCreate := x509.CreateCertificate(rand.Reader, template, root, &key.PublicKey, rootKey)
	if errCreate != nil {
		return nil, errCreate
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pair, errPair := tls.X509KeyPair(certPEM, keyPEM)
	return &pair, errPair
}

func fileExists(path string) bool {
	_, errStat := os.Stat(path)
	return errStat == nil
}

func currentProcessIsAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	return exec.Command("cmd", "/C", "net", "session").Run() == nil
}

func certificateTrusted(certPath string) bool {
	if runtime.GOOS != "windows" || !fileExists(certPath) {
		return false
	}
	cert, errLoad := loadMITMCertificate(certPath)
	if errLoad != nil {
		return false
	}
	fingerprint := sha1.Sum(cert.Raw)
	return exec.Command("certutil", "-user", "-store", "Root", hex.EncodeToString(fingerprint[:])).Run() == nil
}

func mitmDNSStatus() map[string]bool {
	status := make(map[string]bool, len(mitmToolHosts))
	data, errRead := os.ReadFile(windowsHostsPath())
	if errRead != nil {
		return status
	}
	return mitmDNSStatusFor(string(data))
}

func mitmDNSStatusFor(content string) map[string]bool {
	status := make(map[string]bool, len(mitmToolHosts))
	redirected := make(map[string]bool)
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(strings.SplitN(line, "#", 2)[0])
		if len(fields) < 2 || fields[0] != "127.0.0.1" {
			continue
		}
		for _, host := range fields[1:] {
			redirected[strings.ToLower(host)] = true
		}
	}
	for tool, hosts := range mitmToolHosts {
		status[tool] = true
		for _, host := range hosts {
			if !redirected[host] {
				status[tool] = false
				break
			}
		}
	}
	return status
}

func windowsHostsPath() string {
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = `C:\Windows`
	}
	return filepath.Join(root, "System32", "drivers", "etc", "hosts")
}

func setMITMDNS(tool string, enabled bool) error {
	hosts := mitmToolHosts[tool]
	if len(hosts) == 0 {
		return fmt.Errorf("unsupported MITM tool %q", tool)
	}
	if runtime.GOOS != "windows" {
		return fmt.Errorf("automatic DNS setup currently supports Windows only")
	}
	if !currentProcessIsAdmin() {
		return fmt.Errorf("administrator privileges required to update the Windows hosts file")
	}
	path := windowsHostsPath()
	data, errRead := os.ReadFile(path)
	if errRead != nil {
		return errRead
	}
	content, errUpdate := updateMITMHosts(string(data), hosts, enabled)
	if errUpdate != nil {
		return errUpdate
	}
	if errWrite := os.WriteFile(path, []byte(content), 0o644); errWrite != nil {
		return errWrite
	}
	_ = exec.Command("ipconfig", "/flushdns").Run()
	return nil
}

func updateMITMHosts(content string, hosts []string, enabled bool) (string, error) {
	wanted := make(map[string]bool, len(hosts))
	for _, host := range hosts {
		wanted[host] = true
	}
	var kept []string
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		parts := strings.SplitN(line, "#", 2)
		fields := strings.Fields(parts[0])
		owned := len(parts) == 2 && strings.TrimSpace(parts[1]) == "CLIProxyAPI-lite MITM"
		matching := false
		remaining := []string{}
		if len(fields) > 1 {
			for _, host := range fields[1:] {
				if wanted[strings.ToLower(host)] {
					matching = true
				} else {
					remaining = append(remaining, host)
				}
			}
		}
		if matching && !owned && enabled {
			return "", fmt.Errorf("hosts file already contains an unmanaged entry for %s; remove it manually before enabling DNS", strings.Join(hosts, ", "))
		}
		if matching && owned {
			if len(remaining) > 0 {
				kept = append(kept, fields[0]+" "+strings.Join(remaining, " ")+" # CLIProxyAPI-lite MITM")
			}
			continue
		}
		kept = append(kept, line)
	}
	if enabled {
		for _, host := range hosts {
			kept = append(kept, "127.0.0.1 "+host+" # CLIProxyAPI-lite MITM")
		}
	}
	return strings.TrimRight(strings.Join(kept, "\r\n"), "\r\n") + "\r\n", nil
}
