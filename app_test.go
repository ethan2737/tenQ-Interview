package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCachePathUsesProjectDirectory(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}

	want := filepath.Join(workingDir, ".cache", "tenq-interview", "index.json")
	got := defaultCachePath()

	if got != want {
		t.Fatalf("expected cache path %q, got %q", want, got)
	}
}
