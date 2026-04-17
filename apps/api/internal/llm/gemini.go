package llm

import (
	"context"
	"net/http"
	"strings"
)

type GeminiProvider struct {
	client *http.Client
}

func NewGeminiProvider() GeminiProvider {
	return GeminiProvider{client: newHTTPClient()}
}

func (p GeminiProvider) Stream(ctx context.Context, req Request, cfg ProviderConfig) (*StreamSession, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	httpReq, err := newJSONRequest(ctx, http.MethodPost, baseURL+"/models/"+req.Model+":streamGenerateContent?alt=sse", map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": req.Prompt},
				},
			},
		},
	}, map[string]string{
		"x-goog-api-key": cfg.APIKey,
	})
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	session, err := streamSSE(resp, parseGeminiEvent)
	if err != nil {
		return nil, err
	}
	session.Provider = cfg.Name
	session.Model = req.Model
	return session, nil
}
