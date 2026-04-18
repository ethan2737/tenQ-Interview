package agent

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDeepSeekBaseURL   = "https://api.deepseek.com"
	defaultDeepSeekModel     = "deepseek-chat"
	defaultModelScopeBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultModelScopeModel   = "qwen-plus"
)

type ProviderConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Config struct {
	DefaultProvider ProviderName
	DeepSeek        ProviderConfig
	ModelScope      ProviderConfig
}

func LoadConfigFromEnv(rootDir string) (Config, error) {
	if rootDir != "" {
		if err := loadDotEnv(filepath.Join(rootDir, ".env")); err != nil {
			return Config{}, err
		}
	}

	cfg := Config{
		DefaultProvider: ProviderName(strings.TrimSpace(os.Getenv("LLM_PROVIDER_DEFAULT"))),
		DeepSeek: ProviderConfig{
			BaseURL: envOrDefault("DEEPSEEK_BASE_URL", defaultDeepSeekBaseURL),
			APIKey:  strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")),
			Model:   envOrDefault("DEEPSEEK_MODEL", defaultDeepSeekModel),
		},
		ModelScope: ProviderConfig{
			BaseURL: envOrDefault("MODELSCOPE_BASE_URL", defaultModelScopeBaseURL),
			APIKey:  strings.TrimSpace(os.Getenv("MODELSCOPE_API_KEY")),
			Model:   envOrDefault("MODELSCOPE_MODEL", defaultModelScopeModel),
		},
	}

	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = ProviderDeepSeek
	}

	switch cfg.DefaultProvider {
	case ProviderDeepSeek:
		if cfg.DeepSeek.APIKey == "" {
			return Config{}, errors.New("deepseek api key is required")
		}
	case ProviderModelScope:
		if cfg.ModelScope.APIKey == "" {
			return Config{}, errors.New("modelscope api key is required")
		}
	default:
		return Config{}, errors.New("unsupported llm provider")
	}

	return cfg, nil
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
