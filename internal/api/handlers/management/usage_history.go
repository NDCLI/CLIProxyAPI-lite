package management

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/usageview"
)

func (h *Handler) GetUsageHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	c.JSON(http.StatusOK, gin.H{"records": usageview.Snapshot(limit)})
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
