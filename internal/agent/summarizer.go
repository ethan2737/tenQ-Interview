package agent

import (
	"context"
	"errors"
	"strings"
)

type Summarizer struct {
	provider Provider
	version  string
}

func NewSummarizer(provider Provider, version string) *Summarizer {
	return &Summarizer{
		provider: provider,
		version:  version,
	}
}

func (s *Summarizer) ProviderModel() string {
	if s == nil || s.provider == nil {
		return ""
	}
	return s.provider.Model()
}

func (s *Summarizer) PromptVersion() string {
	if s == nil {
		return ""
	}
	return s.version
}

func (s *Summarizer) Summarize(ctx context.Context, req SummarizeRequest) (SummarizeResponse, error) {
	if s == nil || s.provider == nil {
		return SummarizeResponse{}, errors.New("provider is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return SummarizeResponse{}, errors.New("title is required")
	}
	if strings.TrimSpace(req.Body) == "" {
		return SummarizeResponse{}, errors.New("body is required")
	}

	req.PromptVersion = s.version
	req.SystemPrompt = BuildSystemPrompt()
	req.UserPrompt = BuildUserPrompt(req)

	resp, err := s.provider.Summarize(ctx, req)
	if err != nil {
		return SummarizeResponse{}, err
	}
	if strings.TrimSpace(resp.StandardAnswer) == "" {
		return SummarizeResponse{}, errors.New("standard answer is required")
	}
	if len(resp.MemoryOutline) == 0 {
		return SummarizeResponse{}, errors.New("memory outline is required")
	}
	if len(resp.SourceQuotes) == 0 {
		return SummarizeResponse{}, errors.New("source quotes are required")
	}
	if resp.Provider == "" {
		resp.Provider = string(s.provider.Name())
	}
	if resp.Model == "" {
		resp.Model = s.provider.Model()
	}
	return resp, nil
}
