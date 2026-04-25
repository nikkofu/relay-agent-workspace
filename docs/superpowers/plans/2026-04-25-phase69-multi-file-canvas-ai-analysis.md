# Phase 69 Multi-File Canvas AI Analysis Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let Canvas AI Dock analyze the persisted multi-file working set already assembled in a canvas and return a structured preview with summary, observations, and next-step suggestions that users can selectively insert into the canvas.

**Architecture:** Build on Phase 68's persisted `file_ref` blocks. Windsurf detects the file group from the active canvas, snapshots it at trigger time, and sends a structured multi-file analysis request. Gemini aggregates the referenced files, degrades gracefully across mixed extraction quality, and returns a structured analysis result. The Dock renders preview-only AI output until the user explicitly inserts a section into the canvas document.

**Tech Stack:** Next.js 16, React 19, TypeScript, Zustand stores, TipTap canvas editor, Go/Gin API, GORM, LLM integration layer, frontend component tests, Go handler/domain tests.

---

## File Structure

### Existing files likely involved

- `apps/web/components/layout/canvas-ai-dock.tsx`
  - primary Dock UI, request trigger point, and analysis preview container
- `apps/web/components/layout/canvas-panel.tsx`
  - canvas shell and artifact-scoped coordination
- `apps/web/components/layout/canvas-tiptap-editor.tsx`
  - insertion target for approved summary / observations / plan sections
- `apps/web/stores/artifact-store.ts`
  - persisted canvas/artifact document access and mutation
- `apps/web/lib/ai-sidecar.ts`
  - inspect only if shared streaming/result normalization patterns are worth reusing
- `apps/api/internal/handlers/ai.go`
  - likely entry point for request/response delivery if Phase 69 extends existing AI routes
- `apps/api/internal/handlers/artifacts.go`
  - inspect for artifact/canvas context helpers
- `apps/api/internal/fileindex/service.go`
  - inspect for file extraction/content retrieval helpers
- `apps/api/internal/llm/*`
  - likely home for structured multi-file prompt orchestration or response shaping

### New files recommended

- `apps/web/lib/canvas-file-group.ts`
  - helper to extract persisted `file_ref` blocks from active canvas document and create the trigger-time snapshot
- `apps/web/lib/multi-file-analysis.ts`
  - shared Web-side request/response types for Phase 69, kept faithful to the Codex-frozen contract
- `apps/web/components/canvas/file-group-analysis-result.tsx`
  - presentation component for summary / observations / next steps with per-section insertion actions
- `apps/api/internal/handlers/phase69_test.go`
  - focused backend contract tests for request validation and response shape
- `apps/api/internal/handlers/phase69_degradation_test.go`
  - focused backend degradation tests for mixed-quality file groups
- `apps/api/internal/llm/multi_file_canvas_analysis.go`
  - if current AI orchestration code does not already have a clean home for this structured operation

## Contract Freeze

- **Codex authority:** Phase 69 is frozen around persisted `file_ref` blocks as the only multi-file source of truth.
- trigger-time request concept:
  ```json
  {
    "artifact_id": "artifact-123",
    "file_refs": [
      { "file_id": "file-1" },
      { "file_id": "file-2" }
    ],
    "mode": "multi_file_analysis"
  }
  ```
- structured response concept:
  ```json
  {
    "analysis": {
      "summary": "string",
      "observations": ["string"],
      "next_steps": [
        {
          "text": "string",
          "rationale": "string",
          "action_hint": "summarize|compare|decide|share|plan|investigate|custom"
        }
      ]
    }
  }
  ```
- `file_refs[]` is a trigger-time snapshot.
- Dock result is preview state until explicit insertion.
- inserted content is copied from the frozen result snapshot for that analysis run, not re-derived from mutable UI state.
- Web canonical type source:
  - Windsurf should define and consume one shared Phase 69 Web contract module in `apps/web/lib/multi-file-analysis.ts`
  - that module must mirror the Codex-frozen response shape:
    - `summary`
    - `observations[]`
    - `next_steps[].text`
    - `next_steps[].rationale`
    - `next_steps[].action_hint`
  - Dock UI and insertion logic should consume that shared type rather than ad hoc local shapes

## Task 1: Backend Multi-File Analysis Contract And Tests

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Create/Modify: `apps/api/internal/handlers/phase69_test.go`
- Inspect: existing AI route wiring in `apps/api/main.go`

- [ ] **Step 1: Write the failing backend contract tests**

Add tests covering:
- valid multi-file request with `artifact_id` + snapshotted `file_refs[]`
- invalid request with empty or malformed `file_refs[]`
- invalid request shaped like a raw drag payload must be rejected
- stable structured response shape containing:
  - `summary`
  - `observations[]`
  - `next_steps[].text`
  - `next_steps[].rationale`
  - `next_steps[].action_hint`

- [ ] **Step 2: Run the focused backend tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase69 -count=1
```

Expected:
- FAIL until the route/schema logic exists

- [ ] **Step 3: Implement the minimal request/response contract**

Requirements:
- accept trigger-time `artifact_id` + `file_refs[]`
- reject raw drag-payload-shaped input
- return a structured analysis object, not a prose blob

- [ ] **Step 4: Re-run the focused backend tests**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase69 -count=1
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/main.go
git commit -m "feat(api): add phase 69 multi-file analysis contract"
```

## Task 2: Backend File-Group Aggregation And Graceful Degradation

**Owner:** Gemini  
**Files:**
- Modify: AI orchestration files under `apps/api/internal/llm/`
- Inspect/Modify: `apps/api/internal/fileindex/service.go`
- Test: `apps/api/internal/handlers/phase69_degradation_test.go` or a focused domain test file

- [ ] **Step 1: Write the failing degradation tests**

Add tests for:
- all files available -> normal structured result
- some files missing extraction text -> still returns structured result
- large or partial file groups -> still returns either degraded structured result or explicit user-readable failure

- [ ] **Step 2: Run the focused tests to verify failure**

Run the narrowest handler/domain test command covering the new degradation scenarios.

Expected:
- FAIL

- [ ] **Step 3: Implement aggregation logic**

Requirements:
- resolve file group from persisted `file_id`s
- gather extraction text, metadata, and any fallback inputs that already exist
- tolerate partial file quality instead of failing the whole request whenever reasonable
- preserve the snapshotted request semantics

- [ ] **Step 4: Re-run the focused tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/phase69_degradation_test.go apps/api/internal/handlers apps/api/internal/llm apps/api/internal/fileindex
git commit -m "feat(api): add phase 69 file-group analysis degradation handling"
```

## Task 3: Web File-Group Snapshot Helper

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/canvas-file-group.ts`
- Create: `apps/web/lib/multi-file-analysis.ts`
- Inspect/Modify: `apps/web/stores/artifact-store.ts`
- Test: helper test file near the new lib following repo convention

- [ ] **Step 1: Write the failing helper tests**

Test scenarios:
- extract all persisted `file_ref` blocks from active canvas
- produce a stable trigger-time snapshot of `file_id`s
- ignore non-`file_ref` blocks
- preserve order consistently if current canvas document order is used

- [ ] **Step 2: Run the focused web tests to verify failure**

Run the narrowest test command for the helper test file.

Expected:
- FAIL

- [ ] **Step 3: Implement the snapshot helper**

Requirements:
- read persisted canvas document state
- derive `file_refs[]` snapshot at trigger time
- produce one request-ready structure for Dock submission
- define shared Phase 69 request/response types in `apps/web/lib/multi-file-analysis.ts`

- [ ] **Step 4: Re-run the helper tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/canvas-file-group.ts apps/web/lib/multi-file-analysis.ts apps/web/stores/artifact-store.ts apps/web
git commit -m "feat(web): add phase 69 canvas file-group snapshot helper"
```

## Task 4: Dock Trigger State And Structured Result Rendering

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Create: `apps/web/components/canvas/file-group-analysis-result.tsx`
- Test: Dock trigger/render test

- [ ] **Step 1: Write the failing Dock tests**

Minimum scenarios:
- when active canvas has multiple `file_ref` blocks, Dock exposes `Analyze file group`
- when fewer than two `file_ref` blocks exist, Dock does not expose the action
- structured result renders:
  - summary
  - observations
  - next step suggestions

- [ ] **Step 2: Run the focused Dock tests to verify failure**

Run the narrowest test command for the Dock/component tests.

Expected:
- FAIL

- [ ] **Step 3: Implement Dock trigger + preview rendering**

Requirements:
- base trigger visibility on persisted `file_ref` count
- submit the snapshotted file-group request from Task 3
- render preview-only result sections
- keep Dock preview visually distinct from persisted canvas content

- [ ] **Step 4: Re-run the focused Dock tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas/file-group-analysis-result.tsx apps/web
git commit -m "feat(web): add phase 69 dock file-group analysis preview"
```

## Task 5: Per-Section Canvas Insertion From Frozen Result Snapshot

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify: `apps/web/components/layout/canvas-tiptap-editor.tsx`
- Possibly modify: `apps/web/components/layout/canvas-panel.tsx`
- Test: insertion behavior test

- [ ] **Step 1: Write the failing insertion tests**

Minimum cases:
- `Insert summary` inserts only summary content
- `Insert observations` inserts only observations content
- `Insert plan` inserts only next-step plan content
- insertion copies from the frozen returned result snapshot, not mutable live UI text
- inserted content persists as normal canvas content

- [ ] **Step 2: Run the focused insertion tests to verify failure**

Run the narrowest canvas insertion test command.

Expected:
- FAIL

- [ ] **Step 3: Implement section-level insertion**

Requirements:
- no auto-insert after generation
- insert at current cursor when available, otherwise one editor-safe fallback
- preserve clear separation between Dock preview state and persisted document content

- [ ] **Step 4: Re-run the focused insertion tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/layout/canvas-tiptap-editor.tsx apps/web/components/layout/canvas-panel.tsx
git commit -m "feat(canvas): support phase 69 selective analysis insertion"
```

## Task 6: Failure States And Preview/Persisted State Separation

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify: any related analysis preview components
- Test: failure and state-separation tests

- [ ] **Step 1: Write the failing state-separation tests**

Minimum cases:
- analysis failure does not mutate canvas content
- partial backend degradation still renders structured output when returned
- preview remains temporary UI state until explicit insertion
- inserted content remains after refresh even if the Dock preview is gone

- [ ] **Step 2: Run the focused tests to verify failure**

Expected:
- FAIL

- [ ] **Step 3: Implement failure handling and lifecycle separation**

Requirements:
- user-readable Dock error state on total failure
- no implicit insertion from partial or failed analysis
- refresh/reopen does not confuse previously inserted content with temporary preview state

- [ ] **Step 4: Re-run the focused tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas apps/web
git commit -m "fix(web): harden phase 69 analysis preview lifecycle"
```

## Task 7: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md` after implementation lands
- Reference: `docs/superpowers/specs/2026-04-25-phase69-multi-file-canvas-ai-analysis-design.md`

- [ ] **Step 1: Review Gemini's backend diff against the frozen Phase 69 contract**

Verify:
- persisted `file_ref` is still the only source of truth
- trigger-time snapshot semantics are preserved
- response remains structured and stable
- no hidden List/Workflow/Tool side effects were introduced

- [ ] **Step 2: Review Windsurf's Web diff against the frozen Phase 69 contract**

Verify:
- Dock only exposes the action for multi-file canvases
- preview state stays separate from persisted content
- insertion is section-level and explicit
- no auto-write behavior was introduced

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase69 -count=1
```

Backend degradation:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase69.*Degradation' -count=1
```

Web:
- run the focused web checks covering:
  - `canvas-file-group` helper tests
  - `canvas-ai-dock` trigger/render tests
  - section-level insertion tests
  - preview/persisted-state separation tests
- plus static verification:
```bash
cd apps/web && pnpm exec tsc --noEmit
```

- [ ] **Step 4: Update collaboration and release docs**

Update:
- `docs/AGENT-COLLAB.md` with completion notes and follow-up hooks
- `docs/releases/v0.6.xx.md` with concise user-facing release notes

- [ ] **Step 5: Commit**

```bash
git add docs/AGENT-COLLAB.md docs/releases
git commit -m "docs: close out phase 69 multi-file canvas ai analysis"
```

## Parallelization Guidance

Safe parallelism:

- Gemini Task 1 and Windsurf Task 3 can begin in parallel because the conceptual request/response contract is already frozen by the spec.
- Windsurf Task 4 depends on Task 3's snapshot helper.
- Gemini Task 2 can proceed after Task 1 establishes the route/schema.
- Windsurf Task 5 should follow Task 4 once a stable structured preview exists.
- Windsurf Task 6 is the hardening pass after Task 5.
- Codex Task 7 happens after Gemini and Windsurf complete their slices.

## Verification Checklist

- [ ] Active canvas with multiple `file_ref` blocks exposes `Analyze file group`
- [ ] Analysis request uses snapshotted persisted `file_ref` inputs
- [ ] Raw drag-payload-shaped requests are rejected by the backend contract
- [ ] Dock returns summary, observations, and next steps
- [ ] `next_steps[]` includes `text`, `rationale`, and `action_hint`
- [ ] Users can insert summary / observations / plan independently
- [ ] Inserted content persists as normal canvas content
- [ ] Dock preview state is distinct from persisted canvas content
- [ ] Partial file-quality problems degrade analysis instead of forcing unnecessary total failure

## Non-Goals Reminder

Do not add in this phase:

- auto-write AI output into canvas
- direct List item creation
- direct Workflow or Tool execution
- advanced file-group manager UI
- cross-canvas shared analysis sessions
