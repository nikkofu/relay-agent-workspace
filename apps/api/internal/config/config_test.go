package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLLMConfigMergesFilesAndEnvOverrides(t *testing.T) {
	dir := t.TempDir()

	base := `default_provider: openrouter
providers:
  openrouter:
    enabled: true
    kind: openrouter
    base_url: https://openrouter.ai/api/v1
    model: openai/gpt-4o-mini
    api_key: base-key
`
	local := `providers:
  openrouter:
    model: openai/gpt-4.1-mini
  gemini:
    enabled: true
    kind: gemini
    model: gemini-2.5-flash
`
	secrets := `providers:
  gemini:
    api_key: secret-gemini-key
`

	writeTestFile(t, filepath.Join(dir, "llm.base.yaml"), base)
	writeTestFile(t, filepath.Join(dir, "llm.local.yaml"), local)
	writeTestFile(t, filepath.Join(dir, "llm.secrets.local.yaml"), secrets)

	t.Setenv("LLM_DEFAULT_PROVIDER", "gemini")
	t.Setenv("LLM_PROVIDER_GEMINI_BASE_URL", "https://generativelanguage.googleapis.com/v1beta")

	cfg, err := LoadLLMConfig(dir)
	if err != nil {
		t.Fatalf("LoadLLMConfig returned error: %v", err)
	}

	if cfg.DefaultProvider != "gemini" {
		t.Fatalf("expected env override default provider, got %s", cfg.DefaultProvider)
	}
	if cfg.Providers["openrouter"].Model != "openai/gpt-4.1-mini" {
		t.Fatalf("expected local override for openrouter model, got %s", cfg.Providers["openrouter"].Model)
	}
	if !cfg.Providers["openrouter"].Enabled {
		t.Fatalf("expected openrouter enabled flag to be preserved across layered merge")
	}
	if cfg.Providers["gemini"].APIKey != "secret-gemini-key" {
		t.Fatalf("expected secrets override for gemini api key, got %s", cfg.Providers["gemini"].APIKey)
	}
	if cfg.Providers["gemini"].Kind != "gemini" {
		t.Fatalf("expected gemini kind to survive merge, got %s", cfg.Providers["gemini"].Kind)
	}
	if cfg.Providers["gemini"].BaseURL != "https://generativelanguage.googleapis.com/v1beta" {
		t.Fatalf("expected env override base url, got %s", cfg.Providers["gemini"].BaseURL)
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
