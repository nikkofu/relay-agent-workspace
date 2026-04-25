# Phase 70B Typed Execution Targets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce one shared typed execution-target contract across all AI reply surfaces, with deterministic inheritance from analysis default target to step-level override, while landing the first real target-driven execution UX in Canvas.

**Architecture:** Extend the shared AI response contract so Canvas AI, channel `/ask`, and AI DM all carry the same typed execution-target object. Freeze resolution rules server-side and client-side. Let Canvas consume the contract deeply first, with `list` as the only mandatory end-to-end execution flow in this release, while `/ask` and DM remain render-compatible without heavy execution UX.

**Tech Stack:** Next.js 16, React 19, TypeScript, shared AI normalization utilities, Go/Gin API, GORM, LLM orchestration layer, frontend component tests, Go handler/domain tests.

---

## File Structure

### Existing files likely involved

- `apps/web/lib/ai-sidecar.ts`
  - inspect for shared AI normalization patterns if execution-target typing belongs beside sidecar normalization
- `apps/web/components/layout/canvas-ai-dock.tsx`
  - Canvas-first execution-target display and action entry points
- `apps/web/components/canvas/file-group-analysis-result.tsx`
  - likely renderer for default target and step-level override display
- `apps/web/app/workspace/dms/[id]/page.tsx`
  - AI DM contract-compatible rendering surface
- `apps/web/components/message/message-item.tsx`
  - channel `/ask` rendering surface
- `apps/web/lib/analysis-list-draft.ts`
  - inspect/reuse for `list` target-driven execution path
- `apps/api/internal/handlers/ai.go`
  - likely shared response contract surface
- `apps/api/internal/llm/*`
  - likely AI result shaping/orchestration point

### New files recommended

- `apps/web/lib/execution-targets.ts`
  - shared normalized Web types and deterministic resolution helper
- `apps/api/internal/handlers/phase70b_test.go`
  - backend contract tests for default/override/malformed handling
- `apps/api/internal/handlers/phase70b_surface_test.go`
  - cross-surface response-shape tests for Canvas, `/ask`, and AI DM

## Contract Freeze

- **Codex authority:** Phase 70B is frozen around one shared two-layer execution-target contract:
  - `analysis.default_execution_target`
  - `next_steps[].execution_target`
- allowed first-release target types:
  - `list`
  - `workflow`
  - `channel_message`
- deterministic resolution:
  - step target if present and valid
  - otherwise analysis default if present and valid
  - otherwise no target
- malformed handling:
  - backend should omit invalid targets rather than emit ambiguous partial objects
  - shared Web normalization must treat malformed or unknown targets as absent
- first-release behavior depth:
  - only `type` is authoritative
  - extra payload fields may exist later, must be preserved where practical in normalized transport, and must be ignored for first-release behavior
- first-release mandatory real execution flow:
  - `list`
- first-release lighter but typed targets:
  - `workflow`
  - `channel_message`

## Task 1: Backend Shared Execution-Target Contract

**Owner:** Gemini  
**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Create/Modify: `apps/api/internal/handlers/phase70b_test.go`

- [ ] **Step 1: Write the failing backend contract tests**

Add tests covering:
- valid `analysis.default_execution_target`
- valid `next_steps[].execution_target`
- step override wins over analysis default
- missing step target falls back to analysis default
- missing both produces no target
- Canvas-bearing responses with `list` target remain deterministic inputs for the existing list execution path

- [ ] **Step 2: Run the focused backend tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70B -count=1
```

Expected:
- FAIL until shared target support exists

- [ ] **Step 3: Implement minimal shared target schema**

Requirements:
- add typed target support to shared AI response DTOs
- keep `action_hint` separate from target
- resolve only allowed target types
- keep Canvas `list` target output stable enough for the existing Phase 70A list flow to consume deterministically

- [ ] **Step 4: Re-run the focused backend tests**

Run:
```bash
cd apps/api && go test ./internal/handlers -run TestPhase70B -count=1
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers
git commit -m "feat(api): add phase 70b execution target contract"
```

## Task 2: Backend Malformed-Target And Cross-Surface Stability

**Owner:** Gemini  
**Files:**
- Modify/Create: `apps/api/internal/handlers/phase70b_surface_test.go`
- Inspect/Modify: shared AI response shaping code under `apps/api/internal/llm/`

- [ ] **Step 1: Write the failing stability tests**

Add tests covering:
- malformed step target is omitted rather than emitted ambiguously
- malformed analysis default target is omitted
- Canvas, `/ask`, and AI DM responses share the same normalized target contract shape
- extra future payload fields do not become authoritative in first release
- extra future payload fields remain preserved in transport where practical

- [ ] **Step 2: Run the focused stability tests to verify failure**

Run:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase70B(Surfaces|Malformed)' -count=1
```

Expected:
- FAIL

- [ ] **Step 3: Implement malformed handling and surface unification**

Requirements:
- omit invalid targets server-side
- keep target schema identical across AI reply surfaces
- preserve compatibility for surfaces with lighter UX
- preserve future extra target fields in the response shape where practical without making them authoritative in first release

- [ ] **Step 4: Re-run the focused stability tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers apps/api/internal/llm
git commit -m "feat(api): harden phase 70b target inheritance and surfaces"
```

## Task 3: Shared Web Execution-Target Types And Resolution Helper

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/execution-targets.ts`
- Inspect/Modify: `apps/web/lib/ai-sidecar.ts`
- Test: helper tests following repo convention

- [ ] **Step 1: Write the failing Web helper tests**

Add tests for:
- resolve step override over analysis default
- resolve analysis default when step target absent
- malformed/unknown targets normalize to absent
- extra payload fields are ignored for first-release behavior
- extra payload fields are preserved in normalized output where practical

- [ ] **Step 2: Run the focused helper tests to verify failure**

Run the narrowest test command for the new helper test file.

Expected:
- FAIL

- [ ] **Step 3: Implement shared Web execution-target module**

Requirements:
- define normalized target types
- define deterministic resolution helper
- normalize malformed targets to absent
- expose one shared contract for Canvas, `/ask`, and DM consumption
- preserve unknown extra target fields in normalized transport where practical while keeping first-release behavior keyed only on `type`

- [ ] **Step 4: Re-run the focused helper tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/execution-targets.ts apps/web
git commit -m "feat(web): add phase 70b execution target normalization"
```

## Task 4: Canvas Target Display And Required `list` Execution Flow

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify: `apps/web/components/canvas/file-group-analysis-result.tsx`
- Inspect/Modify: `apps/web/lib/analysis-list-draft.ts`
- Test: Canvas target rendering and action tests

- [ ] **Step 1: Write the failing Canvas tests**

Minimum scenarios:
- analysis default target displays correctly
- step override target displays correctly
- inherited target displays correctly
- `list` target exposes a real execution path in Canvas

- [ ] **Step 2: Run the focused Canvas tests to verify failure**

Run the narrowest Canvas/component tests.

Expected:
- FAIL

- [ ] **Step 3: Implement Canvas-first target UX**

Requirements:
- render default and step-level target states
- show `list` as the mandatory real execution path
- keep workflow and channel_message clearly typed even if their UX is lighter

- [ ] **Step 4: Re-run the focused Canvas tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas apps/web/lib/analysis-list-draft.ts
git commit -m "feat(web): add phase 70b canvas target display"
```

## Task 5: `/ask` And AI DM Contract-Compatible Consumption

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/message/message-item.tsx`
- Modify: `apps/web/app/workspace/dms/[id]/page.tsx`
- Test: `/ask` and DM contract-compatibility tests

- [ ] **Step 1: Write the failing compatibility tests**

Minimum scenarios:
- `/ask` can render AI replies with typed execution targets without breaking
- AI DM can render AI replies with typed execution targets without breaking
- neither surface forks or renames target semantics
- Canvas, `/ask`, and DM all consume the same normalized target contract shape before surface-specific rendering

- [ ] **Step 2: Run the focused compatibility tests to verify failure**

Run the narrowest `/ask` and DM tests.

Expected:
- FAIL

- [ ] **Step 3: Implement lightweight contract-compatible rendering**

Requirements:
- consume the shared normalized target contract unchanged
- keep UI light
- do not add heavy execution flows on these surfaces in this phase

- [ ] **Step 4: Re-run the focused compatibility tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/message/message-item.tsx apps/web/app/workspace/dms/[id]/page.tsx apps/web
git commit -m "feat(web): add phase 70b dm and ask target compatibility"
```

## Task 6: Unsupported-Target UX And No-Guessing Hardening

**Owner:** Windsurf  
**Files:**
- Modify: Canvas target rendering components
- Modify: any shared AI rendering helpers if needed
- Test: no-guessing and unsupported-target tests

- [ ] **Step 1: Write the failing hardening tests**

Minimum cases:
- no target at either layer -> no execution route shown
- malformed target -> treated as absent
- unsupported target type in current Canvas UX remains visible but not falsely executable

- [ ] **Step 2: Run the focused hardening tests to verify failure**

Expected:
- FAIL

- [ ] **Step 3: Implement no-guessing/unsupported-target hardening**

Requirements:
- no prose-based inference
- no false execution affordance
- preserve typed visibility where UX is lighter

- [ ] **Step 4: Re-run the focused hardening tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-ai-dock.tsx apps/web/components/canvas apps/web/lib/execution-targets.ts
git commit -m "fix(web): harden phase 70b target resolution"
```

## Task 7: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md`
- Reference: `docs/superpowers/specs/2026-04-26-phase70b-typed-execution-targets-design.md`

- [ ] **Step 1: Review Gemini's backend diff against the frozen Phase 70B contract**

Verify:
- same target schema across Canvas, `/ask`, and DM
- malformed targets are omitted or normalized away consistently
- `action_hint` stays separate from `execution_target`

- [ ] **Step 2: Review Windsurf's Web diff against the frozen Phase 70B contract**

Verify:
- one shared normalized contract is consumed everywhere
- Canvas-first interaction does not fork protocol semantics
- `list` is the mandatory real execution flow
- unsupported targets remain visible without false execution

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase70B|TestPhase70B(Surfaces|Malformed)' -count=1
```

Web:
- run focused checks covering:
  - one shared normalized contract shape consumed by Canvas, `/ask`, and DM before surface-specific rendering
  - execution-target normalization
  - Canvas target display
  - list execution flow visibility
  - `/ask` compatibility
  - DM compatibility
  - no-guessing behavior

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
git commit -m "docs: close out phase 70b typed execution targets"
```

## Parallelization Guidance

Safe parallelism:

- Gemini Task 1 and Windsurf Task 3 can begin in parallel because the contract is frozen.
- Gemini Task 2 depends on Task 1's target schema.
- Windsurf Task 4 depends on Task 3's shared normalized contract.
- Windsurf Task 5 can begin after Task 3 even if Task 4 is still in progress, because `/ask` and DM are light-consumption surfaces.
- Windsurf Task 6 should follow after Task 4/5 harden the actual behaviors.
- Codex Task 7 happens after Gemini and Windsurf complete their slices.

## Verification Checklist

- [ ] All AI reply surfaces may return the same typed execution-target contract
- [ ] Canvas, `/ask`, and DM all consume one shared normalized contract shape before surface-specific rendering
- [ ] `analysis.default_execution_target` and `next_steps[].execution_target` both work
- [ ] step override wins over analysis default
- [ ] malformed targets normalize to absent
- [ ] unknown extra target fields are preserved where practical but ignored for first-release behavior
- [ ] Canvas shows default, override, and inherited targets
- [ ] Canvas has a real end-to-end `list` target-driven execution path
- [ ] `/ask` and DM consume the same target semantics without heavy UX forks
- [ ] missing targets never become guessed actions

## Non-Goals Reminder

Do not add in this phase:

- full execution UX parity across all AI surfaces
- automatic execution
- prose-based target inference
- deep workflow builder UX
- arbitrary nested execution routing
