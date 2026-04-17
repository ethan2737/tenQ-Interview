package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanMarkdownPathsRecursively(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "backend", "runtime")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}

	files := map[string]string{
		filepath.Join(root, "intro.md"):                  "# Intro\n\nHello",
		filepath.Join(nested, "scheduler.MD"):            "# Scheduler\n\nBody",
		filepath.Join(root, "notes.txt"):                 "skip me",
		filepath.Join(root, "backend", "draft.markdown"): "skip me too",
	}

	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("failed to write fixture %s: %v", path, err)
		}
	}

	entries, err := ScanMarkdownPaths(root)
	if err != nil {
		t.Fatalf("ScanMarkdownPaths returned error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 markdown files, got %d", len(entries))
	}

	if entries[0].RelativePath != "backend/runtime/scheduler.MD" {
		t.Fatalf("unexpected first relative path: %q", entries[0].RelativePath)
	}
	if entries[1].RelativePath != "intro.md" {
		t.Fatalf("unexpected second relative path: %q", entries[1].RelativePath)
	}
}

func TestScanMarkdownPathsSupportsSingleFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "gmp.md")
	if err := os.WriteFile(path, []byte("# GMP\n\nBody"), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	entries, err := ScanMarkdownPaths(path)
	if err != nil {
		t.Fatalf("ScanMarkdownPaths returned error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != path {
		t.Fatalf("unexpected path: %q", entries[0].Path)
	}
	if entries[0].RelativePath != "gmp.md" {
		t.Fatalf("unexpected relative path: %q", entries[0].RelativePath)
	}
}
