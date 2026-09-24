package tokensaver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
)

const (
	DefaultHeadroomURL       = "http://127.0.0.1:8787"
	DefaultHeadroomTimeoutMS = 3000
	maxRequestBytes          = 16 << 20
)

type Statistics struct {
	RTKRequests        uint64 `json:"rtk_requests"`
	RTKHits            uint64 `json:"rtk_hits"`
	RTKBytesSaved      uint64 `json:"rtk_bytes_saved"`
	HeadroomRequests   uint64 `json:"headroom_requests"`
	HeadroomApplied    uint64 `json:"headroom_applied"`
	HeadroomBytesSaved uint64 `json:"headroom_bytes_saved"`
	HeadroomFailures   uint64 `json:"headroom_failures"`
	PerRequestBypasses uint64 `json:"per_request_bypasses"`
}

type Snapshot struct {
	Config         config.TokenSaverConfig `json:"config"`
	Statistics     Statistics              `json:"statistics"`
	HeadroomStatus string                  `json:"headroom_status"`
}

// Service applies opt-in compression to supported JSON requests before translation.
type Service struct {
	mu             sync.RWMutex
	cfg            config.TokenSaverConfig
	headroomStatus string
	client         *http.Client
	rtkRequests    atomic.Uint64
	rtkHits        atomic.Uint64
	rtkSaved       atomic.Uint64
	headRequests   atomic.Uint64
	headApplied    atomic.Uint64
	headSaved      atomic.Uint64
	headFailures   atomic.Uint64
	bypasses       atomic.Uint64
}

func NormalizeConfig(value config.TokenSaverConfig) (config.TokenSaverConfig, error) {
	value.HeadroomURL = strings.TrimRight(strings.TrimSpace(value.HeadroomURL), "/")
	if value.HeadroomURL == "" {
		value.HeadroomURL = DefaultHeadroomURL
	}
	parsed, errParse := url.Parse(value.HeadroomURL)
	if errParse != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return value, errors.New("Headroom base URL must be HTTP(S) without credentials, query, or fragment")
	}
	if strings.HasSuffix(strings.TrimRight(parsed.Path, "/"), "/v1/compress") {
		return value, errors.New("Headroom URL must be the service base URL")
	}
	if value.HeadroomTimeoutMS == 0 {
		value.HeadroomTimeoutMS = DefaultHeadroomTimeoutMS
	}
	if value.HeadroomTimeoutMS < 500 || value.HeadroomTimeoutMS > 15000 {
		return value, errors.New("Headroom timeout must be between 500 and 15000 ms")
	}
	return value, nil
}

func NewService(value config.TokenSaverConfig) *Service {
	service := &Service{client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	service.Configure(value)
	return service
}

func (s *Service) Configure(value config.TokenSaverConfig) {
	if s == nil {
		return
	}
	normalized, err := NormalizeConfig(value)
	if err != nil {
		normalized = config.TokenSaverConfig{RTKEnabled: value.RTKEnabled, HeadroomURL: DefaultHeadroomURL, HeadroomTimeoutMS: DefaultHeadroomTimeoutMS}
	}
	s.mu.Lock()
	if s.cfg.HeadroomURL != normalized.HeadroomURL {
		s.headroomStatus = "unknown"
	}
	s.cfg = normalized
	if s.headroomStatus == "" {
		s.headroomStatus = "unknown"
	}
	s.mu.Unlock()
}

func (s *Service) Snapshot() Snapshot {
	if s == nil {
		return Snapshot{}
	}
	s.mu.RLock()
	value := Snapshot{Config: s.cfg, HeadroomStatus: s.headroomStatus}
	s.mu.RUnlock()
	value.Statistics = Statistics{
		RTKRequests: s.rtkRequests.Load(), RTKHits: s.rtkHits.Load(), RTKBytesSaved: s.rtkSaved.Load(),
		HeadroomRequests: s.headRequests.Load(), HeadroomApplied: s.headApplied.Load(), HeadroomBytesSaved: s.headSaved.Load(),
		HeadroomFailures: s.headFailures.Load(), PerRequestBypasses: s.bypasses.Load(),
	}
	return value
}

func (s *Service) currentConfig() config.TokenSaverConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Service) enabled() bool {
	value := s.currentConfig()
	return value.RTKEnabled || value.HeadroomEnabled
}

type restoredBody struct {
	io.Reader
	io.Closer
}

func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s == nil || !s.enabled() || c.Request.Method != http.MethodPost || c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}
		if encoding := c.GetHeader("Content-Encoding"); encoding != "" && !strings.EqualFold(encoding, "identity") {
			c.Next()
			return
		}
		if contentType := c.GetHeader("Content-Type"); contentType != "" && !strings.Contains(strings.ToLower(contentType), "json") {
			c.Next()
			return
		}
		original := c.Request.Body
		body, errRead := io.ReadAll(io.LimitReader(original, maxRequestBytes+1))
		if errRead != nil || len(body) > maxRequestBytes {
			c.Request.Body = restoredBody{Reader: io.MultiReader(bytes.NewReader(body), original), Closer: original}
			c.Next()
			return
		}
		_ = original.Close()
		processed := s.Process(c.Request.Context(), c.Request.Header, body)
		c.Request.Body = io.NopCloser(bytes.NewReader(processed))
		if len(processed) != len(body) {
			c.Request.ContentLength = int64(len(processed))
			c.Request.Header.Set("Content-Length", strconv.Itoa(len(processed)))
		}
		c.Next()
	}
}

// Process preserves the original payload when a filter or external service cannot safely reduce it.
func (s *Service) Process(ctx context.Context, headers http.Header, body []byte) []byte {
	if s == nil || len(body) == 0 || len(body) > maxRequestBytes || !gjson.ValidBytes(body) {
		return body
	}
	value := s.currentConfig()
	if !value.RTKEnabled && !value.HeadroomEnabled {
		return body
	}
	if strings.EqualFold(headers.Get("X-CLIProxy-Token-Saver"), "off") || strings.EqualFold(headers.Get("X-9Router-Token-Saver"), "off") {
		s.bypasses.Add(1)
		return body
	}
	processed := body
	if value.RTKEnabled {
		s.rtkRequests.Add(1)
		compressed, hits := compressRTK(processed)
		if hits > 0 {
			s.rtkHits.Add(uint64(hits))
			s.rtkSaved.Add(uint64(len(processed) - len(compressed)))
			processed = compressed
		}
	}
	if value.HeadroomEnabled && gjson.GetBytes(processed, "messages").IsArray() {
		s.headRequests.Add(1)
		compressed, reason := s.compressHeadroom(ctx, value, processed)
		if reason != "" {
			s.headFailures.Add(1)
			s.setHeadroomStatus(reason)
		} else if len(compressed) < len(processed) {
			s.headApplied.Add(1)
			s.headSaved.Add(uint64(len(processed) - len(compressed)))
			s.setHeadroomStatus("ready")
			processed = compressed
		} else {
			s.setHeadroomStatus("ready")
		}
	}
	return processed
}

func (s *Service) setHeadroomStatus(value string) {
	s.mu.Lock()
	s.headroomStatus = value
	s.mu.Unlock()
}
