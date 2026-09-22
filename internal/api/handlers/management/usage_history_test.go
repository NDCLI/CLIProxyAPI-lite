package management

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/usageview"
)

func TestUsageRecordsAndSummaryUseSameFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	usageview.ConfigurePersistence("")
	usageview.Add(usageview.Record{Timestamp: time.Now().UTC(), Provider: "usage-filter-test", Model: "gpt-5", InputTokens: 2, OutputTokens: 3, TotalTokens: 5})
	handler := &Handler{}
	records := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(records)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/usage/records?provider=usage-filter-test", nil)
	handler.GetUsageRecords(ctx)
	if records.Code != http.StatusOK || !contains(records.Body.String(), `"provider":"usage-filter-test"`) {
		t.Fatalf("records = %d %s", records.Code, records.Body.String())
	}
	summary := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(summary)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/usage/summary?provider=usage-filter-test", nil)
	handler.GetUsageSummary(ctx)
	if summary.Code != http.StatusOK || !contains(summary.Body.String(), `"total_tokens":5`) {
		t.Fatalf("summary = %d %s", summary.Code, summary.Body.String())
	}
}
