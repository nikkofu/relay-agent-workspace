package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
	normalized := strings.ReplaceAll(string(content), "\t", "  ")
	var overlay llm.Config
	if err := yaml.Unmarshal([]byte(normalized), &overlay); err != nil {
		return err
	}

	mergeConfig(cfg, overlay)
	return nil
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

func mergeConfig(dst *llm.Config, src llm.Config) {
	if src.DefaultProvider != "" {
		dst.DefaultProvider = src.DefaultProvider
	}
	if dst.Providers == nil {
		dst.Providers = map[string]llm.ProviderConfig{}
	}

	for name, provider := range src.Providers {
		merged := dst.Providers[name]
		if provider.Name != "" {
			merged.Name = provider.Name
		}
		if provider.Kind != "" {
			merged.Kind = provider.Kind
		}
		if provider.BaseURL != "" {
			merged.BaseURL = provider.BaseURL
		}
		if provider.APIKey != "" {
			merged.APIKey = provider.APIKey
		}
		if provider.Model != "" {
			merged.Model = provider.Model
		}
		if provider.APIStyle != "" {
			merged.APIStyle = provider.APIStyle
		}
		if provider.Enabled {
			merged.Enabled = true
		}
		if provider.Temperature != 0 {
			merged.Temperature = provider.Temperature
		}
		if provider.MaxOutputTokens != 0 {
			merged.MaxOutputTokens = provider.MaxOutputTokens
		}
		if len(provider.Headers) > 0 {
			if merged.Headers == nil {
				merged.Headers = map[string]string{}
			}
			for key, value := range provider.Headers {
				merged.Headers[key] = value
			}
		}
		dst.Providers[name] = merged
	}
}
