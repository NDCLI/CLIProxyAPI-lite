package management

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	_ "github.com/router-for-me/CLIProxyAPI/v7/internal/translator"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
)

func TestMITMMatchingAndModelExtraction(t *testing.T) {
	if got := mitmToolForHost("api.individual.githubcopilot.com"); got != "copilot" {
		t.Fatalf("tool = %q, want copilot", got)
	}
	if !matchesMITMPath("antigravity", "/v1/models/gemini-default:streamGenerateContent") {
		t.Fatal("expected Antigravity streaming path to match")
	}
	if got := extractMITMModel("/v1/models/gemini-default:streamGenerateContent", nil); got != "gemini-default" {
		t.Fatalf("URL model = %q", got)
	}
	if got := extractMITMModel("/chat/completions", []byte(`{"model":"gpt-4o"}`)); got != "gpt-4o" {
		t.Fatalf("body model = %q", got)
	}
}

func TestMITMStatusReportsLastObservedModel(t *testing.T) {
	runtime := newMITMRuntime(t.TempDir(), 8317)
	runtime.lastTool = "antigravity"
	runtime.lastModel = "gemini-3.8-flash-tiered"
	runtime.lastMapped = "xpiki/gpt-6-luna"
	status := runtime.status("missing-cert.crt")
	if status.LastTool != runtime.lastTool || status.LastModel != runtime.lastModel || status.LastMapped != runtime.lastMapped {
		t.Fatalf("last observed route missing from status: %+v", status)
	}
}

func TestCleanupMITMDNSHostsContentOnlyRemovesOwnedRedirects(t *testing.T) {
	content := strings.Join([]string{
		"127.0.0.1 daily-cloudcode-pa.googleapis.com cloudcode-pa.googleapis.com # CLIProxyAPI-lite MITM",
		"127.0.0.1 example.local # user entry",
		"127.0.0.1 api.individual.githubcopilot.com # another tool",
		"",
	}, "\r\n")
	updated, changed, errCleanup := cleanupMITMDNSHostsContent(content)
	if errCleanup != nil || !changed {
		t.Fatalf("cleanup = changed %t, error %v; want removed managed entries", changed, errCleanup)
	}
	if strings.Contains(updated, "# CLIProxyAPI-lite MITM") || strings.Contains(updated, "daily-cloudcode-pa.googleapis.com") || strings.Contains(updated, "cloudcode-pa.googleapis.com") {
		t.Fatalf("managed redirect remains after cleanup: %q", updated)
	}
	for _, entry := range []string{"127.0.0.1 example.local # user entry", "127.0.0.1 api.individual.githubcopilot.com # another tool"} {
		if !strings.Contains(updated, entry) {
			t.Fatalf("cleanup removed unrelated entry %q from %q", entry, updated)
		}
	}
}

func TestMITMDNSPreferencesPersistAcrossRuntimeInstances(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	first := newMITMRuntime(configPath, 8317)
	if errSet := first.setDNSPreference("antigravity", true); errSet != nil {
		t.Fatal(errSet)
	}

	second := newMITMRuntime(configPath, 8317)
	if !second.dnsDesired["antigravity"] {
		t.Fatal("Antigravity DNS preference was not restored")
	}
	if errSet := second.setDNSPreference("antigravity", false); errSet != nil {
		t.Fatal(errSet)
	}
	third := newMITMRuntime(configPath, 8317)
	if third.dnsDesired["antigravity"] {
		t.Fatal("disabled DNS preference was restored")
	}
}

func TestMITMOpenAIResponseTranslatesBackToGemini(t *testing.T) {
	response := []byte(`{"id":"chatcmpl-1","model":"test","choices":[{"index":0,"message":{"role":"assistant","content":"xin chao"},"finish_reason":"stop"}]}`)
	converted := sdktranslator.TranslateNonStream(
		context.Background(),
		sdktranslator.FormatOpenAI,
		sdktranslator.FormatGemini,
		"test",
		[]byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`),
		nil,
		response,
		nil,
	)
	if !bytes.Contains(converted, []byte(`"candidates"`)) || !bytes.Contains(converted, []byte("xin chao")) {
		t.Fatalf("unexpected Gemini response: %s", converted)
	}
}

func TestMITMAntigravityResponseEnvelope(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString(`{"id":"chatcmpl-1","choices":[{"index":0,"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}`))}
	recorder := httptest.NewRecorder()
	if errProxy := proxyAntigravityResponse(recorder, response, "test", []byte(`{"request":{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}}`), nil, false); errProxy != nil {
		t.Fatal(errProxy)
	}
	var result struct {
		Response struct {
			Candidates []json.RawMessage `json:"candidates"`
		} `json:"response"`
	}
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &result); errDecode != nil || len(result.Response.Candidates) == 0 {
		t.Fatalf("missing Antigravity response envelope: %s, error=%v", recorder.Body.String(), errDecode)
	}
	stream := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBufferString("data: {\"id\":\"chatcmpl-1\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"OK\"},\"finish_reason\":\"stop\"}]}\n\n"))}
	recorder = httptest.NewRecorder()
	if errProxy := proxyAntigravityResponse(recorder, stream, "test", nil, nil, true); errProxy != nil {
		t.Fatal(errProxy)
	}
	for _, line := range bytes.Split(recorder.Body.Bytes(), []byte("\n")) {
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}
		if errDecode := json.Unmarshal(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data: "))), &result); errDecode != nil || len(result.Response.Candidates) == 0 {
			t.Fatalf("missing streaming Antigravity response envelope: %s, error=%v", line, errDecode)
		}
	}
}

func TestMITMHTTPClientPreservesHTTP2Redirect(t *testing.T) {
	listener, errListen := net.Listen("tcp4", "127.0.0.1:0")
	if errListen != nil {
		t.Fatal(errListen)
	}
	server := &httptest.Server{Listener: listener, Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
	})}, EnableHTTP2: true}
	server.StartTLS()
	defer server.Close()
	client := directMITMClient()
	transport := client.Transport.(*http.Transport)
	transport.DialContext = (&net.Dialer{}).DialContext
	transport.TLSClientConfig = server.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	defer client.CloseIdleConnections()
	response, errRequest := client.Get(server.URL)
	if errRequest != nil {
		t.Fatal(errRequest)
	}
	defer response.Body.Close()
	if response.ProtoMajor != 2 {
		t.Fatalf("protocol = %q, want HTTP/2", response.Proto)
	}
	if response.StatusCode != http.StatusFound || response.Header.Get("Location") != "/login" {
		t.Fatalf("redirect was not preserved: status=%d location=%q", response.StatusCode, response.Header.Get("Location"))
	}
}

func TestResolveMITMUpstreamBypassesHosts(t *testing.T) {
	mitmDNSCache.Delete("localhost")
	mitmDNSCache.Delete("blocked.test")
	defer mitmDNSCache.Delete("localhost")
	defer mitmDNSCache.Delete("blocked.test")
	packet, errListen := net.ListenPacket("udp4", "127.0.0.1:0")
	if errListen != nil {
		t.Fatal(errListen)
	}
	server := &dns.Server{PacketConn: packet, Handler: dns.HandlerFunc(func(w dns.ResponseWriter, request *dns.Msg) {
		response := new(dns.Msg)
		response.SetReply(request)
		ip := net.IPv4(8, 8, 4, 4)
		if request.Question[0].Name == "blocked.test." {
			ip = net.IPv4(127, 0, 0, 1)
		}
		response.Answer = []dns.RR{&dns.A{Hdr: dns.RR_Header{Name: request.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET}, A: ip}}
		_ = w.WriteMsg(response)
	})}
	go server.ActivateAndServe()
	defer server.Shutdown()

	got, errResolve := resolveMITMUpstreamWithServers(context.Background(), "localhost", []string{packet.LocalAddr().String()})
	if errResolve != nil || got != "8.8.4.4" {
		t.Fatalf("localhost resolved through hosts instead of DNS: ip=%q err=%v", got, errResolve)
	}
	if _, errResolve := resolveMITMUpstreamWithServers(context.Background(), "blocked.test", []string{packet.LocalAddr().String()}); errResolve == nil {
		t.Fatal("loopback DNS answer must not route back into MITM")
	}
}

func TestGenerateMITMLeafCertificateUsesRequestedHost(t *testing.T) {
	rootKey, errKey := rsa.GenerateKey(rand.Reader, 2048)
	if errKey != nil {
		t.Fatal(errKey)
	}
	root := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: mitmCAName},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	rootDER, errCreate := x509.CreateCertificate(rand.Reader, root, root, &rootKey.PublicKey, rootKey)
	if errCreate != nil {
		t.Fatal(errCreate)
	}
	root, errCreate = x509.ParseCertificate(rootDER)
	if errCreate != nil {
		t.Fatal(errCreate)
	}
	pair, errLeaf := generateMITMLeafCertificate("api.individual.githubcopilot.com", root, rootKey)
	if errLeaf != nil {
		t.Fatal(errLeaf)
	}
	leaf, errParse := x509.ParseCertificate(pair.Certificate[0])
	if errParse != nil {
		t.Fatal(errParse)
	}
	if errVerify := leaf.VerifyHostname("api.individual.githubcopilot.com"); errVerify != nil {
		t.Fatalf("leaf certificate hostname mismatch: %v", errVerify)
	}
}
