# Phase 63A Knowledge Ask And Share Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add entity-scoped grounded Q&A, weekly brief snapshot sharing, and brief invalidation realtime events for the AI-native knowledge surfaces.

**Architecture:** Reuse the existing knowledge prompt-building and `AISummary` persistence patterns. Extend the knowledge service with small helper types for Q&A/share payloads, add handler-level orchestration plus websocket invalidation broadcasts, and cover the new contracts through collaboration handler integration tests.

**Tech Stack:** Go, Gin, GORM, SQLite, existing LLM gateway, existing realtime hub, pnpm lint/tsc for repo-wide verification.

---

### Task 1: Lock Phase 63A Contracts With Tests

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Step 1: Add a failing integration test covering:
  - `POST /api/v1/knowledge/entities/:id/ask`
  - `POST /api/v1/knowledge/weekly-brief/:id/share`
  - websocket `knowledge.entity.brief.changed`
- [ ] Step 2: Run the focused handler test and confirm it fails for missing routes/handlers or missing behavior.
- [ ] Step 3: Keep the assertions minimal but exact:
  - ask returns `answer`, `citations[]`, `entity`, `question`
  - weekly brief share returns share payload with stable URL fields
  - significant new entity ref emits `knowledge.entity.brief.changed`
- [ ] Step 4: Re-run the focused handler test after implementation.

### Task 2: Implement Phase 63A Backend

**Files:**
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/internal/knowledge/service.go`
- Modify: `apps/api/internal/knowledge/types.go`

- [ ] Step 1: Add lightweight response types/helpers for entity Q&A and weekly brief share payloads.
- [ ] Step 2: Implement a prompt-builder/helper for grounded entity Q&A reusing entity refs, timeline events, and entity metadata.
- [ ] Step 3: Add `POST /api/v1/knowledge/entities/:id/ask`.
- [ ] Step 4: Add `POST /api/v1/knowledge/weekly-brief/:id/share`.
- [ ] Step 5: Emit websocket `knowledge.entity.brief.changed` from entity-ref/event mutation paths that materially invalidate a cached entity brief.
- [ ] Step 6: Keep behavior cache-safe:
  - share works from existing `AISummary` snapshots only
  - ask requires configured LLM gateway

### Task 3: Docs, Verification, And Release

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Add: `docs/releases/v0.6.12.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`
- Modify: `apps/web/components/agent-collab/agent-collab-data.ts`

- [ ] Step 1: Document the new contracts and Windsurf handoff.
- [ ] Step 2: Run:
  - `go test ./...`
  - `GOCACHE=$(pwd)/.cache/go-build go build ./...`
  - `pnpm --filter relay-agent-workspace lint`
  - `pnpm --filter relay-agent-workspace exec tsc --noEmit`
- [ ] Step 3: Commit, tag `v0.6.12`, and push once verification is green.
