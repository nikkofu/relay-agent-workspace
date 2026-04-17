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

func streamSSE(resp *http.Response, parse func(event string, data []byte) []StreamEvent) (*StreamSession, error) {
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

			parsedEvents := parse(eventName, []byte(data))
			for _, parsed := range parsedEvents {
				if parsed.Text == "" {
					continue
				}
				if parsed.Type == "" {
					parsed.Type = "chunk"
				}
				events <- parsed
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

func parseOpenAIResponsesEvent(eventName string, data []byte) []StreamEvent {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}

	var events []StreamEvent
	appendEvent := func(eventType string, text string) {
		if text == "" {
			return
		}
		events = append(events, StreamEvent{Type: eventType, Text: text})
	}

	isReasoningEvent := strings.Contains(eventName, "reasoning")
	if delta, ok := payload["delta"].(string); ok {
		if isReasoningEvent {
			appendEvent("reasoning", delta)
		} else {
			appendEvent("chunk", delta)
		}
	}

	if output, ok := payload["output"].([]any); ok {
		for _, item := range output {
			obj, ok := item.(map[string]any)
			if !ok {
				continue
			}
			itemType, _ := obj["type"].(string)
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
					if partType, _ := part["type"].(string); strings.Contains(partType, "reasoning") || strings.Contains(itemType, "reasoning") {
						appendEvent("reasoning", text)
					} else {
						appendEvent("chunk", text)
					}
				}
				if summary, ok := part["summary"].(string); ok {
					appendEvent("reasoning", summary)
				}
			}
		}
	}

	if outputText, ok := payload["output_text"].(string); ok {
		appendEvent("chunk", outputText)
	}

	if reasoning, ok := payload["reasoning"].(map[string]any); ok {
		if text, ok := reasoning["text"].(string); ok {
			appendEvent("reasoning", text)
		}
		if summary, ok := reasoning["summary"].(string); ok {
			appendEvent("reasoning", summary)
		}
	}

	return events
}

func parseChatCompletionsEvent(_ string, data []byte) []StreamEvent {
	var payload struct {
		Choices []struct {
			Delta struct {
				Content   string `json:"content"`
				Reasoning string `json:"reasoning"`
			} `json:"delta"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}
	if len(payload.Choices) == 0 {
		return nil
	}

	var events []StreamEvent
	if payload.Choices[0].Delta.Reasoning != "" {
		events = append(events, StreamEvent{Type: "reasoning", Text: payload.Choices[0].Delta.Reasoning})
	}
	if payload.Choices[0].Delta.Content != "" {
		events = append(events, StreamEvent{Type: "chunk", Text: payload.Choices[0].Delta.Content})
	}
	return events
}

func parseGeminiEvent(_ string, data []byte) []StreamEvent {
	var payload struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text    string `json:"text"`
					Thought bool   `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return nil
	}

	var events []StreamEvent
	for _, candidate := range payload.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				if part.Thought {
					events = append(events, StreamEvent{Type: "reasoning", Text: part.Text})
				} else {
					events = append(events, StreamEvent{Type: "chunk", Text: part.Text})
				}
			}
		}
	}
	return events
}
