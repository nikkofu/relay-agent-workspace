# Phase 63F Always-On Knowledge Automation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add channel auto-summarize contracts and compose collaboration signals for the next Windsurf UI pass.

**Architecture:** Reuse the existing channel `AISummary` cache and realtime event hub. Add only one new settings model, then keep the run path in the existing AI handler layer because it reuses channel summary generation and LLM gateway plumbing.

**Tech Stack:** Go, Gin, GORM, SQLite, project realtime hub, existing LLM gateway.

---

### Task 1: Contract Test

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [x] Add `TestPhase63FAutoSummarizeAndComposeRealtimeContracts`.
- [x] Run `go test ./internal/handlers -run TestPhase63F -count=1`.
- [x] Verify RED from missing handlers/model.

### Task 2: Auto-Summarize Model And Routes

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [x] Add `ChannelAutoSummarySetting`.
- [x] Add it to app and test migrations.
- [x] Register `GET|PUT|POST /channels/:id/knowledge/auto-summarize`.

### Task 3: Auto-Summarize Handlers

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`

- [x] Add setting read/write handlers.
- [x] Add run handler that generates a channel summary with provider/model options.
- [x] Emit `channel.summary.updated`.

### Task 4: Compose Collaboration Signals

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/knowledge/types.go`

- [x] Add `ComposeProposedSlot`.
- [x] Add additive `proposed_slots[]` to compose payloads.
- [x] Emit `knowledge.compose.suggestion.generated` for sync and streaming compose.

### Task 5: Documentation And Release

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Create: `docs/releases/v0.6.22.md`

- [x] Document APIs, websocket events, and Windsurf handoff.
- [x] Run full verification.
- [x] Commit, tag, and push `v0.6.22`.
