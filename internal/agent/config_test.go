package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigRequiresApiKeyForEnabledProvider(t *testing.T) {
	t.Setenv("LLM_PROVIDER_DEFAULT", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "")

	_, err := LoadConfigFromEnv("")
	if err == nil {
		t.Fatalf("expected missing api key error")
	}
}

func TestLoadConfigReadsDotEnvWhenProcessEnvMissing(t *testing.T) {
	root := t.TempDir()
	content := "LLM_PROVIDER_DEFAULT=modelscope\nMODELSCOPE_API_KEY=test-key\nMODELSCOPE_MODEL=qwen-max\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	cfg, err := LoadConfigFromEnv(root)
	if err != nil {
		t.Fatalf("LoadConfigFromEnv returned error: %v", err)
	}
	if cfg.DefaultProvider != ProviderModelScope {
		t.Fatalf("unexpected provider: %q", cfg.DefaultProvider)
	}
	if cfg.ModelScope.APIKey != "test-key" {
		t.Fatalf("unexpected api key: %q", cfg.ModelScope.APIKey)
	}
	if cfg.ModelScope.Model != "qwen-max" {
		t.Fatalf("unexpected model: %q", cfg.ModelScope.Model)
	}
}

func TestLoadConfigPrefersEarlierConfigRoot(t *testing.T) {
	firstRoot := t.TempDir()
	secondRoot := t.TempDir()

	if err := os.WriteFile(filepath.Join(firstRoot, ".env"), []byte("LLM_PROVIDER_DEFAULT=deepseek\nDEEPSEEK_API_KEY=first-key\n"), 0o600); err != nil {
		t.Fatalf("failed to write first .env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secondRoot, ".env"), []byte("LLM_PROVIDER_DEFAULT=deepseek\nDEEPSEEK_API_KEY=second-key\n"), 0o600); err != nil {
		t.Fatalf("failed to write second .env: %v", err)
	}

	cfg, err := LoadConfigFromEnv(firstRoot, secondRoot)
	if err != nil {
		t.Fatalf("LoadConfigFromEnv returned error: %v", err)
	}
	if cfg.DeepSeek.APIKey != "first-key" {
		t.Fatalf("expected first config root to win, got %q", cfg.DeepSeek.APIKey)
	}
}
