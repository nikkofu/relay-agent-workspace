# Phase 64C Unified Activity Feed Completion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the unified activity feed for Relay by adding the six remaining Slack-like event types needed by the shared activity rail and hardening the persistence and link mapping behind those events.

**Architecture:** Keep `GET /api/v1/activity/feed` as the single aggregation point, reuse existing durable models first, and add only the minimum mapping or persistence glue required for stable feed rows. Implement the phase in three internal slices so each source family can be tested independently while preserving one public release.

**Tech Stack:** Go, Gin, GORM, SQLite, existing activity/feed handlers, existing message/DM/workflow/artifact persistence, pnpm TypeScript frontend consuming a stable REST contract.

**Spec:** `docs/superpowers/specs/2026-04-23-phase64c-unified-activity-feed-completion-design.md`

---

### Task 1: Phase 64C Contract Tests

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `TestPhase64CArtifactAndToolRunFeedItems`.
- [ ] Add `TestPhase64CReplyAndDMFeedItems`.
- [ ] Add `TestPhase64CMentionAndReactionFeedItems`.
- [ ] Seed only persisted data sources used by the handler; do not fake feed rows directly.
- [ ] Assert each new row contains non-empty `event_type`, `title`, and `occurred_at`.
- [ ] Assert filtering by `event_type`, `channel_id`, and `dm_id` works for the new rows.
- [ ] Run `go test ./internal/handlers -run TestPhase64C -count=1`.
- [ ] Confirm RED before implementation.

### Task 2: Artifact And Tool Run Feed Completion

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Inspect: `apps/api/internal/domain/models.go`

- [ ] Identify the existing artifact persistence fields that can produce stable `artifact_updated` rows.
- [ ] Identify the existing workflow/tool run persistence fields that can produce stable `tool_run` rows.
- [ ] Extend `GetActivityFeed` to append `artifact_updated` rows.
- [ ] Extend `GetActivityFeed` to append `tool_run` rows.
- [ ] Map `link` and `meta` exactly as the spec describes.
- [ ] Ensure missing actor/channel metadata degrades gracefully instead of failing the request.
- [ ] Run `go test ./internal/handlers -run TestPhase64CArtifactAndToolRunFeedItems -count=1`.

### Task 3: Reply And DM Message Feed Completion

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Inspect: `apps/api/internal/domain/models.go`

- [ ] Extend channel message aggregation so messages with non-empty `thread_id` map to `reply`.
- [ ] Ensure thread replies do not also emit as generic `message` items.
- [ ] Add persisted DM message aggregation as `dm_message`.
- [ ] Map stable thread and DM links using existing workspace routes.
- [ ] Resolve actor and scope metadata from persisted message/conversation data, not frontend assumptions.
- [ ] Run `go test ./internal/handlers -run TestPhase64CReplyAndDMFeedItems -count=1`.

### Task 4: Mention And Reaction Feed Completion

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Inspect: `apps/api/internal/handlers/activity.go`
- Inspect: `apps/api/internal/domain/models.go`

- [ ] Trace the lightest durable mention source already present in the codebase.
- [ ] Map deterministic mention rows into the unified feed without inventing new source records.
- [ ] Extend the feed with persisted reaction rows.
- [ ] Skip unresolved mention or reaction records instead of returning broken rows.
- [ ] Ensure `meta.mention_kind`, `meta.emoji`, and `message_id` fields are present where applicable.
- [ ] Run `go test ./internal/handlers -run TestPhase64CMentionAndReactionFeedItems -count=1`.

### Task 5: Feed Filtering, Ordering, And Regression Pass

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Verify all twelve feed event types can coexist in one merged result without duplicate primary rows.
- [ ] Verify `event_type` filtering matches both old and new event types.
- [ ] Verify descending `occurred_at` ordering still holds after adding the new sources.
- [ ] Verify cursor behavior remains stable with mixed old/new rows.
- [ ] Run `go test ./internal/handlers -run TestPhase64BUnifiedActivityFeedContract -count=1`.
- [ ] Run `go test ./internal/handlers -run TestPhase64C -count=1`.

### Task 6: Documentation And Collaboration Sync

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase64-slack-core-convergence.md`
- Create: `docs/releases/v0.6.xx.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`

- [ ] Document the expanded activity feed source coverage and filtering semantics.
- [ ] Leave Windsurf a concrete handoff for the newly-added event types and `meta` payloads.
- [ ] Record that the next rounds should continue items `2` and `3` from the earlier decision.
- [ ] Bump the release version consistently across backend and web package manifests.

### Task 7: Verification And Release

**Files:**
- No additional files required unless verification exposes gaps.

- [ ] Run `go test ./internal/handlers -count=1`.
- [ ] Run `go test ./...`.
- [ ] Run `GOCACHE=$(pwd)/.cache/go-build go build ./...`.
- [ ] Run `pnpm --filter relay-agent-workspace lint`.
- [ ] Run `pnpm --filter relay-agent-workspace exec tsc --noEmit`.
- [ ] Run `git diff --check`.
- [ ] Commit the phase with a release-oriented message.
- [ ] Tag and push the new version.

---

## Execution Notes

- Implement in three internal cuts even if the release ships as one version:
  - `64C-A`: `artifact_updated` + `tool_run`
  - `64C-B`: `reply` + `dm_message`
  - `64C-C`: `mention` + `reaction`
- Do not redesign the feed contract or introduce a generic event framework in this phase.
- Prefer adding small helper functions inside `workspace.go` only when they reduce repeated source-to-feed mapping logic; do not start a broad refactor unless the file becomes unworkable during implementation.
- If mention persistence is weaker than expected, use the most durable existing source already exposed by inbox/mentions/message metadata and document the limitation in the release notes instead of overbuilding a new table in this phase.
- After this round, continue the user’s remembered follow-ups:
  - `2`: keep strengthening the unified feed itself
  - `3`: broaden into larger Slack-like capability work outside the feed slice
