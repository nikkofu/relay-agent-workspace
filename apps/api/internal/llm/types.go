package llm

import "context"

type Request struct {
	Prompt         string `json:"prompt"`
	ChannelID      string `json:"channel_id"`
	ConversationID string `json:"conversation_id,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Model          string `json:"model,omitempty"`
}

type StreamEvent struct {
	Type string
	Text string
}

type StreamSession struct {
	Provider string
	Model    string
	Events   <-chan StreamEvent
	Errors   <-chan error
}

type Provider interface {
	Stream(ctx context.Context, req Request, cfg ProviderConfig) (*StreamSession, error)
}

type ProviderConfig struct {
	Name            string            `yaml:"-"`
	Enabled         bool              `yaml:"enabled"`
	Kind            string            `yaml:"kind"`
	BaseURL         string            `yaml:"base_url"`
	APIKey          string            `yaml:"api_key"`
	Model           string            `yaml:"model"`
	APIStyle        string            `yaml:"api_style"`
	Temperature     float64           `yaml:"temperature"`
	MaxOutputTokens int               `yaml:"max_output_tokens"`
	Headers         map[string]string `yaml:"headers"`
}

type Config struct {
	DefaultProvider string                    `yaml:"default_provider"`
	Providers       map[string]ProviderConfig `yaml:"providers"`
}

func DefaultConfig() Config {
	return Config{
		DefaultProvider: "gemini",
		Providers: map[string]ProviderConfig{
			"openai": {
				Name:     "openai",
				Kind:     "openai",
				BaseURL:  "https://api.openai.com/v1",
				Model:    "gpt-4.1-mini",
				APIStyle: "responses",
				Headers:  map[string]string{},
			},
			"openai-compatible": {
				Name:     "openai-compatible",
				Kind:     "openai-compatible",
				BaseURL:  "https://api.openai.com/v1",
				Model:    "gpt-4.1-mini",
				APIStyle: "responses",
				Headers:  map[string]string{},
			},
			"openrouter": {
				Name:     "openrouter",
				Kind:     "openrouter",
				BaseURL:  "https://openrouter.ai/api/v1",
				Model:    "openai/gpt-4.1-mini",
				APIStyle: "responses",
				Headers:  map[string]string{},
			},
			"gemini": {
				Name:     "gemini",
				Kind:     "gemini",
				BaseURL:  "https://generativelanguage.googleapis.com/v1beta",
				Model:    "gemini-2.5-flash",
				APIStyle: "native",
				Headers:  map[string]string{},
			},
		},
	}
}

func (c *Config) Normalize() {
	for name, provider := range c.Providers {
		provider.Name = name
		if provider.Headers == nil {
			provider.Headers = map[string]string{}
		}
		c.Providers[name] = provider
	}
}
