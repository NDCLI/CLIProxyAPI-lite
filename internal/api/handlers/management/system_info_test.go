package management

import (
	"encoding/json"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/buildinfo"
)

func TestGetSystemInfoContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	(&Handler{}).GetSystemInfo(ctx)

	var response struct {
		SchemaVersion int            `json:"schema_version"`
		Item          systemInfoItem `json:"item"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", response.SchemaVersion)
	}
	if response.Item.Version != buildinfo.Version || response.Item.Commit != buildinfo.Commit || response.Item.BuildDate != buildinfo.BuildDate {
		t.Fatalf("build metadata = %#v, does not match buildinfo", response.Item)
	}
	if response.Item.GoVersion != runtime.Version() || response.Item.OperatingSystem != runtime.GOOS || response.Item.Architecture != runtime.GOARCH {
		t.Fatalf("runtime metadata = %#v, does not match runtime", response.Item)
	}
}
