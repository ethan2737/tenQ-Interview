package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
)

type Entry struct {
	Key           string   `json:"key"`
	Path          string   `json:"path"`
	Title         string   `json:"title"`
	Encoding      string   `json:"encoding"`
	Provider      string   `json:"provider,omitempty"`
	Model         string   `json:"model,omitempty"`
	CardAnswer    string   `json:"cardAnswer"`
	MemoryOutline []string `json:"memoryOutline,omitempty"`
	SourceTexts   []string `json:"sourceTexts"`
	Notes         string   `json:"notes,omitempty"`
	PromptVersion string   `json:"promptVersion,omitempty"`
}

type Store struct {
	Entries map[string]Entry `json:"entries"`
}

func NewStore() *Store {
	return &Store{Entries: map[string]Entry{}}
}

func LoadStore(path string) (*Store, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return NewStore(), nil
	}
	if err != nil {
		return nil, err
	}

	var store Store
	if err := json.Unmarshal(raw, &store); err != nil {
		return nil, err
	}
	if store.Entries == nil {
		store.Entries = map[string]Entry{}
	}
	return &store, nil
}

func (s *Store) Get(key string) (Entry, bool) {
	entry, ok := s.Entries[key]
	return entry, ok
}

func (s *Store) Put(key string, entry Entry) {
	if s.Entries == nil {
		s.Entries = map[string]Entry{}
	}
	s.Entries[key] = entry
}

func (s *Store) Clear() {
	if s == nil {
		return
	}
	s.Entries = map[string]Entry{}
}

func (s *Store) List() []Entry {
	if s == nil || len(s.Entries) == 0 {
		return nil
	}

	entries := make([]Entry, 0, len(s.Entries))
	for _, entry := range s.Entries {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i int, j int) bool {
		if entries[i].Path == entries[j].Path {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Path < entries[j].Path
	})

	return entries
}

func (s *Store) Save(path string) error {
	if s == nil {
		return errors.New("store is nil")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}
