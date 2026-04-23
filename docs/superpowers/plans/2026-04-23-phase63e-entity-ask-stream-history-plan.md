# Phase 63E Entity Ask Stream And History Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add persistent history and SSE streaming for entity-scoped Ask AI.

**Architecture:** Reuse the existing entity ask prompt builder and LLM gateway. Add one domain persistence model for Q&A history, thin handler helpers for persist/list, and two new routes under `/knowledge/entities/:id/ask`.

**Tech Stack:** Go, Gin, GORM, SQLite, existing LLM gateway, existing SSE helpers, existing handler test harness.

---

### Task 1: Lock Contracts With Tests

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add a failing test that sync `POST /knowledge/entities/:id/ask` persists a history row.
- [ ] Add a failing test for `GET /knowledge/entities/:id/ask/history` returning newest-first current-user rows.
- [ ] Add a failing test for `POST /knowledge/entities/:id/ask/stream` SSE events and final persistence.

### Task 2: Add Persistence Model And Migration

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go` test migrations

- [ ] Add `KnowledgeEntityAskAnswer`.
- [ ] Auto-migrate it in production DB init.
- [ ] Auto-migrate it in handler test DB setup.

### Task 3: Implement Handlers

**Files:**
- Modify: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/main.go`

- [ ] Extract common ask request parsing and prompt preparation.
- [ ] Persist answers for sync ask.
- [ ] Implement `GetKnowledgeEntityAskHistory`.
- [ ] Implement `AskKnowledgeEntityStream` using existing `writeSSE`.
- [ ] Register `GET /knowledge/entities/:id/ask/history`.
- [ ] Register `POST /knowledge/entities/:id/ask/stream`.

### Task 4: Release Docs

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Add: `docs/releases/v0.6.20.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`

- [ ] Document new history and stream APIs.
- [ ] Leave Windsurf a concrete handoff for streaming Ask AI and history hydration.
- [ ] Bump both package versions to `0.6.20`.

### Task 5: Verification And Release

- [ ] Run focused Phase 63E tests.
- [ ] Run Phase 63 regression tests.
- [ ] Run `go test ./...`.
- [ ] Run `GOCACHE=$(pwd)/.cache/go-build go build ./...`.
- [ ] Run web lint and TypeScript checks.
- [ ] Commit, tag `v0.6.20`, push `main`, and push the tag.
