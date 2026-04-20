package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCachePathUsesProjectDirectory(t *testing.T) {
	executablePath, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable returned error: %v", err)
	}

	want := filepath.Join(filepath.Dir(executablePath), ".cache", "tenq-interview", "index.json")
	got := defaultCachePath()

	if got != want {
		t.Fatalf("expected cache path %q, got %q", want, got)
	}
}

func TestDefaultCacheBaseDirUsesExecutableDirectory(t *testing.T) {
	executablePath, err := os.Executable()
	if err != nil {
		t.Fatalf("Executable returned error: %v", err)
	}

	want := filepath.Dir(executablePath)
	got := defaultCacheBaseDir()
	if got != want {
		t.Fatalf("expected cache base dir %q, got %q", want, got)
	}
}

func TestConfigRootsPreferExplicitAndExecutableLocations(t *testing.T) {
	got := configRootsFor(
		`C:\workspace\tenq`,
		`C:\release\tenq`,
		`C:\Users\Administrator\AppData\Roaming\tenq-interview`,
		`D:\custom-config`,
	)

	want := []string{
		`D:\custom-config`,
		`C:\release\tenq`,
		`C:\workspace\tenq`,
		`C:\Users\Administrator\AppData\Roaming\tenq-interview`,
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d config roots, got %d (%v)", len(want), len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("expected config root %d to be %q, got %q", index, want[index], got[index])
		}
	}
}

func TestConfigRootsSkipEmptyAndDuplicateEntries(t *testing.T) {
	got := configRootsFor(
		`C:\workspace\tenq`,
		`C:\workspace\tenq`,
		``,
		``,
	)

	want := []string{`C:\workspace\tenq`}
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("expected deduplicated config roots %v, got %v", want, got)
	}
}
