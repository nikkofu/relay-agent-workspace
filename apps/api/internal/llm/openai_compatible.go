package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type OpenAICompatibleProvider struct {
	client *http.Client
}

func NewOpenAICompatibleProvider() OpenAICompatibleProvider {
	return OpenAICompatibleProvider{client: newHTTPClient()}
}

func (p OpenAICompatibleProvider) Stream(ctx context.Context, req Request, cfg ProviderConfig) (*StreamSession, error) {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.APIKey,
	}
	for key, value := range cfg.Headers {
		headers[key] = value
	}

	var (
		httpReq *http.Request
		err     error
	)

	switch cfg.APIStyle {
	case "", "responses":
		httpReq, err = newJSONRequest(ctx, http.MethodPost, baseURL+"/responses", map[string]any{
			"model":  req.Model,
			"input":  req.Prompt,
			"stream": true,
		}, headers)
		if err != nil {
			return nil, err
		}
		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, err
		}
		session, err := streamSSE(resp, parseOpenAIResponsesEvent)
		if err != nil {
			return nil, err
		}
		session.Provider = cfg.Name
		session.Model = req.Model
		return session, nil
	case "chat_completions":
		httpReq, err = newJSONRequest(ctx, http.MethodPost, baseURL+"/chat/completions", map[string]any{
			"model": req.Model,
			"messages": []map[string]string{
				{"role": "user", "content": req.Prompt},
			},
			"stream": true,
		}, headers)
		if err != nil {
			return nil, err
		}
		resp, err := p.client.Do(httpReq)
		if err != nil {
			return nil, err
		}
		session, err := streamSSE(resp, parseChatCompletionsEvent)
		if err != nil {
			return nil, err
		}
		session.Provider = cfg.Name
		session.Model = req.Model
		return session, nil
	default:
		return nil, fmt.Errorf("unsupported api_style: %s", cfg.APIStyle)
	}
}
