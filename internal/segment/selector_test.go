package segment

import "testing"

func TestSelectCandidateSegmentsPrioritisesRelevantParagraphs(t *testing.T) {
	t.Parallel()

	question := "Go 的 GMP 模型是什么？"
	body := "GMP 是 Go 的调度模型，G 表示 goroutine，M 表示线程，P 表示处理器上下文。\n\n这一套设计的目的，是把大量 goroutine 映射到更少的线程上执行。\n\n这段话在讲别的事情，比如简历技巧。"

	segments := SelectCandidateSegments(question, body, 2)
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if segments[0].Text == "" {
		t.Fatalf("expected first segment to be populated")
	}

	if segments[0].Score < segments[1].Score {
		t.Fatalf("expected first segment to be the highest-scoring one")
	}
}

func TestSelectCandidateSegmentsPreservesFencedCodeBlocks(t *testing.T) {
	t.Parallel()

	body := "```go\nfmt.Println(\"hello\")\nfmt.Println(\"world\")\n```\n\n后续说明。"

	segments := SelectCandidateSegments("打印示例", body, 1)
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	if segments[0].Text != "```go\nfmt.Println(\"hello\")\nfmt.Println(\"world\")\n```" {
		t.Fatalf("expected code fence to be preserved, got %q", segments[0].Text)
	}
}
