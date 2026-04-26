# Phase 71 AI Execution History Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add append-only AI execution history events so list/workflow/channel_message executions can be traced from analysis snapshot and next step through draft, confirmation, created object, or failure, with projections in Canvas AI Dock, Activity, and Home.

**Architecture:** Introduce an independent `ExecutionHistoryEvent` fact model. Gemini writes append-only events from the existing list, workflow, and channel_message execution chains and exposes deterministic query/projection APIs. Windsurf consumes the projections in Canvas AI Dock, Activity, and Home without relying on local UI state as the source of truth.

**Tech Stack:** Go/Gin API, GORM, SQLite, existing Activity/Home handlers, Next.js 16, React 19, TypeScript, Zustand stores, Canvas AI Dock components.

---

## File Structure

### Existing files likely involved

- `apps/api/internal/domain/*`
  - add `ExecutionHistoryEvent` model and helpers
- `apps/api/internal/handlers/ai.go`
  - existing AI canvas draft/create/publish endpoints
- `apps/api/internal/handlers/activity.go`
  - project execution events into Activity
- `apps/api/internal/handlers/home.go`
  - project execution summary into Home
- `apps/api/internal/handlers/phase70a*_test.go`
  - inspect existing list draft/create tests
- `apps/api/internal/handlers/phase70c*_test.go`
  - inspect workflow/message draft/create/publish tests
- `apps/web/components/layout/canvas-ai-dock.tsx`
  - render analysis-scoped history
- `apps/web/components/canvas/file-group-analysis-result.tsx`
  - per-step state display
- `apps/web/stores/activity-store.ts`
  - consume Activity projection if store exists
- `apps/web/components/layout/home-dashboard.tsx`
  - consume Home projection

### New files recommended

- `apps/api/internal/handlers/phase71_execution_history_test.go`
  - model/write/query tests
- `apps/api/internal/handlers/phase71_projection_test.go`
  - Activity/Home projection tests
- `apps/web/lib/ai-execution-history.ts`
  - shared Web types and state-resolution helpers
- `apps/web/components/canvas/ai-execution-history-status.tsx`
  - per-step status display

## Contract Freeze

- **Codex authority:** Phase 71 is frozen around append-only execution history facts.
- event model:
  - independent `ExecutionHistoryEvent`
  - append-only writes
  - no mutation of prior events to represent current status
  - current state computed by projection
- event types:
  - `draft_generated`
  - `confirmed`
  - `created`
  - `published`
  - `failed`
- execution families:
  - `list`
  - `workflow`
  - `channel_message`
- Canvas grouping:
  - `analysis_snapshot_id + step locator + execution_target_type`
- Activity:
  - event family `ai_execution`
  - first release includes `created`, `published`, `failed`
  - include `draft_generated` or `confirmed` only when no later event exists for that group
  - default page size: 20
- Home:
  - last 7 days
  - latest 5 successful executions
  - latest 5 failed executions

## Task 0: Codex Contract Preflight Gate

**Owner:** Codex  
**Files:**
- Reference: `docs/superpowers/specs/2026-04-26-phase71-ai-execution-history-design.md`
- Reference: `docs/superpowers/plans/2026-04-26-phase71-ai-execution-history.md`

- [ ] **Step 1: Freeze event taxonomy**

Confirm no changes are needed to:
- event types
- required field matrix
- append-only rule
- Canvas grouping keys
- Activity filtering/dedupe rules
- Home window and limits

- [ ] **Step 2: Publish implementation note in collaboration docs**

Add a short Phase 71 kickoff note to `docs/AGENT-COLLAB.md` before Gemini/Windsurf implementation begins.

- [ ] **Step 3: Commit**

```bash
git add docs/AGENT-COLLAB.md
git commit -m "docs: kick off phase 71 ai execution history"
```

## Task 1: Backend Execution History Model And Append-Only Writes

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/domain/*`
- Create/Modify: `apps/api/internal/handlers/phase71_execution_history_test.go`

- [ ] **Step 1: Write the failing model tests**

Add tests covering:
- required field validation for each event type
- append-only behavior: new lifecycle transition creates a new row
- current state is not stored by mutating prior events
- one-of step locator rule: prefer `next_step_id`, otherwise `step_index`

- [ ] **Step 2: Run the focused model tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase71ExecutionHistoryModel -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement model and helper functions**

Requirements:
- add `ExecutionHistoryEvent`
- add append helper with per-event validation
- include `actor_user_id`, `analysis_snapshot_id`, target type, draft/object IDs where required

- [ ] **Step 4: Re-run the focused model tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/domain apps/api/internal/handlers/phase71_execution_history_test.go
git commit -m "feat(api): add phase 71 execution history model"
```

## Task 2: Backend Event Writes From Existing Execution Chains

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: relevant list/workflow/message execution helpers
- Test: extend `apps/api/internal/handlers/phase71_execution_history_test.go`

- [ ] **Step 1: Write the failing event-write tests**

Add tests covering:
- list draft/create writes `draft_generated`, `confirmed`, `created`
- workflow draft/create writes `draft_generated`, `confirmed`, `created`
- channel_message draft/publish writes `draft_generated`, `confirmed`, `published`
- failure during draft/confirm/create/publish writes `failed` with correct `failure_stage`
- failure events persist even when the API returns an error response
- later lifecycle transitions append new rows instead of mutating previous events on the actual handler/request path

- [ ] **Step 2: Run the focused event-write tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase71ExecutionHistoryWrites -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement append points**

Requirements:
- write events from list, workflow, and channel_message chains
- do not replace existing success/failure behavior
- persist failure events even when API returns an error
- never update a prior history row to represent a later lifecycle state

- [ ] **Step 4: Re-run the focused event-write tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/internal/domain
git commit -m "feat(api): record phase 71 ai execution events"
```

## Task 3: Backend Query API And Canvas Projection

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go` or add focused handler file if repo convention supports it
- Create/Modify: `apps/api/internal/handlers/phase71_execution_history_test.go`
- Modify: `apps/api/main.go`

- [ ] **Step 1: Write the failing query/projection tests**

Add tests covering:
- query by `analysis_snapshot_id`
- grouping by `analysis_snapshot_id + step locator + execution_target_type`
- deterministic ordering by `created_at`, then `id`
- current state is the final event in each group
- object links and failure fields are present when available

- [ ] **Step 2: Run the focused query tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase71ExecutionHistoryQuery -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement query API**

Requirements:
- expose analysis-scoped execution history for Canvas AI Dock
- return grouped current-state data plus enough event detail for failure context

- [ ] **Step 4: Re-run the focused query tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/main.go
git commit -m "feat(api): add phase 71 execution history query"
```

## Task 4: Backend Activity And Home Projections

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/activity.go`
- Modify: `apps/api/internal/handlers/home.go`
- Create/Modify: `apps/api/internal/handlers/phase71_projection_test.go`

- [ ] **Step 1: Write the failing projection tests**

Activity tests:
- includes `created`, `published`, `failed`
- includes `draft_generated` or `confirmed` only when no later event exists for group
- orders by `created_at` desc, then `id` desc
- default page size 20
- dedupes by `analysis_snapshot_id + step locator + execution_target_type + event_type`
- uses duplicate same-group fixtures to prove `draft_generated` and `confirmed` are suppressed after later `created` / `published` / `failed` events exist

Home tests:
- last 7 days window
- latest 5 successful executions
- latest 5 failed executions
- includes source and object links when available

- [ ] **Step 2: Run the focused projection tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase71ExecutionHistoryProjection -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement Activity/Home projections**

Requirements:
- add `ai_execution` Activity family or compatible typed payload
- add Home blocks for recent and failed AI-driven executions
- keep projections bounded to the spec limits

- [ ] **Step 4: Re-run the focused projection tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers
git commit -m "feat(api): project phase 71 ai execution history"
```

## Task 5: Web Types And Canvas Dock Execution Status

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/ai-execution-history.ts`
- Create/Modify: `apps/web/components/canvas/ai-execution-history-status.tsx`
- Modify: `apps/web/components/canvas/file-group-analysis-result.tsx`
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`

- [ ] **Step 1: Write the failing Canvas status tests**

Add tests covering:
- fetch/query analysis-scoped execution history
- group by step locator and target type
- show current state from persisted history
- show created object link
- show persisted failure state after refresh

- [ ] **Step 2: Run the focused Web tests to verify failure**

Run the narrowest Canvas/history tests.

Expected:
- FAIL

- [ ] **Step 3: Implement Canvas execution status rendering**

Requirements:
- use backend history as source of truth
- do not infer execution state from local buttons
- preserve analysis context while showing object/failure state

- [ ] **Step 4: Re-run the focused Web tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/ai-execution-history.ts apps/web/components/canvas apps/web/components/layout/canvas-ai-dock.tsx
git commit -m "feat(web): show phase 71 canvas execution history"
```

## Task 6: Web Activity And Home Consumption

**Owner:** Windsurf  
**Files:**
- Modify: Activity feed components/stores
- Modify: `apps/web/components/layout/home-dashboard.tsx`
- Test: Activity/Home rendering tests

- [ ] **Step 1: Write the failing Activity/Home tests**

Activity:
- render `ai_execution` events
- show target type and event state
- link to source/created object when available

Home:
- render recent AI executions
- render failed AI executions needing attention
- respect backend-provided projection shape without client-side re-deriving limits

- [ ] **Step 2: Run the focused Activity/Home tests to verify failure**

Run the narrowest Activity/Home tests.

Expected:
- FAIL

- [ ] **Step 3: Implement Activity/Home consumption**

Requirements:
- consume backend projections
- do not build a full audit viewer
- keep display compact

- [ ] **Step 4: Re-run the focused Activity/Home tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components apps/web/stores
git commit -m "feat(web): surface phase 71 ai execution activity"
```

## Task 7: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md`
- Reference: `docs/superpowers/specs/2026-04-26-phase71-ai-execution-history-design.md`

- [ ] **Step 1: Review Gemini backend diff against the frozen Phase 71 contract**

Verify:
- append-only event writes
- per-event field matrix
- failure events are persisted
- Activity/Home projections follow bounded rules

- [ ] **Step 2: Review Windsurf Web diff against the frozen Phase 71 contract**

Verify:
- Canvas reads backend history as source of truth
- Activity/Home consume projections
- no local-only execution status after refresh
- no full audit viewer added

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase71ExecutionHistory|TestPhase71ExecutionHistoryProjection' -count=1
```

Web:
- run focused checks covering:
  - Canvas execution history status
  - created object links
  - persisted failure state
  - Activity AI execution events
  - Home recent/failed execution blocks

Static:
```bash
cd apps/web && pnpm exec tsc --noEmit
```

- [ ] **Step 4: Update collaboration and release docs**

Update:
- `docs/AGENT-COLLAB.md` with completion notes and next-phase hooks
- `docs/releases/v0.6.xx.md` with concise user-facing release notes

- [ ] **Step 5: Commit**

```bash
git add docs/AGENT-COLLAB.md docs/releases
git commit -m "docs: close out phase 71 ai execution history"
```

## Parallelization Guidance

Safe parallelism:

- Codex Task 0 must complete before Gemini/Windsurf implementation begins.
- Gemini Task 1 must land before backend event writes.
- Gemini Task 2 and Task 3 are sequential because query API depends on event writes.
- Gemini Task 4 depends on Task 1 and can begin once event model and sample write helpers exist.
- Windsurf Task 5 can start once Gemini Task 3 response shape is stable.
- Windsurf Task 6 can start once Gemini Task 4 projection shape is stable.
- Codex Task 7 happens after Gemini and Windsurf complete their slices.

## Verification Checklist

- [ ] Execution history events are append-only
- [ ] Codex preflight confirms taxonomy and projection boundary before implementation begins
- [ ] Event required fields match the per-event matrix
- [ ] List/workflow/channel_message chains all write history
- [ ] Failure events persist and survive refresh
- [ ] Activity projection dedupe/suppression matches the frozen grouping rules
- [ ] Canvas shows per-step execution state from backend history
- [ ] Activity shows bounded `ai_execution` events
- [ ] Home shows recent and failed AI execution summaries
- [ ] No new execution target types are introduced

## Non-Goals Reminder

Do not add in this phase:

- regulatory audit guarantees
- automatic retry orchestration
- cross-workspace analytics
- graph visualization
- full audit viewer UI
