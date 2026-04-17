package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func newHTTPClient() *http.Client {
	return &http.Client{}
}

func newJSONRequest(ctx context.Context, method string, url string, body any, headers map[string]string) (*http.Request, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func streamSSE(resp *http.Response, parse func(event string, data []byte) []string) (*StreamSession, error) {
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("upstream returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	events := make(chan StreamEvent, 32)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 1024), 1024*1024)

		var eventName string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event:") {
				eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}

			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "" {
				continue
			}
			if data == "[DONE]" {
				return
			}

			texts := parse(eventName, []byte(data))
			for _, text := range texts {
				if text == "" {
					continue
				}
				events <- StreamEvent{Type: "chunk", Text: text}
			}
			eventName = ""
		}

		if err := scanner.Err(); err != nil {
			errs <- err
		}
	}()

	return &StreamSession{
		Events: events,
		Errors: errs,
	}, nil
}

func parseOpenAIResponsesEvent(_ string, data []byte) []string {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}

	var texts []string
	if delta, ok := payload["delta"].(string); ok {
		texts = append(texts, delta)
	}

	if output, ok := payload["output"].([]any); ok {
		for _, item := range output {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			content, ok := obj["content"].([]any)
			if !ok {
				continue
			}
			for _, c := range content {
				part, ok := c.(map[string]any)
				if !ok {
					continue
				}
				if text, ok := part["text"].(string); ok {
					texts = append(texts, text)
				}
			}
		}
	}

	if outputText, ok := payload["output_text"].(string); ok {
		texts = append(texts, outputText)
	}

	return texts
}

func parseChatCompletionsEvent(_ string, data []byte) []string {
	var payload struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	if len(payload.Choices) == 0 {
		return nil
	}
	if payload.Choices[0].Delta.Content == "" {
		return nil
	}
	return []string{payload.Choices[0].Delta.Content}
}

func parseGeminiEvent(_ string, data []byte) []string {
	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}

	var texts []string
	for _, candidate := range payload.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				texts = append(texts, part.Text)
			}
		}
	}
	return texts
}
