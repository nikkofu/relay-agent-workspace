# Phase 63B AI Compose Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a grounded `POST /api/v1/ai/compose` reply-suggestion endpoint for channel and thread composers.

**Architecture:** Reuse the existing AI gateway and knowledge/citation helpers. Build a small compose-grounding helper that gathers recent conversation context and matched entities, then have the handler generate a bounded set of send-ready suggestions with citations and context entity metadata.

**Tech Stack:** Go, Gin, GORM, SQLite, existing LLM gateway, existing message and knowledge services, pnpm lint/tsc for repo-wide verification.

---

### Task 1: Lock The Compose Contract With Tests

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Step 1: Add a failing handler test for `POST /api/v1/ai/compose`.
- [ ] Step 2: Cover:
  - channel compose
  - thread compose
  - draft-driven entity grounding
  - missing AI gateway
- [ ] Step 3: Run the focused test and verify it fails before implementation.
- [ ] Step 4: Re-run after implementation and keep assertions tight on `suggestions[]`, `citations[]`, and `context_entities[]`.

### Task 2: Implement Compose Grounding And Handler

**Files:**
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/knowledge/types.go`
- Modify: `apps/api/internal/knowledge/service.go`

- [ ] Step 1: Add lightweight DTOs for compose request/response.
- [ ] Step 2: Add a helper that gathers:
  - recent thread or channel messages
  - recent channel knowledge refs
  - draft-matched entities
- [ ] Step 3: Add `POST /api/v1/ai/compose`.
- [ ] Step 4: Keep this phase synchronous and stateless.
- [ ] Step 5: Ensure bounded output:
  - `intent` limited to `reply` only
  - default `limit=3`
  - `limit<1` normalizes to `3`
  - max `limit=5`

### Task 3: Docs, Verification, And Release

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Modify: `apps/web/components/agent-collab/agent-collab-data.ts`
- Add: `docs/releases/v0.6.14.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`

- [ ] Step 1: Document the Phase 63B API and Windsurf handoff.
- [ ] Step 2: Run:
  - `go test ./...`
  - `GOCACHE=$(pwd)/.cache/go-build go build ./...`
  - `pnpm --filter relay-agent-workspace lint`
  - `pnpm --filter relay-agent-workspace exec tsc --noEmit`
- [ ] Step 3: Commit, tag `v0.6.14`, and push after verification.
