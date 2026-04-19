package agent

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/unicode"
)

func clearAgentEnv(t *testing.T) {
	t.Helper()
	t.Setenv("LLM_PROVIDER_DEFAULT", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("DEEPSEEK_BASE_URL", "")
	t.Setenv("DEEPSEEK_MODEL", "")
	t.Setenv("MODELSCOPE_API_KEY", "")
	t.Setenv("MODELSCOPE_BASE_URL", "")
	t.Setenv("MODELSCOPE_MODEL", "")
}

func TestLoadConfigRequiresApiKeyForEnabledProvider(t *testing.T) {
	clearAgentEnv(t)
	t.Setenv("LLM_PROVIDER_DEFAULT", "deepseek")
	t.Setenv("DEEPSEEK_API_KEY", "")

	_, err := LoadConfigFromEnv("")
	if err == nil {
		t.Fatalf("expected missing api key error")
	}
}

func TestLoadConfigReadsDotEnvWhenProcessEnvMissing(t *testing.T) {
	clearAgentEnv(t)
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
	clearAgentEnv(t)
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

func TestLoadConfigFindsDotEnvInParentDirectory(t *testing.T) {
	clearAgentEnv(t)
	root := t.TempDir()
	child := filepath.Join(root, "build", "bin")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("failed to create child dir: %v", err)
	}

	content := "LLM_PROVIDER_DEFAULT=deepseek\nDEEPSEEK_API_KEY=parent-key\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write parent .env: %v", err)
	}

	cfg, err := LoadConfigFromEnv(child)
	if err != nil {
		t.Fatalf("LoadConfigFromEnv returned error: %v", err)
	}
	if cfg.DefaultProvider != ProviderDeepSeek {
		t.Fatalf("unexpected provider: %q", cfg.DefaultProvider)
	}
	if cfg.DeepSeek.APIKey != "parent-key" {
		t.Fatalf("expected parent config to be loaded, got %q", cfg.DeepSeek.APIKey)
	}
}

func TestLoadConfigReadsUTF16DotEnv(t *testing.T) {
	clearAgentEnv(t)
	root := t.TempDir()
	content := "LLM_PROVIDER_DEFAULT=deepseek\nDEEPSEEK_API_KEY=utf16-key\nDEEPSEEK_MODEL=deepseek-chat\n"
	encoded, err := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder().Bytes([]byte(content))
	if err != nil {
		t.Fatalf("failed to encode .env as utf16: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), encoded, 0o600); err != nil {
		t.Fatalf("failed to write utf16 .env: %v", err)
	}

	cfg, err := LoadConfigFromEnv(root)
	if err != nil {
		t.Fatalf("LoadConfigFromEnv returned error: %v", err)
	}
	if cfg.DefaultProvider != ProviderDeepSeek {
		t.Fatalf("unexpected provider: %q", cfg.DefaultProvider)
	}
	if cfg.DeepSeek.APIKey != "utf16-key" {
		t.Fatalf("expected utf16 api key to be loaded, got %q", cfg.DeepSeek.APIKey)
	}
}
