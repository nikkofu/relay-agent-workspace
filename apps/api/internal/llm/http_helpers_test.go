package llm

import "testing"

func TestParseOpenAIResponsesEventExtractsReasoningAndChunks(t *testing.T) {
	events := parseOpenAIResponsesEvent("response.reasoning.delta", []byte(`{"delta":"Thinking step"}`))
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "reasoning" || events[0].Text != "Thinking step" {
		t.Fatalf("unexpected reasoning event: %#v", events[0])
	}
}

func TestParseGeminiEventExtractsThoughtPartsAsReasoning(t *testing.T) {
	events := parseGeminiEvent("", []byte(`{"candidates":[{"content":{"parts":[{"text":"internal thought","thought":true},{"text":"final answer"}]}}]}`))
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != "reasoning" || events[0].Text != "internal thought" {
		t.Fatalf("unexpected first event: %#v", events[0])
	}
	if events[1].Type != "chunk" || events[1].Text != "final answer" {
		t.Fatalf("unexpected second event: %#v", events[1])
	}
}
