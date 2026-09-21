package management

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

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
