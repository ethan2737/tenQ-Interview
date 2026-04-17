package cache

import (
	"path/filepath"
	"testing"
)

func TestStoreSaveAndLoadRoundTrip(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "cache.json")

	store := NewStore()
	store.Put("abc", Entry{
		Key:         "abc",
		Path:        "docs/gmp.md",
		Title:       "Go 的 GMP 模型是什么？",
		Encoding:    "utf-8",
		CardAnswer:  "GMP 是 Go 的调度模型。",
		SourceTexts: []string{"GMP 是 Go 的调度模型。"},
	})

	if err := store.Save(path); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := LoadStore(path)
	if err != nil {
		t.Fatalf("LoadStore returned error: %v", err)
	}

	entry, ok := loaded.Get("abc")
	if !ok {
		t.Fatalf("expected stored key to be present")
	}
	if entry.Title != "Go 的 GMP 模型是什么？" {
		t.Fatalf("unexpected title: %q", entry.Title)
	}
}
