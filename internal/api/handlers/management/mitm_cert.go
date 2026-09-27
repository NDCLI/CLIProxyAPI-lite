package management

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const mitmCAName = "Lumina Root CA"

var errUnsupportedCertStore = errors.New("automatic certificate installation is not supported on this operating system")

func (h *Handler) DownloadMITMCA(c *gin.Context) {
	certPath, _, errEnsure := h.ensureMITMCA()
	if errEnsure != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errEnsure.Error()})
		return
	}
	c.FileAttachment(certPath, "Lumina-Root-CA.crt")
}

func (h *Handler) InstallMITMCert(c *gin.Context) {
	certPath, _, errEnsure := h.ensureMITMCA()
	if errEnsure != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errEnsure.Error()})
		return
	}
	if errInstall := installMITMCertificate(certPath); errInstall != nil {
		status := http.StatusInternalServerError
		if errors.Is(errInstall, errUnsupportedCertStore) {
			status = http.StatusNotImplemented
		}
		c.JSON(status, gin.H{"error": errInstall.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Lumina Root CA installed in the current user's trusted root store"})
}

func (h *Handler) UninstallMITMCert(c *gin.Context) {
	certPath, _ := h.mitmCAPaths()
	cert, errLoad := loadMITMCertificate(certPath)
	if errors.Is(errLoad, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lumina Root CA has not been generated"})
		return
	}
	if errLoad != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errLoad.Error()})
		return
	}
	if errUninstall := uninstallMITMCertificate(cert); errUninstall != nil {
		status := http.StatusInternalServerError
		if errors.Is(errUninstall, errUnsupportedCertStore) {
			status = http.StatusNotImplemented
		}
		c.JSON(status, gin.H{"error": errUninstall.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Lumina Root CA removed from the current user's trusted root store"})
}

func (h *Handler) mitmCAPaths() (string, string) {
	baseDir := "."
	if h != nil && h.configFilePath != "" {
		baseDir = filepath.Dir(h.configFilePath)
	}
	certDir := filepath.Join(baseDir, "certs")
	return filepath.Join(certDir, "ca.crt"), filepath.Join(certDir, "ca.key")
}

func (h *Handler) ensureMITMCA() (string, string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	certPath, keyPath := h.mitmCAPaths()
	if _, errCert := loadMITMCertificate(certPath); errCert == nil {
		if _, errKey := os.Stat(keyPath); errKey == nil {
			return certPath, keyPath, nil
		}
	}

	if errDir := os.MkdirAll(filepath.Dir(certPath), 0o700); errDir != nil {
		return "", "", fmt.Errorf("create MITM certificate directory: %w", errDir)
	}

	privateKey, errKey := rsa.GenerateKey(rand.Reader, 2048)
	if errKey != nil {
		return "", "", fmt.Errorf("generate MITM CA private key: %w", errKey)
	}
	serial, errSerial := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if errSerial != nil {
		return "", "", fmt.Errorf("generate MITM CA serial: %w", errSerial)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   mitmCAName,
			Organization: []string{"Lumina"},
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        true,
	}
	certDER, errCreate := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if errCreate != nil {
		return "", "", fmt.Errorf("create MITM CA certificate: %w", errCreate)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if errWrite := os.WriteFile(keyPath, keyPEM, 0o600); errWrite != nil {
		return "", "", fmt.Errorf("write MITM CA private key: %w", errWrite)
	}
	if errWrite := os.WriteFile(certPath, certPEM, 0o644); errWrite != nil {
		return "", "", fmt.Errorf("write MITM CA certificate: %w", errWrite)
	}
	return certPath, keyPath, nil
}

func loadMITMCertificate(path string) (*x509.Certificate, error) {
	data, errRead := os.ReadFile(path)
	if errRead != nil {
		return nil, errRead
	}
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid MITM CA certificate: %s", path)
	}
	cert, errParse := x509.ParseCertificate(block.Bytes)
	if errParse != nil {
		return nil, fmt.Errorf("parse MITM CA certificate: %w", errParse)
	}
	return cert, nil
}

func installMITMCertificate(certPath string) error {
	if runtime.GOOS != "windows" {
		return errUnsupportedCertStore
	}
	output, errRun := hiddenMITMCommand("certutil", "-addstore", "-user", "Root", certPath).CombinedOutput()
	if errRun != nil {
		return fmt.Errorf("install MITM CA: %s: %w", strings.TrimSpace(string(output)), errRun)
	}
	return nil
}

func uninstallMITMCertificate(cert *x509.Certificate) error {
	if runtime.GOOS != "windows" {
		return errUnsupportedCertStore
	}
	fingerprint := sha1.Sum(cert.Raw)
	thumbprint := strings.ToUpper(hex.EncodeToString(fingerprint[:]))
	output, errRun := hiddenMITMCommand("certutil", "-delstore", "-user", "Root", thumbprint).CombinedOutput()
	if errRun != nil {
		return fmt.Errorf("uninstall MITM CA: %s: %w", strings.TrimSpace(string(output)), errRun)
	}
	return nil
}
