package agent

import (
	"context"
	"testing"
)

type stubProvider struct {
	response SummarizeResponse
	err      error
	model    string
}

func (s stubProvider) Name() ProviderName { return ProviderDeepSeek }
func (s stubProvider) Model() string      { return s.model }
func (s stubProvider) Summarize(ctx context.Context, req SummarizeRequest) (SummarizeResponse, error) {
	if s.err != nil {
		return SummarizeResponse{}, s.err
	}
	return s.response, nil
}

func TestSummarizerBuildsStructuredResult(t *testing.T) {
	provider := stubProvider{
		model: "deepseek-chat",
		response: SummarizeResponse{
			StandardAnswer: "这是一个适合背诵的标准答案，长度介于一百五十到两百二十字之间，表达更像真实面试回答。",
			MemoryOutline:  []string{"定义", "机制", "场景"},
			SourceQuotes:   []string{"原文依据 1", "原文依据 2"},
		},
	}

	s := NewSummarizer(provider, PromptVersion)
	got, err := s.Summarize(context.Background(), SummarizeRequest{
		Title:         "GMP 是什么？",
		Body:          "GMP 是 Go 的调度模型。",
		CandidateText: []string{"GMP 是 Go 的调度模型。"},
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if got.StandardAnswer == "" {
		t.Fatalf("expected standard answer")
	}
	if len(got.MemoryOutline) == 0 || len(got.SourceQuotes) == 0 {
		t.Fatalf("expected structured summary")
	}
	if got.Provider != string(ProviderDeepSeek) {
		t.Fatalf("unexpected provider: %q", got.Provider)
	}
	if got.Model != "deepseek-chat" {
		t.Fatalf("unexpected model: %q", got.Model)
	}
}
