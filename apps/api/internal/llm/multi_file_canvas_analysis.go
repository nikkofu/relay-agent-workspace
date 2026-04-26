package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type MultiFileAnalysisResult struct {
	Summary                string `json:"summary"`
	DefaultExecutionTarget *struct {
		Type string `json:"type"`
	} `json:"default_execution_target,omitempty"`
	Observations []string `json:"observations"`
	NextSteps    []struct {
		Text            string `json:"text"`
		Rationale       string `json:"rationale"`
		ActionHint      string `json:"action_hint"`
		ExecutionTarget *struct {
			Type string `json:"type"`
		} `json:"execution_target,omitempty"`
	} `json:"next_steps"`
}

func AnalyzeFileGroup(ctx context.Context, gateway Provider, cfg Config, files []FileAnalysisInput) (*MultiFileAnalysisResult, error) {
	prompt := BuildAnalysisPrompt(files)
	
	req := Request{
		Prompt: prompt,
	}

	providerCfg := cfg.Providers[cfg.DefaultProvider]
	
	session, err := gateway.Stream(ctx, req, providerCfg)
	if err != nil {
		return nil, err
	}

	var fullText strings.Builder
	for {
		select {
		case event, ok := <-session.Events:
			if !ok {
				return ParseAnalysisResult(fullText.String())
			}
			fullText.WriteString(event.Text)
		case err := <-session.Errors:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

type FileAnalysisInput struct {
	ID          string
	Name        string
	ContentType string
	ContentText string
	Summary     string
}

func BuildAnalysisPrompt(files []FileAnalysisInput) string {
	var b strings.Builder
	b.WriteString("Analyze the following group of files and provide a structured JSON response.\n\n")
	
	for _, f := range files {
		b.WriteString(fmt.Sprintf("FILE: %s (%s)\n", f.Name, f.ID))
		if f.ContentText != "" {
			b.WriteString(fmt.Sprintf("CONTENT:\n%s\n", f.ContentText))
		} else if f.Summary != "" {
			b.WriteString(fmt.Sprintf("SUMMARY: %s\n", f.Summary))
		} else {
			b.WriteString("CONTENT: (Not available)\n")
		}
		b.WriteString("---\n")
	}

	b.WriteString("\nResponse MUST be a single JSON object with the following schema:\n")
	b.WriteString("{\n")
	b.WriteString("  \"summary\": \"Overall merged summary\",\n")
	b.WriteString("  \"default_execution_target\": {\"type\": \"list|workflow|channel_message\"},\n")
	b.WriteString("  \"observations\": [\"concise finding 1\", \"concise finding 2\"],\n")
	b.WriteString("  \"next_steps\": [\n")
	b.WriteString("    {\n")
	b.WriteString("      \"text\": \"Step text\",\n")
	b.WriteString("      \"rationale\": \"Short why\",\n")
	b.WriteString("      \"action_hint\": \"summarize|compare|decide|share|plan|investigate|custom\",\n")
	b.WriteString("      \"execution_target\": {\"type\": \"list|workflow|channel_message\"}\n")
	b.WriteString("    }\n")
	b.WriteString("  ]\n")
	b.WriteString("}\n")
	
	return b.String()
}

func ParseAnalysisResult(raw string) (*MultiFileAnalysisResult, error) {
	// Simple JSON extraction in case AI wraps it in markdown
	jsonStr := raw
	if start := strings.Index(raw, "{"); start != -1 {
		if end := strings.LastIndex(raw, "}"); end != -1 && end > start {
			jsonStr = raw[start : end+1]
		}
	}

	var result MultiFileAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// Fallback for malformed AI output
		return &MultiFileAnalysisResult{
			Summary:      "Failed to parse AI output. Raw: " + raw,
			Observations: []string{"Error: " + err.Error()},
		}, nil
	}

	// Phase 70B: Harden execution targets
	isValid := func(t string) bool {
		switch t {
		case "list", "workflow", "channel_message":
			return true
		}
		return false
	}

	if result.DefaultExecutionTarget != nil && !isValid(result.DefaultExecutionTarget.Type) {
		result.DefaultExecutionTarget = nil
	}

	for i := range result.NextSteps {
		if result.NextSteps[i].ExecutionTarget != nil && !isValid(result.NextSteps[i].ExecutionTarget.Type) {
			result.NextSteps[i].ExecutionTarget = nil
		}
	}

	return &result, nil
}
