package usageview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const maxRecords = 1000

type Record struct {
	Timestamp           time.Time `json:"timestamp"`
	Provider            string    `json:"provider"`
	Model               string    `json:"model"`
	Alias               string    `json:"alias,omitempty"`
	Endpoint            string    `json:"endpoint,omitempty"`
	InputTokens         int64     `json:"input_tokens"`
	CachedTokens        int64     `json:"cached_tokens,omitempty"`
	CacheCreationTokens int64     `json:"cache_creation_tokens,omitempty"`
	OutputTokens        int64     `json:"output_tokens"`
	ReasoningTokens     int64     `json:"reasoning_tokens,omitempty"`
	TotalTokens         int64     `json:"total_tokens"`
	LatencyMs           int64     `json:"latency_ms,omitempty"`
	Failed              bool      `json:"failed"`
}

var history struct {
	sync.RWMutex
	items []Record
	subs  map[chan struct{}]struct{}
	path  string
}

// ConfigurePersistence loads usage history stored next to the active config.
func ConfigurePersistence(configPath string) {
	configPath = strings.TrimSpace(configPath)
	if configPath == "" {
		return
	}
	path := filepath.Join(filepath.Dir(configPath), "usage-history.json")
	history.Lock()
	defer history.Unlock()
	if history.path == path {
		return
	}
	history.path = path
	body, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var records []Record
	if json.Unmarshal(body, &records) != nil {
		return
	}
	if len(records) > maxRecords {
		records = records[len(records)-maxRecords:]
	}
	history.items = append(history.items, records...)
	if len(history.items) > maxRecords {
		history.items = history.items[len(history.items)-maxRecords:]
	}
}

// Subscribe returns a notification channel for completed usage records.
// Notifications are coalesced so a slow dashboard cannot block request handling.
func Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	history.Lock()
	if history.subs == nil {
		history.subs = make(map[chan struct{}]struct{})
	}
	history.subs[ch] = struct{}{}
	history.Unlock()
	return ch, func() {
		history.Lock()
		if _, ok := history.subs[ch]; ok {
			delete(history.subs, ch)
			close(ch)
		}
		history.Unlock()
	}
}

func Add(record Record) {
	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now()
	}
	record.Provider = strings.TrimSpace(record.Provider)
	record.Model = strings.TrimSpace(record.Model)
	record.Alias = strings.TrimSpace(record.Alias)
	record.Endpoint = strings.TrimSpace(record.Endpoint)
	history.Lock()
	history.items = append(history.items, record)
	if len(history.items) > maxRecords {
		history.items = append([]Record(nil), history.items[len(history.items)-maxRecords:]...)
	}
	for ch := range history.subs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	path := history.path
	items := append([]Record(nil), history.items...)
	history.Unlock()
	if path != "" {
		if body, err := json.Marshal(items); err == nil {
			_ = os.WriteFile(path, body, 0o600)
		}
	}
}

func Snapshot(limit int) []Record {
	if limit <= 0 || limit > maxRecords {
		limit = maxRecords
	}
	history.RLock()
	defer history.RUnlock()
	start := len(history.items) - limit
	if start < 0 {
		start = 0
	}
	out := append([]Record(nil), history.items[start:]...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Timestamp.After(out[j].Timestamp) })
	return out
}
