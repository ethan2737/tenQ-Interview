package agent

import "context"

type ProviderName string

const (
	ProviderDeepSeek   ProviderName = "deepseek"
	ProviderModelScope ProviderName = "modelscope"
)

const PromptVersion = "v1"

type SummarizeRequest struct {
	Title         string
	Body          string
	CandidateText []string
	SystemPrompt  string
	UserPrompt    string
	PromptVersion string
}

type SummarizeResponse struct {
	Provider       string   `json:"provider,omitempty"`
	Model          string   `json:"model,omitempty"`
	StandardAnswer string   `json:"standard_answer"`
	MemoryOutline  []string `json:"memory_outline"`
	SourceQuotes   []string `json:"source_quotes"`
	Notes          string   `json:"notes,omitempty"`
}

type Provider interface {
	Name() ProviderName
	Model() string
	Summarize(ctx context.Context, req SummarizeRequest) (SummarizeResponse, error)
}
