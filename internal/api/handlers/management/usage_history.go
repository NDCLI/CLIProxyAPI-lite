package management

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/usageview"
)

func (h *Handler) GetUsageHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	c.JSON(http.StatusOK, gin.H{"records": usageview.Snapshot(limit)})
}

type usageSummary struct {
	Requests        int   `json:"requests"`
	Failed          int   `json:"failed"`
	InputTokens     int64 `json:"input_tokens"`
	CachedTokens    int64 `json:"cached_tokens"`
	OutputTokens    int64 `json:"output_tokens"`
	ReasoningTokens int64 `json:"reasoning_tokens"`
	TotalTokens     int64 `json:"total_tokens"`
	LatencyMs       int64 `json:"latency_ms"`
}

func usageFilterFromRequest(c *gin.Context) (usageview.Filter, error) {
	filter := usageview.Filter{
		Provider: strings.TrimSpace(c.Query("provider")),
		Model:    strings.TrimSpace(c.Query("model")),
	}
	if value := strings.TrimSpace(c.Query("status")); value != "" {
		switch strings.ToLower(value) {
		case "ok":
			failed := false
			filter.Failed = &failed
		case "failed":
			failed := true
			filter.Failed = &failed
		default:
			return usageview.Filter{}, strconv.ErrSyntax
		}
	}
	for _, query := range []struct {
		name   string
		target *time.Time
	}{{"from", &filter.From}, {"to", &filter.To}} {
		value := strings.TrimSpace(c.Query(query.name))
		if value == "" {
			continue
		}
		parsed, errParse := time.Parse(time.RFC3339, value)
		if errParse != nil {
			return usageview.Filter{}, errParse
		}
		*query.target = parsed
	}
	return filter, nil
}

func usageLimit(c *gin.Context) int {
	limit, errParse := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if errParse != nil || limit <= 0 {
		return 100
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

// GetUsageRecords returns normalized, filterable persisted usage records.
func (h *Handler) GetUsageRecords(c *gin.Context) {
	filter, errFilter := usageFilterFromRequest(c)
	if errFilter != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_usage_filter", "message": "Invalid usage filter"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": usageview.FilteredSnapshot(usageLimit(c), filter), "next_cursor": nil})
}

// GetUsageSummary returns totals over the same filterable usage read model.
func (h *Handler) GetUsageSummary(c *gin.Context) {
	filter, errFilter := usageFilterFromRequest(c)
	if errFilter != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_usage_filter", "message": "Invalid usage filter"}})
		return
	}
	summary := usageSummary{}
	for _, record := range usageview.FilteredSnapshot(1000, filter) {
		summary.Requests++
		if record.Failed {
			summary.Failed++
		}
		summary.InputTokens += record.InputTokens
		summary.CachedTokens += record.CachedTokens
		summary.OutputTokens += record.OutputTokens
		summary.ReasoningTokens += record.ReasoningTokens
		summary.TotalTokens += record.TotalTokens
		summary.LatencyMs += record.LatencyMs
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": summary})
}

// StreamUsageHistory pushes a fresh snapshot whenever a provider request completes.
// The management middleware authenticates the long-lived connection before it starts.
func (h *Handler) StreamUsageHistory(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	notifications, unsubscribe := usageview.Subscribe()
	defer unsubscribe()
	write := func() bool {
		payload, err := json.Marshal(gin.H{"records": usageview.Snapshot(200)})
		if err != nil {
			return false
		}
		if _, err = c.Writer.Write([]byte("data: ")); err != nil {
			return false
		}
		if _, err = c.Writer.Write(payload); err != nil {
			return false
		}
		if _, err = c.Writer.Write([]byte("\n\n")); err != nil {
			return false
		}
		c.Writer.Flush()
		return true
	}

	if !write() {
		return
	}
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-notifications:
			if !write() {
				return
			}
		}
	}
}
