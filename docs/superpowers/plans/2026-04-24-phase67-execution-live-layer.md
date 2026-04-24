# Phase 67 Execution Live Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Relay's execution surfaces live, traceable, and lighter to refresh by adding execution websocket events, source-message jump/highlight behavior, a dedicated unread-mention count endpoint, and trend signals for Home execution pulse.

**Architecture:** Build on Phase 66 contracts rather than redesigning execution. Gemini extends backend websocket and count/trend contracts; Windsurf consumes them incrementally in the execution panel, message list, primary nav, and Home. Home uses event-driven stale marking plus debounced refresh instead of full row-level websocket reconciliation.

**Tech Stack:** Go, Gin, GORM, SQLite, existing realtime hub and structured workspace handlers, Next.js, Zustand stores, existing websocket hook, existing message list and Home dashboard.

**Spec:** `docs/superpowers/specs/2026-04-24-phase67-execution-live-layer-design.md`

---

### Task 1: Gemini Backend Contract Tests For Phase 67

**Files:**
- Modify: `apps/api/internal/handlers/structured_workspace_test.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/internal/handlers/realtime_test_helpers.go` if needed

- [ ] Add `TestPhase67ListItemEventsEmitLifecyclePayloads`.
- [ ] Add `TestPhase67ToolRunEventsEmitLifecyclePayloads`.
- [ ] Add `TestPhase67UnreadMentionCountsEndpoint`.
- [ ] Add `TestPhase67HomeExecutionPulseIncludesTrendFields`.
- [ ] Run `go test ./internal/handlers -run TestPhase67 -count=1`.
- [ ] Confirm RED before implementation.

### Task 2: Gemini Backend - Realtime Execution Event Lifecycle

**Files:**
- Modify: `apps/api/internal/handlers/structured_workspace.go`
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/handlers/structured_workspace_test.go`

- [ ] Freeze and emit list lifecycle events:
  - `list.item.created`
  - `list.item.updated`
  - `list.item.deleted`
- [ ] Freeze and emit tool lifecycle events:
  - `tool.run.started`
  - `tool.run.updated`
  - `tool.run.completed`
- [ ] Include `channel_id`, `list_id` or `run`, and a stable event timestamp in each payload.
- [ ] Freeze list-event `item` payload shape for Web consumption, including:
  - `id`
  - `list_id`
  - `content`
  - `is_completed`
  - `assigned_to`
  - `due_at`
  - `source_message_id`
  - `source_channel_id`
  - `source_snippet`
  - `updated_at`
- [ ] Freeze tool-event `run` payload shape for Web consumption, including:
  - `id`
  - `tool_id`
  - `tool_name`
  - `status`
  - `summary`
  - `channel_id`
  - `writeback_target`
  - `writeback`
  - `started_at`
  - `finished_at`
  - `duration_ms`
- [ ] Preserve existing `tool.run.updated` compatibility if already consumed anywhere.
- [ ] Run `cd apps/api && go test ./internal/handlers -run 'TestPhase67(ListItemEventsEmitLifecyclePayloads|ToolRunEventsEmitLifecyclePayloads)' -count=1`.

### Task 3: Gemini Backend - Lightweight Mention Count Endpoint

**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Modify: `apps/api/main.go`

- [ ] Add `GET /api/v1/me/unread-counts`.
- [ ] Return the frozen payload shape:
  - `{"counts":{"unread_mention_count":<int>}}`
- [ ] Source the count from the same durable mention/inbox semantics used in Phase 65C.
- [ ] Keep this endpoint small and safe for frequent refresh.
- [ ] Run `cd apps/api && go test ./internal/handlers -run TestPhase67UnreadMentionCountsEndpoint -count=1`.

### Task 4: Gemini Backend - Home Execution Pulse Trend Fields

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Extend `channel_execution_pulse[]` with additive trend fields:
  - `open_item_delta_7d`
  - `overdue_delta_7d`
  - `recent_tool_failure_count`
- [ ] Keep existing row ids and summary fields backward-compatible.
- [ ] Ensure empty channels do not emit misleading non-zero trend values.
- [ ] Run `cd apps/api && go test ./internal/handlers -run TestPhase67HomeExecutionPulseIncludesTrendFields -count=1`.

### Task 5: Windsurf Web - Consume Realtime Execution Events

**Files:**
- Modify: `apps/web/hooks/use-websocket.ts`
- Modify: `apps/web/stores/list-store.ts`
- Modify: `apps/web/stores/tool-store.ts`
- Modify: `apps/web/components/channel/channel-lists-panel.tsx`
- Modify: `apps/web/components/channel/channel-tools-panel.tsx`

- [ ] Consume `list.item.created`, `list.item.updated`, and `list.item.deleted` in the websocket hook.
- [ ] Consume `tool.run.started`, `tool.run.updated`, and `tool.run.completed` in the websocket hook.
- [ ] Add store-level append/update/remove actions so execution panels update incrementally instead of full re-fetching.
- [ ] Prevent duplicate rows and cross-channel contamination.
- [ ] Verify execution panels update live while remaining stable under repeated events.

### Task 6: Windsurf Web - Home Execution Stale Marking And Debounced Refresh

**Files:**
- Modify: `apps/web/stores/workspace-store.ts`
- Modify: `apps/web/hooks/use-websocket.ts`
- Modify: `apps/web/components/layout/home-dashboard.tsx`

- [ ] Add a scoped stale marker for Home execution data.
- [ ] On receipt of any Phase 67 execution event, mark Home execution stale.
- [ ] When Home is mounted, debounce a refresh of execution aggregates rather than full row-level reconciliation.
- [ ] Keep existing Home behavior intact when the execution blocks are absent or stale.

### Task 7: Windsurf Web - Source Message Jump And Flash Highlight

**Files:**
- Modify: `apps/web/components/message/message-list.tsx`
- Modify: `apps/web/components/message/message-item.tsx`
- Modify: `apps/web/app/workspace/page.tsx` if routing coordination is needed
- Modify: `apps/web/components/channel/channel-lists-panel.tsx`

- [ ] Preserve the existing `/workspace?c=<channel>#msg-<message>` link convention.
- [ ] Add stable DOM ids on message rows: `msg-<message_id>`.
- [ ] Watch URL hash changes and scroll the target message into view when present.
- [ ] Apply a temporary highlight treatment for 2-4 seconds, then clear it.
- [ ] Fail soft when the target message is absent: route to the right channel without breaking the page.

### Task 8: Windsurf Web - Lightweight Mention Badge Refresh

**Files:**
- Modify: `apps/web/components/layout/primary-nav.tsx`
- Modify: `apps/web/stores/activity-store.ts`
- Modify: `apps/web/stores/workspace-store.ts` if shared count storage is preferred

- [ ] Prefer `GET /api/v1/me/unread-counts` for the Activity mention badge over eager `GET /home` reads.
- [ ] Keep websocket mention/read updates as the first path when available.
- [ ] Use the lightweight endpoint as refresh/fallback synchronization, not as a replacement for live updates.
- [ ] Verify badge counts stay consistent after reads, reloads, and mention websocket events.

### Task 9: Windsurf Web - Home Execution Pulse Trend Rendering

**Files:**
- Modify: `apps/web/components/layout/home-execution-blocks.tsx`
- Modify: `apps/web/types/index.ts`

- [ ] Render `open_item_delta_7d`, `overdue_delta_7d`, and `recent_tool_failure_count` additively in `Channel Execution Pulse`.
- [ ] Use compact directionality UI such as delta badges or short sparkline-like treatment only if backed directly by payload fields.
- [ ] Avoid inventing derived trend semantics client-side beyond simple display formatting.
- [ ] Keep empty/zero trends visually quiet.

### Task 10: Codex Coordination, Documentation, And Verification

**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `CHANGELOG.md`
- Create: `docs/releases/v0.6.42.md`

- [ ] Record frozen Phase 67 contracts in `docs/AGENT-COLLAB.md`.
- [ ] Publish the execution order and parallel-safe boundaries:
  - Gemini Tasks 1-4
  - Windsurf Tasks 5-9
- [ ] After Gemini lands backend work, confirm Windsurf's consumer tasks remain aligned with the frozen event and endpoint shapes.
- [ ] Run `git diff --check`.
- [ ] After Gemini reports completion, verify backend with:
  - `cd apps/api && go test ./internal/handlers -run TestPhase67 -count=1`
- [ ] After Windsurf reports completion, verify the agreed Web checks they provide or repo-standard checks if supplied.

---

## Execution Notes

- Treat this as one phase with two implementation tracks and one control track:
  - `Gemini`: Tasks 1-4
  - `Windsurf`: Tasks 5-9
  - `Codex`: Task 10
- Preferred sequencing:
  - Gemini ships Tasks 1-4 first or in parallel where file scope permits
  - Windsurf can begin Task 7 immediately because hash/highlight is mostly frontend-local
  - Windsurf should wait for Gemini's frozen event names and `/me/unread-counts` endpoint before finishing Tasks 5, 6, and 8
- Do not broaden this phase into new workflow builders, new AI orchestration chains, or analytics-heavy dashboards.
- Home realtime in this phase is invalidate-and-refresh, not per-row websocket reconstruction.
