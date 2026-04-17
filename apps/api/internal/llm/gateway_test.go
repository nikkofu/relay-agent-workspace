package llm

import (
	"context"
	"testing"
)

type stubProvider struct{}

func (stubProvider) Stream(_ context.Context, req Request, cfg ProviderConfig) (*StreamSession, error) {
	events := make(chan StreamEvent, 1)
	errs := make(chan error, 1)
	events <- StreamEvent{Type: "chunk", Text: req.Prompt}
	close(events)
	close(errs)

	return &StreamSession{
		Provider: cfg.Name,
		Model:    cfg.Model,
		Events:   events,
		Errors:   errs,
	}, nil
}

func TestGatewayUsesDefaultProviderWhenRequestDoesNotOverride(t *testing.T) {
	gateway := NewGateway(Config{
		DefaultProvider: "gemini",
		Providers: map[string]ProviderConfig{
			"gemini": {
				Name:    "gemini",
				Kind:    "gemini",
				Enabled: true,
				APIKey:  "test-key",
				Model:   "gemini-2.5-flash",
			},
		},
	}, map[string]Provider{
		"gemini": stubProvider{},
	})

	session, err := gateway.Stream(context.Background(), Request{Prompt: "hello"})
	if err != nil {
		t.Fatalf("Stream returned error: %v", err)
	}

	if session.Provider != "gemini" {
		t.Fatalf("expected default provider gemini, got %s", session.Provider)
	}
	if session.Model != "gemini-2.5-flash" {
		t.Fatalf("expected configured model, got %s", session.Model)
	}
}
