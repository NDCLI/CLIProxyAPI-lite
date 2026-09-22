package combo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStorePersistsOrderedTargets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	store := New(path)
	def, errSave := store.Save(Definition{ID: "fast", Name: "Fast", Model: "fast", Enabled: true, Targets: []Target{{Provider: "codex", Model: "gpt-5"}, {Provider: "claude", Model: "sonnet"}}})
	if errSave != nil || len(def.Targets) != 2 {
		t.Fatalf("save = %#v %v", def, errSave)
	}
	if _, errRead := os.Stat(filepath.Join(filepath.Dir(path), "combos.json")); errRead != nil {
		t.Fatal(errRead)
	}
	reloaded := New(path).List()
	if len(reloaded) != 1 || reloaded[0].Targets[1].Provider != "claude" {
		t.Fatalf("reloaded = %#v", reloaded)
	}
}
