package usageview

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSnapshotReturnsNewestFirst(t *testing.T) {
	history.Lock()
	history.items = nil
	history.Unlock()
	Add(Record{Timestamp: time.Unix(1, 0), Model: "old"})
	Add(Record{Timestamp: time.Unix(2, 0), Model: "new"})
	got := Snapshot(10)
	if len(got) != 2 || got[0].Model != "new" {
		t.Fatalf("snapshot = %#v", got)
	}
}

func TestConfigurePersistenceReloadsHistory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	history.Lock()
	history.items = nil
	history.path = ""
	history.Unlock()
	ConfigurePersistence(configPath)
	Add(Record{Timestamp: time.Unix(3, 0), Provider: "test", Model: "model", TotalTokens: 24})

	history.Lock()
	history.items = nil
	history.path = ""
	history.Unlock()
	ConfigurePersistence(configPath)
	got := Snapshot(10)
	if len(got) != 1 || got[0].TotalTokens != 24 {
		t.Fatalf("reloaded snapshot = %#v", got)
	}
}

func TestFilteredSnapshotMatchesProviderModelAndStatus(t *testing.T) {
	history.Lock()
	history.items = nil
	history.Unlock()
	failed := true
	Add(Record{Timestamp: time.Unix(1, 0), Provider: "codex", Model: "gpt-5", Failed: true})
	Add(Record{Timestamp: time.Unix(2, 0), Provider: "claude", Model: "sonnet", Alias: "fast", Failed: false})
	got := FilteredSnapshot(10, Filter{Provider: "claude", Model: "fast", Failed: &failed})
	if len(got) != 0 {
		t.Fatalf("filtered failed records = %#v, want none", got)
	}
	got = FilteredSnapshot(10, Filter{Provider: "claude", Model: "fast"})
	if len(got) != 1 || got[0].Model != "sonnet" {
		t.Fatalf("filtered records = %#v", got)
	}
}
