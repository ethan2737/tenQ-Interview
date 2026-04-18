package agent

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestOpenAICompatibleProviderParsesStructuredResponse(t *testing.T) {
	provider := &openAICompatibleProvider{
		name:    ProviderDeepSeek,
		baseURL: "https://example.com",
		apiKey:  "key",
		model:   "deepseek-chat",
		client: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body := `{"choices":[{"message":{"content":"{\"standard_answer\":\"标准答案\",\"memory_outline\":[\"定义\",\"机制\"],\"source_quotes\":[\"依据1\",\"依据2\"]}"}}]}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	got, err := provider.Summarize(context.Background(), SummarizeRequest{
		SystemPrompt: "system",
		UserPrompt:   "user",
	})
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if got.StandardAnswer != "标准答案" {
		t.Fatalf("unexpected answer: %q", got.StandardAnswer)
	}
	if len(got.MemoryOutline) != 2 {
		t.Fatalf("unexpected memory outline count: %d", len(got.MemoryOutline))
	}
}

func TestNewProviderRequiresConfiguredAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderName
		cfg      Config
		wantErr  string
	}{
		{
			name:     "deepseek requires key",
			provider: ProviderDeepSeek,
			cfg: Config{
				DeepSeek: ProviderConfig{BaseURL: "https://api.deepseek.com", Model: "deepseek-chat"},
			},
			wantErr: "deepseek api key is required",
		},
		{
			name:     "modelscope requires key",
			provider: ProviderModelScope,
			cfg: Config{
				ModelScope: ProviderConfig{BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Model: "qwen-plus"},
			},
			wantErr: "modelscope api key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProvider(tt.provider, tt.cfg)
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected error %q, got %v", tt.wantErr, err)
			}
		})
	}
}
