package llm

import "context"

type OpenRouterProvider struct {
	OpenAICompatibleProvider
}

func NewOpenRouterProvider() OpenRouterProvider {
	return OpenRouterProvider{OpenAICompatibleProvider: NewOpenAICompatibleProvider()}
}

func (p OpenRouterProvider) Stream(ctx context.Context, req Request, cfg ProviderConfig) (*StreamSession, error) {
	if cfg.APIStyle == "" {
		cfg.APIStyle = "responses"
	}
	return p.OpenAICompatibleProvider.Stream(ctx, req, cfg)
}
