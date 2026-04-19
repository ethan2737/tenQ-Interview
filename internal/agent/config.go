package agent

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/unicode"
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

func LoadConfigFromEnv(rootDirs ...string) (Config, error) {
	for _, rootDir := range uniqueConfigRoots(expandConfigRoots(rootDirs...)...) {
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

func uniqueConfigRoots(rootDirs ...string) []string {
	seen := make(map[string]struct{}, len(rootDirs))
	roots := make([]string, 0, len(rootDirs))
	for _, rootDir := range rootDirs {
		rootDir = strings.TrimSpace(rootDir)
		if rootDir == "" {
			continue
		}
		if _, ok := seen[rootDir]; ok {
			continue
		}
		seen[rootDir] = struct{}{}
		roots = append(roots, rootDir)
	}
	return roots
}

func expandConfigRoots(rootDirs ...string) []string {
	expanded := make([]string, 0, len(rootDirs)*4)
	for _, rootDir := range rootDirs {
		rootDir = strings.TrimSpace(rootDir)
		if rootDir == "" {
			continue
		}

		current := rootDir
		for {
			expanded = append(expanded, current)
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	return expanded
}

func loadDotEnv(path string) error {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	decoded, err := decodeDotEnvBytes(raw)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(strings.NewReader(decoded))
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

func decodeDotEnvBytes(raw []byte) (string, error) {
	switch {
	case bytes.HasPrefix(raw, []byte{0xEF, 0xBB, 0xBF}):
		return string(raw[3:]), nil
	case bytes.HasPrefix(raw, []byte{0xFF, 0xFE}):
		decoded, err := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Bytes(raw)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	case bytes.HasPrefix(raw, []byte{0xFE, 0xFF}):
		decoded, err := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder().Bytes(raw)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	default:
		return string(raw), nil
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
