# Phase 63H AI Automation Suite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver the next AI-native backend automation wave for Relay by implementing three connected capabilities in one coordinated phase:

1. compose activity digest aggregation
2. entity brief auto-regeneration
3. schedule booking from structured compose slots

**Architecture:** Keep the handler layer thin, reuse the existing LLM/knowledge/realtime stack, and add only the minimum durable backend models needed for observability and restart safety. Phase 63H should reuse existing `AISummary`, compose persistence, and websocket broadcast patterns instead of introducing a separate orchestration framework.

**Tech Stack:** Go, Gin, GORM, SQLite, existing realtime hub, existing LLM gateway.

**Spec:** `docs/superpowers/specs/2026-04-23-phase63h-ai-automation-suite-design.md`

---

### Task 1: Contract Tests For The Whole Phase

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `TestPhase63HComposeActivityDigest`.
- [ ] Add `TestPhase63HEntityBriefAutomationLifecycle`.
- [ ] Add `TestPhase63HScheduleBookingLifecycle`.
- [ ] Extend test DB migrations for new models/fields before implementation.
- [ ] Run `go test ./internal/handlers -run TestPhase63H -count=1`.
- [ ] Confirm RED from missing models/routes/handlers before implementation.

### Task 2: Compose Activity Digest Model Updates And Query API

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add nullable `user_id` to `AIComposeActivity`.
- [ ] Add any required indexes for scope/time/user filtering.
- [ ] Update compose persistence to record `user_id` when available.
- [ ] Add `GET /api/v1/ai/compose/activity/digest`.
- [ ] Support `workspace_id`, `channel_id`, `dm_id`, `window`, `start_at`, `end_at`, `intent`, `group_by`, and `limit`.
- [ ] Enforce scope precedence and `window=custom` validation exactly as defined in the spec.
- [ ] Treat historical rows with null `user_id` as `"unknown"` only in `group_by=user` breakdowns.
- [ ] Exclude null `user_id` rows from `summary.unique_users` while still counting them in total activity and non-user groupings.

### Task 3: Entity Brief Automation Persistence And Service

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/internal/handlers/knowledge_scheduler.go`
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Create: `apps/api/internal/automation/entity_brief_jobs.go`

- [ ] Add `AIAutomationJob`.
- [ ] Add indexes needed for entity brief job lookups and stale sweeps.
- [ ] Implement enqueue-or-return-existing logic inside a transaction.
- [ ] Use deterministic dedupe keys with the 2-minute stale bucket from the spec.
- [ ] Reuse the current entity brief generation path instead of duplicating prompt/persistence logic.
- [ ] Emit `knowledge.entity.brief.regen.queued`.
- [ ] Emit `knowledge.entity.brief.regen.started`.
- [ ] Continue emitting `knowledge.entity.brief.generated` on success.
- [ ] Emit `knowledge.entity.brief.regen.failed` with error context safe for clients.
- [ ] Add startup wiring in `main.go` so the periodic stale sweep actually runs with the API process.

### Task 4: Entity Brief Automation APIs And Worker Hooks

**Files:**
- Modify: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/internal/automation/entity_brief_jobs.go`

- [ ] Add `GET /api/v1/knowledge/entities/:id/brief/automation`.
- [ ] Add `POST /api/v1/knowledge/entities/:id/brief/automation/run`.
- [ ] Add `POST /api/v1/knowledge/entities/:id/brief/automation/retry`.
- [ ] Ensure `/run` returns an existing `pending/running` job instead of duplicating work.
- [ ] Ensure `/retry` creates a fresh row only when the latest terminal job is `failed` or `succeeded`.
- [ ] Hook `knowledge.entity.brief.changed` so stale entities enqueue background regeneration.
- [ ] Add the periodic stale sweep so long-stale entities without active jobs are recovered.

### Task 5: Schedule Booking Persistence And Core Handler

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Create: `apps/api/internal/automation/schedule_booking.go`

- [ ] Add `AIScheduleBooking`.
- [ ] Persist `ics_content` inline for first-version durability.
- [ ] Persist `requested_by` from the current user context for auditability.
- [ ] Validate that exactly one of `channel_id` or `dm_id` is present.
- [ ] Require `compose_id` and a complete `slot`.
- [ ] Add `POST /api/v1/ai/schedule/book`.
- [ ] Create internal booking rows with `provider=internal` as the default path.
- [ ] Accept `google`, `outlook`, and `open` provider values without requiring a configured adapter in first version.
- [ ] Return `delivery.mode="ics_fallback"` when a provider adapter is unavailable.

### Task 6: Schedule Booking Read/Cancel APIs And Realtime

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/main.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/internal/automation/schedule_booking.go`

- [ ] Add `GET /api/v1/ai/schedule/bookings`.
- [ ] Add `GET /api/v1/ai/schedule/bookings/:id`.
- [ ] Add `POST /api/v1/ai/schedule/bookings/:id/cancel`.
- [ ] Emit `schedule.event.booked`.
- [ ] Emit `schedule.event.updated` when booking state changes after creation.
- [ ] Emit `schedule.event.cancelled`.
- [ ] Make cancel idempotent and preserve the booking row as durable history.

### Task 7: Verification

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/internal/handlers/ai_test.go`

- [ ] Run `go test ./internal/handlers -run TestPhase63H -count=1`.
- [ ] Run `go test ./internal/handlers -count=1`.
- [ ] Run `go test ./...`.
- [ ] Run `GOCACHE=$(pwd)/.cache/go-build go build ./...`.
- [ ] Run `pnpm --filter relay-agent-workspace lint`.
- [ ] Run `pnpm --filter relay-agent-workspace exec tsc --noEmit`.
- [ ] Run `git diff --check`.

### Task 8: Documentation, Collaboration, And Release

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase8-api-expansion.md`
- Create: `docs/releases/v0.6.26.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`

- [ ] Document the new digest API, entity automation APIs/events, and schedule booking APIs/events.
- [ ] Update collaboration notes for Windsurf with concrete UI payload and polling/WS expectations.
- [ ] Sync project introduction where needed if the new AI automation surfaces change the public product story.
- [ ] Bump versions consistently for the new release.
- [ ] Commit, tag, and push `v0.6.26`.

---

## Execution Notes

- Start with Task 1 and Task 2 so the digest contract and `user_id` migration land before more automation code depends on them.
- Task 3 and Task 4 should share one implementation branch of work because the APIs depend on the persisted job model and worker hooks.
- Task 5 and Task 6 should keep provider integration intentionally shallow in first version. Internal booking + durable ICS fallback is the required bar.
- Do not introduce a generic workflow engine in this phase.
- Prefer isolated helper/service files under `apps/api/internal/automation/` for new stateful logic, while keeping handlers focused on validation and response shaping.
