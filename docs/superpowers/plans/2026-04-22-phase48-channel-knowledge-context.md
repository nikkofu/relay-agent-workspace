# Phase 48 Channel Knowledge Context Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose channel-level knowledge context and ensure citation/search results hydrate entity links created by message/file auto-linking.

**Architecture:** Keep the knowledge layer as the single source of truth for entity refs. Extend citation hydration so `KnowledgeEntityRef` links are honored without requiring separate `KnowledgeEvidenceLink` rows, then add a thin channel context handler that aggregates recent entity refs for messages and files in a channel.

**Tech Stack:** Go, Gin, GORM, SQLite, existing `internal/knowledge` package, existing handler test harness.

---

### Task 1: Hydrate citations from KnowledgeEntityRef

**Files:**
- Modify: `apps/api/internal/knowledge/service_test.go`
- Modify: `apps/api/internal/knowledge/service.go`

- [ ] **Step 1: Write the failing test**
  - Add a test proving `Lookup` returns `entity_id` and `entity_title` for a message linked through `KnowledgeEntityRef{ref_kind:"message", ref_id:"msg-1"}`.

- [ ] **Step 2: Run the focused test**
  - Run: `cd apps/api && go test ./internal/knowledge -run TestLookupHydratesEntityFromKnowledgeEntityRef`
  - Expected: fail because lookup only checks `KnowledgeEvidenceLink`.

- [ ] **Step 3: Implement minimal hydration**
  - Extend `attachEntity` fallback logic to check `KnowledgeEntityRef` by `ref_kind/ref_id`.
  - For file chunks, also allow `ref_kind=file` using `source_ref`.

- [ ] **Step 4: Verify**
  - Run: `cd apps/api && go test ./internal/knowledge`

### Task 2: Add channel knowledge context API

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/internal/knowledge/types.go`
- Modify: `apps/api/internal/knowledge/service.go`
- Modify: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/main.go`

- [ ] **Step 1: Write the failing API test**
  - Add a test for `GET /api/v1/channels/:id/knowledge` returning recent entity refs for messages and files in that channel, including entity title and source context.

- [ ] **Step 2: Run the focused test**
  - Run: `cd apps/api && go test ./internal/handlers -run TestGetChannelKnowledgeContextReturnsRecentRefs`
  - Expected: fail because route/handler do not exist.

- [ ] **Step 3: Add service response types and query**
  - Add `ChannelKnowledgeContext`, `ChannelKnowledgeRef`, and `GetChannelKnowledgeContext`.
  - Query message refs by joining `messages`.
  - Query file refs by joining `file_assets`.
  - Sort by ref creation time descending and cap with `limit`.

- [ ] **Step 4: Add handler and route**
  - Register `GET /api/v1/channels/:id/knowledge`.
  - Return `{ "context": ... }`.

- [ ] **Step 5: Verify**
  - Run focused handler test, then `go test ./internal/knowledge`, `go test ./internal/handlers`, `go test ./...`, `go build ./...`, and frontend lint.

### Task 3: Release docs and collaboration handoff

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `README.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `apps/web/components/agent-collab/agent-collab-data.ts`
- Modify: `package.json`
- Modify: `apps/web/package.json`
- Add: `docs/releases/v0.5.88.md`

- [ ] **Step 1: Document API contract**
  - Include `GET /api/v1/channels/:id/knowledge` and citation hydration behavior.

- [ ] **Step 2: Give Windsurf the next UI task**
  - Ask Windsurf to render channel knowledge banners/sidebars and rely on hydrated citation entity links.

- [ ] **Step 3: Release**
  - Commit, tag `v0.5.88`, push `main` and tag.
