package llm

import (
	"context"
	"errors"
	"fmt"
)

type Gateway struct {
	config    Config
	providers map[string]Provider
}

func NewGateway(config Config, providers map[string]Provider) *Gateway {
	config.Normalize()
	return &Gateway{
		config:    config,
		providers: providers,
	}
}

func (g *Gateway) Stream(ctx context.Context, req Request) (*StreamSession, error) {
	providerName := req.Provider
	if providerName == "" {
		providerName = g.config.DefaultProvider
	}

	providerConfig, ok := g.config.Providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not configured: %s", providerName)
	}
	if !providerConfig.Enabled {
		return nil, fmt.Errorf("provider disabled: %s", providerName)
	}

	provider, ok := g.providers[providerName]
	if !ok {
		provider, ok = g.providers[providerConfig.Kind]
	}
	if !ok {
		return nil, fmt.Errorf("provider implementation not found: %s", providerName)
	}

	if providerConfig.APIKey == "" {
		return nil, errors.New("provider api key is empty")
	}
	if req.Model == "" {
		req.Model = providerConfig.Model
	}

	return provider.Stream(ctx, req, providerConfig)
}
