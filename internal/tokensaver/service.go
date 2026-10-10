package tokensaver

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
)

const (
	maxRequestBytes = 16 << 20
)

type Statistics struct {
	RTKRequests        uint64 `json:"rtk_requests"`
	RTKHits            uint64 `json:"rtk_hits"`
	RTKBytesSaved      uint64 `json:"rtk_bytes_saved"`
	PerRequestBypasses uint64 `json:"per_request_bypasses"`
}

type Snapshot struct {
	Config     config.TokenSaverConfig `json:"config"`
	Statistics Statistics              `json:"statistics"`
}

// Service applies opt-in compression to supported JSON requests before translation.
type Service struct {
	mu          sync.RWMutex
	cfg         config.TokenSaverConfig
	rtkRequests atomic.Uint64
	rtkHits     atomic.Uint64
	rtkSaved    atomic.Uint64
	bypasses    atomic.Uint64
}

func NewService(value config.TokenSaverConfig) *Service {
	service := &Service{}
	service.Configure(value)
	return service
}

func (s *Service) Configure(value config.TokenSaverConfig) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.cfg = value
	s.mu.Unlock()
}

func (s *Service) Snapshot() Snapshot {
	if s == nil {
		return Snapshot{}
	}
	s.mu.RLock()
	value := Snapshot{Config: s.cfg}
	s.mu.RUnlock()
	value.Statistics = Statistics{
		RTKRequests:        s.rtkRequests.Load(),
		RTKHits:            s.rtkHits.Load(),
		RTKBytesSaved:      s.rtkSaved.Load(),
		PerRequestBypasses: s.bypasses.Load(),
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
	return value.RTKEnabled
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

// Process preserves the original payload when a filter cannot safely reduce it.
func (s *Service) Process(ctx context.Context, headers http.Header, body []byte) []byte {
	if s == nil || len(body) == 0 || len(body) > maxRequestBytes || !gjson.ValidBytes(body) {
		return body
	}
	value := s.currentConfig()
	if !value.RTKEnabled {
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
	return processed
}
