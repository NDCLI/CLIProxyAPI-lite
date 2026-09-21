package management

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestEnsureMITMCAWritesAndReusesCertificate(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	h := NewHandler(&config.Config{}, configPath, nil)

	certPath, keyPath, errEnsure := h.ensureMITMCA()
	if errEnsure != nil {
		t.Fatalf("ensureMITMCA failed: %v", errEnsure)
	}
	cert, errLoad := loadMITMCertificate(certPath)
	if errLoad != nil {
		t.Fatalf("loadMITMCertificate failed: %v", errLoad)
	}
	if !cert.IsCA || cert.Subject.CommonName != mitmCAName {
		t.Fatalf("unexpected certificate: IsCA=%v CN=%q", cert.IsCA, cert.Subject.CommonName)
	}
	firstKey, errRead := os.ReadFile(keyPath)
	if errRead != nil {
		t.Fatalf("read private key: %v", errRead)
	}

	_, _, errEnsure = h.ensureMITMCA()
	if errEnsure != nil {
		t.Fatalf("second ensureMITMCA failed: %v", errEnsure)
	}
	secondKey, errRead := os.ReadFile(keyPath)
	if errRead != nil {
		t.Fatalf("read reused private key: %v", errRead)
	}
	if string(firstKey) != string(secondKey) {
		t.Fatal("ensureMITMCA replaced an existing CA")
	}
}
