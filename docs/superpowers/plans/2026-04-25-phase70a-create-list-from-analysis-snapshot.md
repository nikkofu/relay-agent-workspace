# Phase 70A Create List From Analysis Snapshot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let Canvas AI Dock convert a frozen structured analysis snapshot into one reviewable List draft for the current channel, then create exactly one list with multiple items only after explicit user confirmation.

**Architecture:** Build on Phase 69 structured analysis snapshots. Gemini provides a two-step backend flow: draft generation from an immutable `analysis_snapshot_id`, then confirm-create from an immutable `draft_id`. Windsurf keeps the entire review/create flow inside Canvas AI Dock, never reparses prose, and never auto-creates. Relay advances from AI recommendation to executable work object through an explicit draft boundary.

**Tech Stack:** Next.js 16, React 19, TypeScript, Zustand stores, TipTap canvas editor, Go/Gin API, GORM, structured workspace/list APIs, frontend component tests, Go handler/domain tests.

---

## File Structure

### Existing files likely involved

- `apps/web/components/layout/canvas-ai-dock.tsx`
  - existing structured analysis surface and the likely entry point for `Create list from plan`
- `apps/web/components/canvas/file-group-analysis-result.tsx`
  - existing Phase 69 structured result renderer; likely home for the create-list affordance
- `apps/web/stores/list-store.ts`
  - existing list creation/fetching flows that should be reused for final open/list refresh behavior
- `apps/web/stores/artifact-store.ts`
  - canvas context and artifact/channel metadata source
- `apps/api/internal/handlers/ai.go`
  - likely home for draft-generation route if kept under AI namespace
- `apps/api/internal/handlers/structured_workspace.go`
  - inspect existing list creation handlers and shared DTOs
- `apps/api/internal/domain/*`
  - inspect for list domain models and creation helpers

### New files recommended

- `apps/web/lib/analysis-list-draft.ts`
  - shared Phase 70A request/response types for `analysis_snapshot_id -> draft` and `draft_id -> create`
- `apps/web/components/canvas/analysis-list-draft-preview.tsx`
  - Dock-local draft preview component with title, items, channel, confirm state, and failure state
- `apps/api/internal/handlers/phase70a_test.go`
  - request validation + two-step contract tests
- `apps/api/internal/handlers/phase70a_create_test.go`
  - confirm-create tests proving one list + many items and proper failure handling

## Contract Freeze

- **Codex authority:** Phase 70A is frozen around an immutable two-step chain:
  - `analysis_snapshot_id -> draft_id -> confirm-create`
- conceptual draft-generation input:
  ```json
  {
    "artifact_id": "artifact-123",
    "channel_id": "channel-123",
    "analysis_snapshot_id": "analysis-123"
  }
  ```
- conceptual draft-generation output:
  ```json
  {
    "draft": {
      "draft_id": "draft-123",
      "channel_id": "channel-123",
      "title": "Launch readiness follow-up",
      "items": [
        { "title": "Confirm missing launch dependencies" }
      ]
    }
  }
  ```
- conceptual confirm-create input:
  ```json
  {
    "draft_id": "draft-123"
  }
  ```
- first-release transformation rules:
  - exactly one draft item per `next_steps[]` entry
  - preserve original `next_steps[]` order
  - item schema is title-only in first release
  - no assignee, due date, priority, or freeform item description generation

## Task 1: Backend Draft Generation Contract

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Create/Modify: `apps/api/internal/handlers/phase70a_test.go`
- Inspect: route wiring in `apps/api/main.go`

- [ ] **Step 1: Write the failing backend draft-generation tests**

Add tests covering:
- valid `analysis_snapshot_id` + `artifact_id` + `channel_id` -> returns `draft_id`, title, and ordered items
- missing or invalid `analysis_snapshot_id` -> rejected
- malformed request -> rejected
- no prose-only fallback allowed

- [ ] **Step 2: Run the focused backend tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70A -count=1
```

Expected:
- FAIL until the route/schema exists

- [ ] **Step 3: Implement minimal draft-generation contract**

Requirements:
- accept immutable `analysis_snapshot_id`
- derive exactly one title plus one ordered item per `next_steps[]`
- return immutable `draft_id`
- reject invalid/unstructured sources

- [ ] **Step 4: Re-run the focused backend tests**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70A -count=1
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/main.go
git commit -m "feat(api): add phase 70a analysis-to-list draft contract"
```

## Task 2: Backend Confirm-Create Contract

**Owner:** Gemini  
**Files:**
- Modify: existing list creation handlers/helpers under `apps/api/internal/handlers/structured_workspace.go` and related domain files
- Create/Modify: `apps/api/internal/handlers/phase70a_create_test.go`

- [ ] **Step 1: Write the failing confirm-create tests**

Add tests covering:
- valid `draft_id` -> creates exactly one list with multiple items
- invalid or missing `draft_id` -> rejected
- stale/missing draft -> create fails clearly
- failure path does not return success-like payload

- [ ] **Step 2: Run the focused confirm-create tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70ACreate -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement minimal confirm-create flow**

Requirements:
- consume `draft_id`, not mutable snapshot text
- create exactly one list in the specified/default channel
- create one ordered list item per draft item
- do not recompute from mutable Dock state at confirm time

- [ ] **Step 4: Re-run the focused confirm-create tests**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70ACreate -count=1
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/internal/domain
git commit -m "feat(api): add phase 70a draft confirm-create flow"
```

## Task 3: Shared Web Types For Draft/Create Flow

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/analysis-list-draft.ts`
- Inspect/Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Test: helper/type-consumer test following repo convention

- [ ] **Step 1: Write the failing Web contract tests**

Add tests for:
- draft-generation response shape with `draft_id`, `channel_id`, `title`, `items[]`
- confirm-create input shape with `draft_id`
- no support for plain-text/prose-only conversion source

- [ ] **Step 2: Run the focused Web tests to verify failure**

Run the narrowest test command for the new helper/type tests.

Expected:
- FAIL

- [ ] **Step 3: Implement the shared Web contract module**

Requirements:
- define request/response types for draft generation
- define request/response types for confirm-create
- keep Web rendering and submission logic tied to these shared types

- [ ] **Step 4: Re-run the focused Web tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/analysis-list-draft.ts apps/web
git commit -m "feat(web): add phase 70a analysis-list draft contract"
```

## Task 4: Dock Action Visibility And Draft Preview

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify/Create: `apps/web/components/canvas/analysis-list-draft-preview.tsx`
- Possibly modify: `apps/web/components/canvas/file-group-analysis-result.tsx`
- Test: Dock visibility + draft preview rendering tests

- [ ] **Step 1: Write the failing Dock tests**

Minimum scenarios:
- `Create list from plan` appears only for valid structured analysis snapshots
- clicking the action requests a draft using `analysis_snapshot_id`
- draft preview renders title, items, and target channel
- invalid/missing snapshot does not expose the action

- [ ] **Step 2: Run the focused Dock tests to verify failure**

Run the narrowest Dock/component tests.

Expected:
- FAIL

- [ ] **Step 3: Implement Dock visibility and preview**

Requirements:
- tie the action to the immutable snapshot identity
- render draft preview in-place in Dock
- keep draft preview distinct from the underlying analysis view
- default target channel from current canvas context

- [ ] **Step 4: Re-run the focused Dock tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas apps/web
git commit -m "feat(web): add phase 70a dock draft preview"
```

## Task 5: Explicit Confirm-Create UX And Success/Failure States

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify: `apps/web/components/canvas/analysis-list-draft-preview.tsx`
- Possibly modify: `apps/web/stores/list-store.ts`
- Test: confirm-create flow tests

- [ ] **Step 1: Write the failing create-flow tests**

Minimum cases:
- `Confirm create` submits only `draft_id`
- success state shows `Open list` and `Back to analysis`
- failure keeps draft preview visible and clearly uncreated
- no auto-create happens on draft generation
- mutating current Dock analysis state after draft open does not change the create payload; confirm still uses the previously returned `draft_id`

- [ ] **Step 2: Run the focused create-flow tests to verify failure**

Run the narrowest create-flow test command.

Expected:
- FAIL

- [ ] **Step 3: Implement explicit confirm-create flow**

Requirements:
- confirm with immutable `draft_id`
- no hidden recompute from current Dock text
- success state only after backend create succeeds
- preserve analysis snapshot and draft state on failure

- [ ] **Step 4: Re-run the focused create-flow tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas apps/web/stores/list-store.ts
git commit -m "feat(web): add phase 70a confirm-create flow"
```

## Task 6: Navigation And Channel-Context Integration

**Owner:** Windsurf  
**Files:**
- Modify: Dock success-state components
- Modify: any list-opening/navigation glue needed in current workspace shell
- Test: navigation/action tests

- [ ] **Step 1: Write the failing navigation tests**

Minimum cases:
- `Open list` resolves to the created list in the current channel context
- `Back to analysis` returns to the prior analysis view without losing it unnecessarily
- default channel in preview matches the canvas channel context
- returning to analysis does not silently swap the existing draft preview to a newer snapshot without explicit user action

- [ ] **Step 2: Run the focused navigation tests to verify failure**

Expected:
- FAIL

- [ ] **Step 3: Implement navigation/state handoff**

Requirements:
- respect current channel context
- open the created list without breaking the Dock flow
- keep the analysis view reachable after success
- preserve the already-open draft preview's binding to its original snapshot/draft chain until the user explicitly abandons or regenerates it

- [ ] **Step 4: Re-run the focused navigation tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas apps/web
git commit -m "feat(web): add phase 70a list navigation handoff"
```

## Task 7: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md`
- Reference: `docs/superpowers/specs/2026-04-25-phase70a-create-list-from-analysis-snapshot-design.md`

- [ ] **Step 1: Review Gemini's backend diff against the frozen Phase 70A contract**

Verify:
- immutable `analysis_snapshot_id` is required for draft generation
- immutable `draft_id` is required for confirm-create
- one-list-many-items shape is preserved
- no hidden prose parsing or AI rerun was introduced

- [ ] **Step 2: Review Windsurf's Web diff against the frozen Phase 70A contract**

Verify:
- action only appears for valid structured snapshots
- draft preview is Dock-local
- confirm-create submits only `draft_id`
- failure states preserve source analysis/draft context

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase70A|TestPhase70ACreate' -count=1
```

Web:
- run focused checks covering:
  - draft/create contract types
  - Dock visibility
  - draft preview rendering
  - confirm-create success/failure
  - navigation handoff

End-to-end immutable-chain smoke:
- run one explicit flow proving:
  - a valid `analysis_snapshot_id` produces a `draft_id`
  - Dock preview binds to that returned `draft_id`
  - mutating/replacing the visible analysis state after draft open does not change confirm payload
  - confirm-create succeeds or fails strictly based on `draft_id`, not recomputed snapshot text

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
git commit -m "docs: close out phase 70a create list from analysis snapshot"
```

## Parallelization Guidance

Safe parallelism:

- Gemini Task 1 and Windsurf Task 3 can begin in parallel because the contract is already frozen.
- Gemini Task 2 depends on Task 1's draft-generation shape.
- Windsurf Task 4 depends on Task 3's shared Web types.
- Windsurf Task 5 depends on Task 4's in-Dock draft preview.
- Windsurf Task 6 is the final handoff/navigation layer.
- Codex Task 7 happens after Gemini and Windsurf complete their slices.

## Verification Checklist

- [ ] `Create list from plan` appears only for valid structured analysis snapshots
- [ ] Draft generation requires `analysis_snapshot_id`
- [ ] Draft generation returns immutable `draft_id`
- [ ] Draft preview stays inside Dock
- [ ] Confirm-create submits only `draft_id`
- [ ] Changing Dock analysis state after draft open does not change the confirm payload
- [ ] One list with many items is created
- [ ] Default channel matches current canvas channel
- [ ] Failure does not show false success and does not lose analysis context

## Non-Goals Reminder

Do not add in this phase:

- conversion from arbitrary canvas text
- prose reparsing from inserted plan text
- auto-create without confirmation
- multiple-list generation
- assignee/due-date/priority inference
- automatic workflow/tool execution
