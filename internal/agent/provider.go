package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type openAICompatibleProvider struct {
	name    ProviderName
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

type chatCompletionRequest struct {
	Model       string         `json:"model"`
	Messages    []chatMessage  `json:"messages"`
	Temperature float64        `json:"temperature"`
	ResponseFmt responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func NewProvider(name ProviderName, cfg Config) (Provider, error) {
	client := &http.Client{Timeout: 45 * time.Second}

	switch name {
	case ProviderDeepSeek:
		if strings.TrimSpace(cfg.DeepSeek.APIKey) == "" {
			return nil, errors.New("deepseek api key is required")
		}
		return &openAICompatibleProvider{
			name:    ProviderDeepSeek,
			baseURL: strings.TrimRight(cfg.DeepSeek.BaseURL, "/"),
			apiKey:  cfg.DeepSeek.APIKey,
			model:   cfg.DeepSeek.Model,
			client:  client,
		}, nil
	case ProviderModelScope:
		if strings.TrimSpace(cfg.ModelScope.APIKey) == "" {
			return nil, errors.New("modelscope api key is required")
		}
		return &openAICompatibleProvider{
			name:    ProviderModelScope,
			baseURL: strings.TrimRight(cfg.ModelScope.BaseURL, "/"),
			apiKey:  cfg.ModelScope.APIKey,
			model:   cfg.ModelScope.Model,
			client:  client,
		}, nil
	default:
		return nil, errors.New("unsupported provider")
	}
}

func (p *openAICompatibleProvider) Name() ProviderName {
	return p.name
}

func (p *openAICompatibleProvider) Model() string {
	return p.model
}

func (p *openAICompatibleProvider) Summarize(ctx context.Context, req SummarizeRequest) (SummarizeResponse, error) {
	payload := chatCompletionRequest{
		Model: p.model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		Temperature: 0.2,
		ResponseFmt: responseFormat{Type: "json_object"},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return SummarizeResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return SummarizeResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return SummarizeResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SummarizeResponse{}, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return SummarizeResponse{}, fmt.Errorf("%s provider returned %d: %s", p.name, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return SummarizeResponse{}, err
	}
	if len(completion.Choices) == 0 {
		return SummarizeResponse{}, errors.New("provider returned no choices")
	}

	var result SummarizeResponse
	if err := json.Unmarshal([]byte(completion.Choices[0].Message.Content), &result); err != nil {
		return SummarizeResponse{}, err
	}
	return result, nil
}
