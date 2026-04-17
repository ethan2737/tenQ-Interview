package card

import (
	"strings"
	"testing"

	"tenq-interview/internal/segment"
)

func TestGenerateCardBuildsBoundedAnswerFromSourceSegments(t *testing.T) {
	t.Parallel()

	segments := []segment.CandidateSegment{
		{
			Text:  "GMP 是 Go 的协程调度模型，G 表示 goroutine，M 表示线程，P 表示处理器上下文。调度器通过它把大量 goroutine 分配到更少的线程上运行。",
			Score: 6,
		},
		{
			Text:  "这能提升并发执行效率，同时避免线程数量失控带来的开销。",
			Score: 3,
		},
	}

	card, err := GenerateCard("Go 的 GMP 模型是什么？", segments)
	if err != nil {
		t.Fatalf("GenerateCard returned error: %v", err)
	}

	if card.Question != "Go 的 GMP 模型是什么？" {
		t.Fatalf("unexpected question: %q", card.Question)
	}

	if got := len([]rune(card.Answer)); got > 250 {
		t.Fatalf("expected answer length <= 250 runes, got %d", got)
	}

	if !strings.Contains(card.Answer, "GMP") {
		t.Fatalf("expected answer to preserve key source wording")
	}

	if len(card.Sources) == 0 {
		t.Fatalf("expected card sources to be retained")
	}
}
