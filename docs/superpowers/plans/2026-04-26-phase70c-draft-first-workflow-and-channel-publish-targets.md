# Phase 70C Draft-First Workflow And Channel Publish Targets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend Relay's draft-first AI execution model so `workflow` and `channel_message` targets can generate reviewable drafts inside Canvas AI Dock and only create/publish after explicit confirmation.

**Architecture:** Build on Phase 70B's typed execution-target contract and Phase 70A's draft-first execution discipline. Gemini adds shallow but typed workflow/message draft generation plus immutable draft identities and explicit confirm endpoints. Windsurf lands the full execution UX in Canvas AI Dock while keeping `/ask` and DM contract-compatible without heavy execution UI.

**Tech Stack:** Next.js 16, React 19, TypeScript, shared execution-target normalization, Go/Gin API, GORM, workflow/list/channel domain models, frontend component tests, Go handler/domain tests.

---

## File Structure

### Existing files likely involved

- `apps/web/lib/execution-targets.ts`
  - shared target normalization and inheritance helper from Phase 70B
- `apps/web/components/layout/canvas-ai-dock.tsx`
  - Canvas-first execution surface
- `apps/web/components/canvas/file-group-analysis-result.tsx`
  - likely entry point for target-specific execution affordances
- `apps/web/lib/analysis-list-draft.ts`
  - inspect patterns for draft identity and confirm flow
- `apps/web/stores/list-store.ts`
  - inspect existing success/navigation handoff patterns
- `apps/api/internal/handlers/ai.go`
  - likely draft-generation route surface
- `apps/api/internal/handlers/structured_workspace.go`
  - inspect existing workflow/list creation helpers if relevant
- `apps/api/internal/domain/*`
  - inspect workflow/message/list persistence helpers

### New files recommended

- `apps/web/lib/execution-drafts.ts`
  - shared Phase 70C request/response types for workflow/message draft and confirm actions
- `apps/web/components/canvas/workflow-draft-preview.tsx`
  - Dock-local workflow draft preview
- `apps/web/components/canvas/channel-message-draft-preview.tsx`
  - Dock-local message draft preview
- `apps/api/internal/handlers/phase70c_workflow_test.go`
  - workflow draft + confirm tests
- `apps/api/internal/handlers/phase70c_message_test.go`
  - message draft + confirm-publish tests

## Contract Freeze

- **Codex authority:** Phase 70C extends Phase 70B with two new draft-first execution families:
  - `workflow`
  - `channel_message`
- conceptual workflow target shape:
  ```json
  {
    "type": "workflow",
    "workflow_draft": {
      "title": "string",
      "goal": "string",
      "steps": [
        { "title": "string" }
      ]
    }
  }
  ```
- conceptual channel_message target shape:
  ```json
  {
    "type": "channel_message",
    "message_draft": {
      "channel_id": "channel-123",
      "body": "string"
    }
  }
  ```
- draft identity rules:
  - workflow draft generation returns immutable workflow draft ID
  - message draft generation returns immutable message draft ID
  - confirm-create/publish consumes only that immutable draft ID
- Canvas-first rollout:
  - Canvas gets real execution UX
  - `/ask` and DM remain contract-compatible but light

## Task 1: Backend Workflow Draft Contract

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Create/Modify: `apps/api/internal/handlers/phase70c_workflow_test.go`

- [ ] **Step 1: Write the failing workflow draft tests**

Add tests covering:
- valid `workflow` target -> returns immutable workflow draft ID plus shallow draft payload
- missing required workflow draft fields -> rejected or no execution draft
- first-release step schema remains title-only

- [ ] **Step 2: Run the focused workflow draft tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70CWorkflow -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement minimal workflow draft-generation contract**

Requirements:
- accept valid resolved workflow target from a frozen analysis context
- return immutable workflow draft ID
- keep payload shallow:
  - title
  - goal
  - ordered title-only steps

- [ ] **Step 4: Re-run the focused workflow draft tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers
git commit -m "feat(api): add phase 70c workflow draft contract"
```

## Task 2: Backend Workflow Confirm-Create Flow

**Owner:** Gemini  
**Files:**
- Modify: workflow/domain creation helpers under `apps/api/internal/domain/`
- Modify/Create: `apps/api/internal/handlers/phase70c_workflow_test.go`

- [ ] **Step 1: Write the failing workflow confirm-create tests**

Add tests covering:
- valid workflow draft ID -> creates workflow only after confirm
- invalid/missing workflow draft ID -> rejected
- no recompute from mutable client text
- failure path does not show success-like result

- [ ] **Step 2: Run the focused workflow confirm tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase70CWorkflow(Create|Confirm)' -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement minimal workflow confirm-create**

Requirements:
- consume immutable workflow draft ID only
- no auto-create during draft generation
- no full builder logic in first release

- [ ] **Step 4: Re-run the focused workflow confirm tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/internal/domain
git commit -m "feat(api): add phase 70c workflow confirm-create"
```

## Task 3: Backend Channel Message Draft And Confirm-Publish

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Create/Modify: `apps/api/internal/handlers/phase70c_message_test.go`

- [ ] **Step 1: Write the failing message draft/publish tests**

Add tests covering:
- valid `channel_message` target -> returns immutable message draft ID plus shallow payload
- confirm-publish uses only immutable message draft ID
- missing required message draft fields -> rejected or no draft
- no auto-publish during draft generation
- no recompute from mutable client text or live Dock state at confirm-publish time

- [ ] **Step 2: Run the focused message tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70CMessage -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement minimal message draft/publish flow**

Requirements:
- draft payload remains shallow:
  - channel_id
  - body
- confirm-publish consumes immutable message draft ID only
- do not recompute publish body from mutable client text at confirm time
- first release stays single-channel

- [ ] **Step 4: Re-run the focused message tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers
git commit -m "feat(api): add phase 70c message draft publish flow"
```

## Task 4: Shared Web Draft Contract Types

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/execution-drafts.ts`
- Inspect/Modify: `apps/web/lib/execution-targets.ts`
- Test: helper/type tests following repo convention

- [ ] **Step 1: Write the failing Web draft-contract tests**

Add tests for:
- workflow draft response shape + immutable workflow draft ID
- channel message draft response shape + immutable message draft ID
- confirm-create/publish request shapes consume only draft IDs

- [ ] **Step 2: Run the focused helper tests to verify failure**

Run the narrowest test command for the new helper test file.

Expected:
- FAIL

- [ ] **Step 3: Implement shared draft-contract module**

Requirements:
- define workflow draft types
- define message draft types
- define confirm-create/publish request types
- keep Canvas execution UI bound to these shared types

- [ ] **Step 4: Re-run the focused helper tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/execution-drafts.ts apps/web
git commit -m "feat(web): add phase 70c execution draft types"
```

## Task 5: Canvas Workflow Draft Preview And Confirm-Create UX

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/components/canvas/workflow-draft-preview.tsx`
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify: `apps/web/components/canvas/file-group-analysis-result.tsx`
- Test: workflow draft preview and confirm tests

- [ ] **Step 1: Write the failing workflow UX tests**

Minimum cases:
- valid workflow target exposes workflow draft action in Canvas
- workflow draft preview shows title, goal, ordered steps
- confirm-create uses only immutable workflow draft ID
- failure keeps draft visible and unexecuted

- [ ] **Step 2: Run the focused workflow UX tests to verify failure**

Run the narrowest Canvas workflow tests.

Expected:
- FAIL

- [ ] **Step 3: Implement workflow draft UX**

Requirements:
- Dock-local draft preview
- explicit confirm-create
- success/failure state without losing analysis context
- no full workflow builder expansion

- [ ] **Step 4: Re-run the focused workflow UX tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/canvas/workflow-draft-preview.tsx apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas/file-group-analysis-result.tsx apps/web
git commit -m "feat(web): add phase 70c workflow draft preview"
```

## Task 6: Canvas Channel Message Draft Preview And Confirm-Publish UX

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/components/canvas/channel-message-draft-preview.tsx`
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Test: channel-message draft preview and publish tests

- [ ] **Step 1: Write the failing message UX tests**

Minimum cases:
- valid channel_message target exposes draft action in Canvas
- message draft preview shows target channel and message body
- confirm-publish uses only immutable message draft ID
- failure keeps draft visible and unpublished

- [ ] **Step 2: Run the focused message UX tests to verify failure**

Run the narrowest Canvas message tests.

Expected:
- FAIL

- [ ] **Step 3: Implement message draft UX**

Requirements:
- Dock-local draft preview
- explicit confirm-publish
- single-channel first-release behavior
- success/failure state without losing analysis context

- [ ] **Step 4: Re-run the focused message UX tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/canvas/channel-message-draft-preview.tsx apps/web/components/layout/canvas-ai-dock.tsx apps/web
git commit -m "feat(web): add phase 70c channel message draft preview"
```

## Task 7: `/ask` And AI DM Contract-Compatible Target Deepening

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/message/message-item.tsx`
- Modify: `apps/web/app/workspace/dms/[id]/page.tsx`
- Test: `/ask` and DM compatibility tests

- [ ] **Step 1: Write the failing compatibility tests**

Minimum cases:
- `/ask` remains render-safe when workflow/message draft-capable targets appear
- AI DM remains render-safe when workflow/message draft-capable targets appear
- neither surface grows heavy execution UI in this phase
- both surfaces preserve the shared normalized target semantics without renaming or reinterpretation

- [ ] **Step 2: Run the focused compatibility tests to verify failure**

Run the narrowest `/ask` and DM tests.

Expected:
- FAIL

- [ ] **Step 3: Implement lightweight compatibility consumption**

Requirements:
- consume deepened target contract safely
- keep interaction light outside Canvas
- no protocol fork
- preserve the same normalized draft-target semantics used by Canvas even when execution UI is lighter

- [ ] **Step 4: Re-run the focused compatibility tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/message/message-item.tsx apps/web/app/workspace/dms/[id]/page.tsx apps/web
git commit -m "feat(web): add phase 70c dm and ask draft-target compatibility"
```

## Task 8: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md`
- Reference: `docs/superpowers/specs/2026-04-26-phase70c-draft-first-workflow-and-channel-publish-targets-design.md`

- [ ] **Step 1: Review Gemini's backend diff against the frozen Phase 70C contract**

Verify:
- shallow workflow/message payloads only
- immutable draft IDs for both families
- confirm-create/publish consumes only immutable draft IDs
- no auto-execution

- [ ] **Step 2: Review Windsurf's Web diff against the frozen Phase 70C contract**

Verify:
- Canvas gets the full execution UX
- `/ask` and DM stay contract-compatible but lighter
- draft previews remain Dock-local
- success/failure states preserve analysis context

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase70CWorkflow|TestPhase70CMessage' -count=1
```

Web:
- run focused checks covering:
  - workflow draft preview
  - message draft preview
  - confirm-create workflow
  - confirm-publish message
  - `/ask` compatibility
  - DM compatibility
  - `/ask` and DM preserve the same normalized target semantics without surface-specific reinterpretation

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
git commit -m "docs: close out phase 70c workflow and channel publish targets"
```

## Parallelization Guidance

Safe parallelism:

- Gemini Task 1 and Windsurf Task 4 can begin in parallel because the draft contract is frozen.
- Gemini Task 2 depends on Task 1's workflow draft shape.
- Gemini Task 3 can proceed in parallel with Task 2 once the shared draft-first patterns are clear.
- Windsurf Task 5 depends on Task 4's shared draft types.
- Windsurf Task 6 can proceed in parallel with Task 5 after Task 4 lands.
- Windsurf Task 7 follows once the deeper target contract is stable on Web.
- Codex Task 8 happens after Gemini and Windsurf complete their slices.

## Verification Checklist

- [ ] Workflow targets can generate Dock-local drafts
- [ ] Channel_message targets can generate Dock-local drafts
- [ ] Both draft families use immutable draft IDs
- [ ] Confirm-create/publish consumes only immutable draft IDs
- [ ] Canvas owns the full execution UX
- [ ] `/ask` and DM stay contract-compatible without heavy execution UI or semantic drift
- [ ] Failure does not show false success and does not lose analysis context
- [ ] Draft-first execution grammar is now consistent across list/workflow/channel_message

## Non-Goals Reminder

Do not add in this phase:

- full workflow builder UX
- advanced parameter binding
- multi-channel publish
- auto-execution or auto-publish
- prose re-parsing from arbitrary edited text
