package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessorProcessesMarkdownFileIntoCard(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "gmp.md")
	content := "# Go 的 GMP 模型是什么？\n\nGMP 是 Go 的调度模型，G 表示 goroutine，M 表示线程，P 表示处理器上下文。\n\n它让调度器可以把大量 goroutine 分配到更少的线程上执行。"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	processor := NewProcessor()
	result, err := processor.ProcessFile(path)
	if err != nil {
		t.Fatalf("ProcessFile returned error: %v", err)
	}

	if result.Title != "Go 的 GMP 模型是什么？" {
		t.Fatalf("unexpected title: %q", result.Title)
	}

	if result.Card.Answer == "" {
		t.Fatalf("expected answer to be generated")
	}

	if !strings.Contains(result.Card.Answer, "GMP") {
		t.Fatalf("expected answer to preserve source wording")
	}

	if result.Encoding == "" {
		t.Fatalf("expected encoding metadata")
	}
}
