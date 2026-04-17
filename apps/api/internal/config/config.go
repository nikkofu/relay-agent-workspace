package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/nikkofu/relay-agent-workspace/api/internal/llm"
)

func LoadLLMConfig(dir string) (llm.Config, error) {
	cfg := llm.DefaultConfig()

	files := []string{
		filepath.Join(dir, "llm.base.yaml"),
		filepath.Join(dir, "llm.local.yaml"),
		filepath.Join(dir, "llm.secrets.local.yaml"),
	}

	for _, path := range files {
		if err := mergeConfigFile(path, &cfg); err != nil {
			return llm.Config{}, err
		}
	}

	applyEnvOverrides(&cfg)
	cfg.Normalize()
	return cfg, nil
}

func mergeConfigFile(path string, cfg *llm.Config) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return yaml.Unmarshal(content, cfg)
}

func applyEnvOverrides(cfg *llm.Config) {
	if value := os.Getenv("LLM_DEFAULT_PROVIDER"); value != "" {
		cfg.DefaultProvider = value
	}

	for name, provider := range cfg.Providers {
		prefix := "LLM_PROVIDER_" + normalizeProviderEnvName(name) + "_"
		if value := os.Getenv(prefix + "API_KEY"); value != "" {
			provider.APIKey = value
		}
		if value := os.Getenv(prefix + "BASE_URL"); value != "" {
			provider.BaseURL = value
		}
		if value := os.Getenv(prefix + "MODEL"); value != "" {
			provider.Model = value
		}
		if value := os.Getenv(prefix + "API_STYLE"); value != "" {
			provider.APIStyle = value
		}
		if value := os.Getenv(prefix + "ENABLED"); value != "" {
			if parsed, err := strconv.ParseBool(value); err == nil {
				provider.Enabled = parsed
			}
		}
		cfg.Providers[name] = provider
	}
}

func normalizeProviderEnvName(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		if r == '-' {
			out = append(out, '_')
			continue
		}
		if r >= 'a' && r <= 'z' {
			out = append(out, r-32)
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
