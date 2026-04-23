# Phase 63D AI Compose DM And Intent Expansion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add DM parity, intent variants, and feedback summary APIs to the AI composer.

**Architecture:** Keep the existing channel compose implementation as the default path, add scope-aware helpers for DM conversations, and preserve the existing compose response shape with additive `dm_id` fields. Feedback summary uses the existing `AIComposeFeedback` persistence model.

**Tech Stack:** Go, Gin, GORM, SQLite, existing LLM gateway, existing handler test harness, pnpm lint/tsc for repo verification.

---

### Task 1: Lock Phase 63D API Contracts

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add a failing test for `POST /api/v1/ai/compose` with `dm_id`, `intent=followup`, and recent DM messages.
- [ ] Add a failing test for `POST /api/v1/ai/compose/stream` with `dm_id`, `intent=summarize`, and SSE payload checks.
- [ ] Add a failing test that all intents `reply|summarize|followup|schedule` are accepted and `intent=unknown` returns `400`.
- [ ] Add a failing test for DM-scoped `POST /api/v1/ai/compose/:id/feedback`.
- [ ] Add a failing test for `GET /api/v1/ai/compose/:id/feedback/summary`.

### Task 2: Implement Scope-Aware Compose

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/main.go`

- [ ] Extend `composeInput` with `DMID`.
- [ ] Update binding validation to require exactly one of `channel_id` or `dm_id`.
- [ ] Add `isSupportedComposeIntent`.
- [ ] Add DM recent-message loading and prompt construction.
- [ ] Add `dm_id` to `composePayload` and SSE `start` / `suggestion.done` events.
- [ ] Register `GET /api/v1/ai/compose/:id/feedback/summary`.

### Task 3: Extend Feedback Persistence Contract

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/handlers/ai.go`

- [ ] Add `DMConversationID` JSON field to `AIComposeFeedback`.
- [ ] Allow feedback requests to target exactly one of `channel_id` or `dm_id`.
- [ ] Validate channel or DM existence.
- [ ] Persist the selected scope.
- [ ] Implement summary aggregation by `compose_id`.

### Task 4: Release And Collaboration Docs

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Add: `docs/releases/v0.6.18.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`

- [ ] Document the new APIs and request/response expectations.
- [ ] Leave Windsurf a concrete UI handoff for enabling DM AI Suggest and intent selector UI.
- [ ] Bump versions to `0.6.18`.

### Task 5: Verification And Release

- [ ] Run focused handler tests for Phase 63D.
- [ ] Run `go test ./...`.
- [ ] Run `GOCACHE=$(pwd)/.cache/go-build go build ./...`.
- [ ] Run web lint and TypeScript checks.
- [ ] Commit, tag `v0.6.18`, and push `main` plus the tag.
