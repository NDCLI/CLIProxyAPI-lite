// Package tailnet exposes the proxy HTTP handler on a Tailscale network using an
// embedded tsnet node, optionally publishing it to the internet via Tailscale Funnel.
package tailnet

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	log "github.com/sirupsen/logrus"
	"tailscale.com/tsnet"
)

// DefaultHostname is used when no hostname is configured.
const DefaultHostname = "cliproxyapi"

// Status describes the current state of the embedded Tailscale node.
type Status struct {
	Enabled      bool     `json:"enabled"`
	Running      bool     `json:"running"`
	State        string   `json:"state"`
	Hostname     string   `json:"hostname"`
	Funnel       bool     `json:"funnel"`
	AuthURL      string   `json:"auth_url,omitempty"`
	IPs          []string `json:"ips,omitempty"`
	DNSName      string   `json:"dns_name,omitempty"`
	URLs         []string `json:"urls,omitempty"`
	Error        string   `json:"error,omitempty"`
	AuthKeyGiven bool     `json:"auth_key_set"`
}

type listenerEndpoint struct {
	Scheme string
	Port   string
}

// Manager owns the lifecycle of a single embedded tsnet node.
type Manager struct {
	mu        sync.Mutex
	handler   http.Handler
	baseDir   string
	cfg       config.TailscaleConfig
	srv       *tsnet.Server
	httpSrvs  []*http.Server
	cancel    context.CancelFunc
	lastError string
	urls      []string
	dnsName   string
	ips       []string
	listeners []listenerEndpoint
}

var defaultManager = &Manager{}

// Default returns the process-wide manager.
func Default() *Manager { return defaultManager }

// SetHandler sets the HTTP handler served on the tailnet and the base directory
// used to resolve the default state directory.
func (m *Manager) SetHandler(h http.Handler, baseDir string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handler = h
	m.baseDir = baseDir
}

// Start brings the node up with the given configuration. It returns once the
// node has been started; tailnet login and listener setup continue in background.
func (m *Manager) Start(cfg config.TailscaleConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.srv != nil {
		return errors.New("tailscale is already running")
	}
	if m.handler == nil {
		return errors.New("tailscale: http handler not configured")
	}
	hostname := strings.TrimSpace(cfg.Hostname)
	if hostname == "" {
		hostname = DefaultHostname
	}
	cfg.Hostname = hostname
	dir := strings.TrimSpace(cfg.StateDir)
	if dir == "" {
		dir = filepath.Join(m.baseDir, "tsnet-state")
	}
	// On Windows, if the folder is named "tailscale", tsnet attempts to set restrictive Administrator ACLs
	// via SetNamedSecurityInfo, which fails with "Access is denied" when running without full elevation.
	// Ensure directory name does not end with "tailscale".
	if strings.EqualFold(filepath.Base(dir), "tailscale") {
		dir = filepath.Join(dir, "state")
	}
	if errMkdir := os.MkdirAll(dir, 0o755); errMkdir != nil {
		return fmt.Errorf("tailscale: create state dir: %w", errMkdir)
	}
	authKey := strings.TrimSpace(cfg.AuthKey)
	if authKey == "" {
		authKey = strings.TrimSpace(os.Getenv("TS_AUTHKEY"))
	}

	srv := &tsnet.Server{
		Hostname:  hostname,
		Dir:       dir,
		AuthKey:   authKey,
		Ephemeral: cfg.Ephemeral,
		Logf:      func(format string, args ...any) { log.Tracef("tsnet: "+format, args...) },
		UserLogf:  func(format string, args ...any) { log.Infof("tailscale: "+format, args...) },
	}
	if errStart := srv.Start(); errStart != nil {
		return fmt.Errorf("tailscale: start node: %w", errStart)
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.srv = srv
	m.cfg = cfg
	m.cancel = cancel
	m.lastError = ""
	m.urls = nil
	go m.bringUp(ctx, srv, cfg, m.handler)
	return nil
}

// bringUp waits for tailnet login and then attaches the HTTP listeners.
func (m *Manager) bringUp(ctx context.Context, srv *tsnet.Server, cfg config.TailscaleConfig, handler http.Handler) {
	if _, errUp := srv.Up(ctx); errUp != nil {
		if !errors.Is(errUp, context.Canceled) {
			m.setError(fmt.Sprintf("tailscale up failed: %v", errUp))
		}
		return
	}

	type listenSpec struct {
		ln     net.Listener
		scheme string
		port   string
	}
	var specs []listenSpec
	domain := ""
	if domains := srv.CertDomains(); len(domains) > 0 {
		domain = domains[0]
	}
	var detectedIPs []string
	if lc, errLC := srv.LocalClient(); errLC == nil {
		if ipn, errStatus := lc.StatusWithoutPeers(ctx); errStatus == nil {
			if ipn.Self != nil && domain == "" {
				domain = strings.TrimSuffix(ipn.Self.DNSName, ".")
			}
			for _, ip := range ipn.TailscaleIPs {
				detectedIPs = append(detectedIPs, ip.String())
			}
		}
	}

	lnHTTP, errHTTP := srv.Listen("tcp", ":80")
	if errHTTP != nil {
		m.setError(fmt.Sprintf("tailscale listen :80 failed: %v", errHTTP))
		return
	}
	specs = append(specs, listenSpec{ln: lnHTTP, scheme: "http", port: "80"})

	if cfg.Funnel {
		lnFunnel, errFunnel := srv.ListenFunnel("tcp", ":443")
		if errFunnel != nil {
			m.setError(fmt.Sprintf("tailscale funnel failed (enable HTTPS + Funnel in admin console): %v", errFunnel))
		} else {
			specs = append(specs, listenSpec{ln: lnFunnel, scheme: "https", port: "443"})
		}
	} else if cfg.HTTPS && domain != "" {
		lnTLS, errTLS := srv.ListenTLS("tcp", ":443")
		if errTLS != nil {
			m.setError(fmt.Sprintf("tailscale https failed (enable HTTPS in admin console): %v", errTLS))
		} else {
			specs = append(specs, listenSpec{ln: lnTLS, scheme: "https", port: "443"})
		}
	}

	m.mu.Lock()
	if m.srv != srv {
		m.mu.Unlock()
		for _, spec := range specs {
			_ = spec.ln.Close()
		}
		return
	}
	m.dnsName = domain
	m.ips = detectedIPs
	m.listeners = nil
	for _, spec := range specs {
		m.listeners = append(m.listeners, listenerEndpoint{Scheme: spec.scheme, Port: spec.port})
		httpSrv := &http.Server{Handler: handler}
		m.httpSrvs = append(m.httpSrvs, httpSrv)
		host := domain
		if host == "" {
			host = cfg.Hostname
		}
		m.urls = append(m.urls, fmt.Sprintf("%s://%s", spec.scheme, host))
		go func(s *http.Server, ln net.Listener) {
			if errServe := s.Serve(ln); errServe != nil && !errors.Is(errServe, http.ErrServerClosed) && !errors.Is(errServe, net.ErrClosed) {
				log.Warnf("tailscale: http serve error: %v", errServe)
			}
		}(httpSrv, spec.ln)
	}
	urls := append([]string(nil), m.urls...)
	m.mu.Unlock()
	log.Infof("tailscale: proxy available at %s", strings.Join(urls, ", "))
}

func (m *Manager) setError(msg string) {
	log.Warn(msg)
	m.mu.Lock()
	m.lastError = msg
	m.mu.Unlock()
}

// Stop shuts the node down. It is a no-op when not running.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	srv := m.srv
	httpSrvs := m.httpSrvs
	cancel := m.cancel
	m.srv = nil
	m.httpSrvs = nil
	m.cancel = nil
	m.urls = nil
	m.dnsName = ""
	m.ips = nil
	m.listeners = nil
	m.mu.Unlock()
	if srv == nil {
		return nil
	}
	if cancel != nil {
		cancel()
	}
	for _, s := range httpSrvs {
		if errShutdown := s.Shutdown(ctx); errShutdown != nil {
			log.Debugf("tailscale: http shutdown: %v", errShutdown)
		}
	}
	if errClose := srv.Close(); errClose != nil {
		return fmt.Errorf("tailscale: close node: %w", errClose)
	}
	return nil
}

// Status reports the current node status.
func (m *Manager) Status(ctx context.Context, enabled bool) Status {
	m.mu.Lock()
	srv := m.srv
	cfg := m.cfg
	st := Status{
		Enabled:      enabled,
		Running:      srv != nil,
		Hostname:     cfg.Hostname,
		Funnel:       cfg.Funnel,
		Error:        m.lastError,
		URLs:         append([]string(nil), m.urls...),
		AuthKeyGiven: strings.TrimSpace(cfg.AuthKey) != "",
		State:        "Stopped",
	}
	m.mu.Unlock()
	if srv == nil {
		return st
	}
	lc, errLC := srv.LocalClient()
	if errLC != nil {
		st.State = "Starting"
		return st
	}
	ipn, errStatus := lc.StatusWithoutPeers(ctx)
	if errStatus != nil {
		st.State = "Starting"
		return st
	}
	st.State = ipn.BackendState
	st.AuthURL = ipn.AuthURL
	for _, ip := range ipn.TailscaleIPs {
		st.IPs = append(st.IPs, ip.String())
	}
	if ipn.Self != nil {
		st.DNSName = strings.TrimSuffix(ipn.Self.DNSName, ".")
	}
	m.mu.Lock()
	if st.DNSName != "" {
		m.dnsName = st.DNSName
	}
	if len(st.IPs) > 0 {
		m.ips = append([]string(nil), st.IPs...)
	}
	m.mu.Unlock()
	if len(st.URLs) == 0 && (st.DNSName != "" || len(st.IPs) > 0) {
		if st.DNSName != "" {
			st.URLs = append(st.URLs, fmt.Sprintf("http://%s", st.DNSName))
		}
		for _, ip := range st.IPs {
			st.URLs = append(st.URLs, fmt.Sprintf("http://%s", ip))
		}
	}
	return st
}

// IsOwnEndpoint checks whether the URL points to this running Tailscale node,
// verifying both that the host/IP matches and that the scheme and port match an active listener.
func (m *Manager) IsOwnEndpoint(parsed *url.URL) bool {
	if m == nil || parsed == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.srv == nil || len(m.urls) == 0 {
		return false
	}

	reqScheme := strings.ToLower(parsed.Scheme)
	reqPort := parsed.Port()
	if reqPort == "" {
		if reqScheme == "https" {
			reqPort = "443"
		} else if reqScheme == "http" {
			reqPort = "80"
		}
	}
	matchedListener := false
	for _, u := range m.urls {
		pu, err := url.Parse(u)
		if err != nil {
			continue
		}
		uScheme := strings.ToLower(pu.Scheme)
		uPort := pu.Port()
		if uPort == "" {
			if uScheme == "https" {
				uPort = "443"
			} else if uScheme == "http" {
				uPort = "80"
			}
		}
		if uScheme == reqScheme && uPort == reqPort {
			matchedListener = true
			break
		}
	}
	if !matchedListener {
		return false
	}

	host := parsed.Hostname()
	if host == "" {
		return false
	}
	if strings.EqualFold(host, m.cfg.Hostname) {
		return true
	}
	if m.dnsName != "" && (strings.EqualFold(host, m.dnsName) || strings.EqualFold(host, strings.TrimSuffix(m.dnsName, "."))) {
		return true
	}
	for _, ip := range m.ips {
		if strings.EqualFold(host, ip) {
			return true
		}
	}
	for _, u := range m.urls {
		if pu, err := url.Parse(u); err == nil {
			if strings.EqualFold(pu.Hostname(), host) {
				return true
			}
		}
	}
	return false
}

// ApplyConfig starts or stops the node to match cfg. It restarts the node when
// connection-relevant settings changed.
func (m *Manager) ApplyConfig(ctx context.Context, cfg config.TailscaleConfig) error {
	m.mu.Lock()
	if m.handler == nil {
		m.mu.Unlock()
		return nil
	}
	running := m.srv != nil
	same := m.cfg.Hostname == effectiveHostname(cfg) && m.cfg.Funnel == cfg.Funnel && m.cfg.HTTPS == cfg.HTTPS &&
		m.cfg.AuthKey == cfg.AuthKey && m.cfg.StateDir == cfg.StateDir && m.cfg.Ephemeral == cfg.Ephemeral
	m.mu.Unlock()
	switch {
	case !cfg.Enable && running:
		return m.Stop(ctx)
	case cfg.Enable && !running:
		return m.Start(cfg)
	case cfg.Enable && running && !same:
		if errStop := m.Stop(ctx); errStop != nil {
			return errStop
		}
		return m.Start(cfg)
	}
	return nil
}

func effectiveHostname(cfg config.TailscaleConfig) string {
	if h := strings.TrimSpace(cfg.Hostname); h != "" {
		return h
	}
	return DefaultHostname
}
