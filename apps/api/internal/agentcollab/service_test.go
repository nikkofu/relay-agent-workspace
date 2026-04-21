package agentcollab

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

const sampleAgentCollabMarkdown = `# Relay Agent Workspace: Team Collaboration Hub

## 📋 Task Board

| Status | Task | Assigned To | Deadline | Description |
| :--- | :--- | :--- | :--- | :--- |
| 🟢 Done | Core API v0.2.0 | Codex | 2026-04-16 | Auth, Workspace, Channel, and Message REST APIs. |
| 🔴 Pending | Agent-Collab Sync Service | Codex | TBD | File watcher and parser for AGENT-COLLAB.md. |

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Gemini** | ` + "`executing-plans`" + ` | v0.2.1 Release & Handover | 100% |
| **Codex** | ` + "`awaiting`" + ` | - | - |
`

func TestParseMarkdownExtractsTaskBoardAndActiveSuperpowers(t *testing.T) {
	snapshot, err := ParseMarkdown([]byte(sampleAgentCollabMarkdown))
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	if len(snapshot.TaskBoard) != 2 {
		t.Fatalf("expected 2 task board rows, got %d", len(snapshot.TaskBoard))
	}
	if snapshot.TaskBoard[1].Task != "Agent-Collab Sync Service" {
		t.Fatalf("unexpected task row: %#v", snapshot.TaskBoard[1])
	}
	if snapshot.TaskBoard[1].AssignedTo != "Codex" {
		t.Fatalf("unexpected assigned_to: %s", snapshot.TaskBoard[1].AssignedTo)
	}

	if len(snapshot.ActiveSuperpowers) != 2 {
		t.Fatalf("expected 2 active superpower rows, got %d", len(snapshot.ActiveSuperpowers))
	}
	if snapshot.ActiveSuperpowers[0].Agent != "Gemini" {
		t.Fatalf("unexpected agent row: %#v", snapshot.ActiveSuperpowers[0])
	}
	if snapshot.ActiveSuperpowers[0].CurrentSkill != "executing-plans" {
		t.Fatalf("unexpected current_skill: %s", snapshot.ActiveSuperpowers[0].CurrentSkill)
	}
}

func TestParseMarkdownReturnsDirectCommTargetsAndToolArrays(t *testing.T) {
	content := `# Relay Agent Workspace: Team Collaboration Hub

## 👥 Member Profiles

| Name | Role | Specialty | Primary Tools |
| :--- | :--- | :--- | :--- |
| **Codex** | API/Backend Agent | Go, Gin | ` + "`apps/api`" + `, ` + "`internal/`" + ` |
| **Windsurf** | Web/UI Agent | TypeScript, UX | ` + "`apps/web`" + `, write_file, multi_edit |

## 📋 Task Board

| Status | Task | Assigned To | Deadline | Description |
| :--- | :--- | :--- | :--- | :--- |
| 🟢 Done | Contract Hardening | Codex | 2026-04-21 | Normalize agent-collab payloads. |

## ⚡️ Active Superpowers (Live State)

| Agent | Current Skill | Active Task | Progress |
| :--- | :--- | :--- | :--- |
| **Codex** | api-architecture | Contract hardening | 100% |

## 💬 Communication Log

### 2026-04-21 - Direct Handoff
- **Windsurf → Codex**: "Please add the to field."
- **Codex**: "Broadcast note."
`

	snapshot, err := ParseMarkdown([]byte(content))
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}

	if len(snapshot.Members) != 2 {
		t.Fatalf("expected 2 members, got %#v", snapshot.Members)
	}
	if got := snapshot.Members[0].PrimaryToolsArray; len(got) != 2 || got[0] != "apps/api" || got[1] != "internal/" {
		t.Fatalf("expected normalized primary tools array, got %#v", got)
	}

	if len(snapshot.CommLog) != 2 {
		t.Fatalf("expected 2 comm log entries, got %#v", snapshot.CommLog)
	}
	if snapshot.CommLog[0].From != "Windsurf" || snapshot.CommLog[0].To != "Codex" {
		t.Fatalf("expected direct comm target, got %#v", snapshot.CommLog[0])
	}
	if snapshot.CommLog[1].From != "Codex" || snapshot.CommLog[1].To != "" {
		t.Fatalf("expected broadcast comm with empty to, got %#v", snapshot.CommLog[1])
	}

	raw, err := json.Marshal(snapshot.CommLog[1])
	if err != nil {
		t.Fatalf("failed to marshal comm log entry: %v", err)
	}
	if !strings.Contains(string(raw), `"to":""`) {
		t.Fatalf("expected broadcast comm log JSON to include empty to field, got %s", raw)
	}
}

func TestSyncFileBroadcastsAgentCollabEvent(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()

	client := realtime.NewTestClient(2)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	dir := t.TempDir()
	path := filepath.Join(dir, "AGENT-COLLAB.md")
	if err := os.WriteFile(path, []byte(sampleAgentCollabMarkdown), 0o644); err != nil {
		t.Fatalf("failed to write markdown fixture: %v", err)
	}

	service := NewService(path, hub)
	if err := service.SyncFile(); err != nil {
		t.Fatalf("SyncFile returned error: %v", err)
	}

	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("did not receive event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}

	if event.Type != "agent_collab.sync" {
		t.Fatalf("unexpected event type: %s", event.Type)
	}
	if event.ChannelID != "ch-collab" {
		t.Fatalf("unexpected channel id: %s", event.ChannelID)
	}

	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("payload was not decoded as map: %#v", event.Payload)
	}
	if len(payload["task_board"].([]any)) != 2 {
		t.Fatalf("unexpected task_board payload: %#v", payload["task_board"])
	}
	if len(payload["active_superpowers"].([]any)) != 2 {
		t.Fatalf("unexpected active_superpowers payload: %#v", payload["active_superpowers"])
	}
}

func TestWatcherBroadcastsAfterFileWrite(t *testing.T) {
	hub := realtime.NewHub()
	go hub.Run()

	client := realtime.NewTestClient(4)
	hub.RegisterTestClient(client)
	defer hub.UnregisterTestClient(client)

	dir := t.TempDir()
	path := filepath.Join(dir, "AGENT-COLLAB.md")
	if err := os.WriteFile(path, []byte(sampleAgentCollabMarkdown), 0o644); err != nil {
		t.Fatalf("failed to write markdown fixture: %v", err)
	}

	service := NewService(path, hub)
	if err := service.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer func() {
		if err := service.Close(); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	}()

	if _, err := client.Receive(2 * time.Second); err != nil {
		t.Fatalf("did not receive initial sync event: %v", err)
	}

	updatedMarkdown := strings.Replace(sampleAgentCollabMarkdown, "100%", "100%+", 1)
	if err := os.WriteFile(path, []byte(updatedMarkdown), 0o644); err != nil {
		t.Fatalf("failed to update markdown fixture: %v", err)
	}

	raw, err := client.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("did not receive watcher sync event: %v", err)
	}

	var event realtime.Event
	if err := json.Unmarshal(raw, &event); err != nil {
		t.Fatalf("failed to decode event: %v", err)
	}

	payload, ok := event.Payload.(map[string]any)
	if !ok {
		t.Fatalf("payload was not decoded as map: %#v", event.Payload)
	}

	activeSuperpowers := payload["active_superpowers"].([]any)
	firstRow := activeSuperpowers[0].(map[string]any)
	if firstRow["progress"] != "100%+" {
		t.Fatalf("watcher did not broadcast updated progress: %#v", firstRow)
	}
}
