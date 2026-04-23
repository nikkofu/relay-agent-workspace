# Phase 63G Compose Activity Persistence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist AI compose suggestion activity and expose it through a list API for co-drafting observer surfaces.

**Architecture:** Add one GORM model and reuse the existing compose finalize path. The list endpoint is read-only and follows existing newest-first filtered list patterns.

**Tech Stack:** Go, Gin, GORM, SQLite, project realtime hub.

---

### Task 1: Contract Test

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [x] Add `TestPhase63GComposeActivityPersistsAndLists`.
- [x] Run `go test ./internal/handlers -run TestPhase63G -count=1`.
- [x] Confirm RED from missing model/handler.

### Task 2: Persistence

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [x] Add `AIComposeActivity`.
- [x] Add production and test AutoMigrate.

### Task 3: Compose Hook And API

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/main.go`

- [x] Persist activity for sync and finalized stream compose.
- [x] Enrich websocket event with `activity`.
- [x] Add `GET /api/v1/ai/compose/activity`.

### Task 4: Documentation And Release

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Create: `docs/releases/v0.6.24.md`

- [x] Document endpoint, fields, websocket payload, and Windsurf handoff.
- [x] Run full verification.
- [x] Commit, tag, and push `v0.6.24`.
