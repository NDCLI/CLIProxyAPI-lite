// Package combo stores ordered model routing definitions for management APIs.
package combo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Target struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type Definition struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Model   string   `json:"model"`
	Enabled bool     `json:"enabled"`
	Vision  bool     `json:"vision"`
	Targets []Target `json:"targets"`
}

type Store struct {
	mu    sync.RWMutex
	path  string
	items []Definition
}

func New(configPath string) *Store {
	store := &Store{path: filepath.Join(filepath.Dir(configPath), "combos.json")}
	store.load()
	return store
}

func (s *Store) load() {
	if s == nil || s.path == "" {
		return
	}
	body, errRead := os.ReadFile(s.path)
	if errRead != nil {
		return
	}
	var items []Definition
	if json.Unmarshal(body, &items) == nil {
		s.items = items
	}
}

func validate(def Definition) (Definition, error) {
	def.ID, def.Name, def.Model = strings.TrimSpace(def.ID), strings.TrimSpace(def.Name), strings.TrimSpace(def.Model)
	if def.ID == "" || def.Name == "" || def.Model == "" {
		return Definition{}, fmt.Errorf("id, name, and model are required")
	}
	if len(def.Targets) == 0 {
		return Definition{}, fmt.Errorf("at least one target is required")
	}
	for index := range def.Targets {
		def.Targets[index].Provider = strings.ToLower(strings.TrimSpace(def.Targets[index].Provider))
		def.Targets[index].Model = strings.TrimSpace(def.Targets[index].Model)
		if def.Targets[index].Provider == "" || def.Targets[index].Model == "" {
			return Definition{}, fmt.Errorf("target %d requires provider and model", index+1)
		}
	}
	return def, nil
}

func (s *Store) List() []Definition {
	if s == nil {
		return []Definition{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Definition(nil), s.items...)
}

func (s *Store) Save(def Definition) (Definition, error) {
	if s == nil {
		return Definition{}, fmt.Errorf("combo store unavailable")
	}
	def, errValidate := validate(def)
	if errValidate != nil {
		return Definition{}, errValidate
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.items {
		if s.items[index].ID == def.ID {
			s.items[index] = def
			return def, s.persistLocked()
		}
	}
	s.items = append(s.items, def)
	return def, s.persistLocked()
}

func (s *Store) Delete(id string) error {
	if s == nil {
		return fmt.Errorf("combo store unavailable")
	}
	id = strings.TrimSpace(id)
	s.mu.Lock()
	defer s.mu.Unlock()
	for index := range s.items {
		if s.items[index].ID == id {
			s.items = append(s.items[:index], s.items[index+1:]...)
			return s.persistLocked()
		}
	}
	return os.ErrNotExist
}

func (s *Store) persistLocked() error {
	if errDir := os.MkdirAll(filepath.Dir(s.path), 0o700); errDir != nil {
		return errDir
	}
	body, errMarshal := json.MarshalIndent(s.items, "", "  ")
	if errMarshal != nil {
		return errMarshal
	}
	return os.WriteFile(s.path, body, 0o600)
}
