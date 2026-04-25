# Phase 68 File Archive + Canvas Convergence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users drag files from Files, message attachments, and search results into an open canvas, inserting one persisted compact `file_ref` card at the current cursor position.

**Architecture:** Reuse the existing file archive and canvas systems. Normalize every file drag source into one frozen `file-to-canvas` client payload, then persist one source-agnostic `file_ref` block inside the canvas document. Keep file archive as the source of truth and route card clicks back into existing file preview/detail UX.

**Tech Stack:** Next.js 16, React 19, TypeScript, Zustand stores, TipTap canvas editor, Go/Gin API, GORM, Vitest/Jest-style frontend tests, Go handler/unit tests.

---

## File Structure

### Existing files likely involved

- `apps/web/app/workspace/files/page.tsx`
  - Files archive UI and the primary drag source entry point.
- `apps/web/components/message/file-attachment-card.tsx`
  - Rich message attachment UI; secondary drag source.
- `apps/web/components/layout/canvas-panel.tsx`
  - Canvas shell and likely coordination point for editor-open state and preview navigation.
- `apps/web/components/layout/canvas-tiptap-editor.tsx`
  - TipTap editor integration and the likely insertion/drop target.
- `apps/web/stores/file-store.ts`
  - Existing file data fetching and preview access.
- `apps/web/stores/artifact-store.ts`
  - Existing canvas/artifact persistence surface.
- `apps/api/internal/handlers/files.go`
  - Existing file response shapes and preview/detail routes.
- `apps/api/internal/handlers/artifacts.go`
  - Existing artifact persistence if canvas block schema support needs backend normalization.

### New files recommended

- `apps/web/lib/file-to-canvas.ts`
  - Shared `file-to-canvas` payload types, normalization helpers, drag payload encode/decode helpers.
- `apps/web/components/canvas/file-ref-card.tsx`
  - Compact canvas card for persisted `file_ref` blocks.
- `apps/web/components/search/file-search-result-row.tsx`
  - Only if search results are not already in a focused reusable component; otherwise modify the existing file result row component directly.
- `apps/api/internal/handlers/files_drag_payload_test.go`
  - Only if Gemini introduces backend-side normalization helpers worth unit testing as isolated logic.

### Contract freeze

- **Codex authority:** the Phase 68 shared client drag schema is frozen by the spec:
  - drag-time payload:
    ```json
    {
      "kind": "file-to-canvas",
      "file": {
        "id": "file-123",
        "title": "Quarterly Plan.pdf",
        "mime_type": "application/pdf",
        "size": 245120,
        "preview_url": "/api/v1/files/file-123/preview"
      },
      "source": "files"
    }
    ```
  - persisted canvas block:
    ```json
    {
      "type": "file_ref",
      "file_id": "file-123",
      "title": "Quarterly Plan.pdf",
      "mime_type": "application/pdf",
      "size": 245120,
      "preview_url": "/api/v1/files/file-123/preview"
    }
    ```
- Gemini may stabilize source file payloads.
- Windsurf must normalize every surface into the frozen drag payload before canvas insertion.

## Task 1: Backend Payload Audit And Stabilization

**Owner:** Gemini  
**Files:**
- Inspect: `apps/api/internal/handlers/files.go`
- Inspect: `apps/api/internal/handlers/artifacts.go`
- Inspect: any file/search/message attachment response builders touched by the current web surfaces
- Test: `apps/api/internal/handlers/*_test.go`

- [ ] **Step 1: Write the failing backend contract tests**

Add targeted tests for whichever existing handlers provide file objects to:
- Files page
- message attachment cards
- file search results

Minimum assertions:
- file ID is stable
- title/name is present
- mime type is present or derivable as a stable string
- size is present
- preview/detail route is present or canonically derivable

- [ ] **Step 2: Run targeted backend tests to verify the gap**

Run:
```bash
cd apps/api && go test ./internal/handlers -run 'Test(FilePayload|FileSearch|MessageFiles)' -count=1
```

Expected:
- at least one failing assertion or a clear proof that the current surfaces already expose enough stable data

- [ ] **Step 3: Implement the minimal backend normalization needed**

Allowed changes:
- add or normalize lightweight file fields on existing responses
- factor a shared serializer/helper if three surfaces currently drift

Forbidden changes:
- no dedicated `POST /canvas/file-ref`
- no new orchestration endpoint
- no AI file transformation behavior

- [ ] **Step 4: Re-run targeted backend tests**

Run:
```bash
cd apps/api && go test ./internal/handlers -run 'Test(FilePayload|FileSearch|MessageFiles)' -count=1
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers
git commit -m "feat(api): stabilize phase 68 file payloads"
```

## Task 2: Shared Web Drag Contract

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/lib/file-to-canvas.ts`
- Test: web unit test file near the lib, following existing repo convention

- [ ] **Step 1: Write the failing normalization tests**

Add tests for:
- Files page file shape -> normalized drag payload
- message attachment file shape -> normalized drag payload
- search file result shape -> normalized drag payload
- invalid input -> explicit rejection

Exact assertions:
- output `kind === "file-to-canvas"`
- output `file.id/title/mime_type/size` present
- `source` preserved only as informational metadata

- [ ] **Step 2: Run the web unit tests to verify failure**

Run the narrowest existing web test command for the new file-contract test.
If there is no single-file command in repo scripts, use the repo's established `pnpm test` variant for one spec file.

Expected:
- FAIL until the helper exists

- [ ] **Step 3: Implement the normalization helper**

Implement:
- Type definitions for drag payload and persisted `file_ref`
- source-specific normalizers
- one source-agnostic encoder/decoder for drag transfer
- strict rejection for raw per-surface shapes reaching canvas insertion

- [ ] **Step 4: Re-run the same web unit tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/lib/file-to-canvas.ts apps/web
git commit -m "feat(web): add shared phase 68 file-to-canvas contract"
```

## Task 3: Files Page As Primary Drag Source

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/app/workspace/files/page.tsx`
- Possibly modify: `apps/web/stores/file-store.ts`
- Test: Files page drag interaction test following existing conventions

- [ ] **Step 1: Write the failing Files-page drag test**

Test behavior:
- a file row/card can start a drag
- drag data contains the normalized `file-to-canvas` payload
- no source-specific raw object leaks into the drag payload

- [ ] **Step 2: Run the Files-page test to verify failure**

Run the narrowest test command for the affected test file.

Expected:
- FAIL

- [ ] **Step 3: Implement primary drag-source wiring**

Add:
- draggable affordance on file rows/cards
- normalized payload generation using `apps/web/lib/file-to-canvas.ts`

Do not add:
- canvas-specific logic into Files page

- [ ] **Step 4: Re-run the Files-page test**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/app/workspace/files/page.tsx apps/web/stores/file-store.ts apps/web
git commit -m "feat(web): add phase 68 files-page drag source"
```

## Task 4: Message Attachment And Search As Secondary Drag Sources

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/message/file-attachment-card.tsx`
- Modify: existing search result component that renders file hits
- Test: message attachment drag test
- Test: search file result drag test

- [ ] **Step 1: Write the failing secondary-source tests**

Test behavior:
- file attachment card emits the normalized drag payload
- file search result emits the normalized drag payload
- both payloads match the same schema as Files page for the same source file

- [ ] **Step 2: Run the tests to verify failure**

Run the narrowest tests for the touched components.

Expected:
- FAIL

- [ ] **Step 3: Implement drag-source reuse**

Use the shared normalization helper from Task 2.

Do not:
- fork drag schema logic per component
- add separate canvas insertion code for these sources

- [ ] **Step 4: Re-run the tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/message/file-attachment-card.tsx apps/web/components apps/web
git commit -m "feat(web): add phase 68 attachment and search drag sources"
```

## Task 5: Canvas Drop Handling And Persisted `file_ref` Blocks

**Owner:** Windsurf  
**Files:**
- Modify: `apps/web/components/layout/canvas-tiptap-editor.tsx`
- Modify: `apps/web/components/layout/canvas-panel.tsx`
- Possibly modify: `apps/web/stores/artifact-store.ts`
- Test: canvas drop/insertion test
- Test: persistence/reload test

- [ ] **Step 1: Write the failing canvas insertion tests**

Minimum scenarios:
- valid drop at active selection inserts exactly one `file_ref` block
- missing active selection inserts exactly once at the safe fallback insertion point
- invalid payload is rejected and document content is unchanged
- no open canvas rejects safely
- persisted canvas reload re-renders the same `file_ref` block

- [ ] **Step 2: Run the canvas tests to verify failure**

Run the narrowest canvas editor test command.

Expected:
- FAIL

- [ ] **Step 3: Implement drop handling**

Implementation requirements:
- accept only the normalized drag payload
- map payload to the frozen persisted `file_ref` block shape
- insert into the current selection when valid
- use one editor-defined safe fallback insertion point when no selection exists
- abort without document mutation on invalid drop conditions

- [ ] **Step 4: Re-run the canvas tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout/canvas-tiptap-editor.tsx apps/web/components/layout/canvas-panel.tsx apps/web/stores/artifact-store.ts
git commit -m "feat(canvas): support phase 68 file reference drops"
```

## Task 6: Compact File Reference Card Rendering And Preview Navigation

**Owner:** Windsurf  
**Files:**
- Create: `apps/web/components/canvas/file-ref-card.tsx`
- Modify: the canvas node rendering path that chooses block renderers
- Possibly modify: `apps/web/app/workspace/files/page.tsx` only if preview opening needs shared navigation glue
- Test: file-ref rendering test
- Test: click-through preview navigation test

- [ ] **Step 1: Write the failing renderer tests**

Assertions:
- a persisted `file_ref` block renders as a compact card
- card shows file name, mime/type hint, size, preview/open action
- card click resolves into the existing file preview/detail UX
- no giant inline inspector appears in canvas

- [ ] **Step 2: Run the renderer tests to verify failure**

Run the narrowest renderer/navigation tests.

Expected:
- FAIL

- [ ] **Step 3: Implement the compact card**

Constraints:
- no inline expanded details panel
- no duplicated archive metadata editor
- keep the card visually compact and document-friendly

- [ ] **Step 4: Re-run the renderer tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/canvas/file-ref-card.tsx apps/web/components/layout apps/web/components
git commit -m "feat(canvas): render phase 68 compact file reference cards"
```

## Task 7: Soft-Failure UX And Regression Verification

**Owner:** Windsurf  
**Files:**
- Modify: any drop target UI and editor feedback components touched in Tasks 3-6
- Test: regression tests covering failure hints and document immutability

- [ ] **Step 1: Write the failing failure-mode tests**

Minimum cases:
- invalid drag payload -> user gets lightweight failure hint
- no open canvas -> user gets lightweight failure hint
- rejected drop leaves persisted document JSON unchanged

- [ ] **Step 2: Run the regression tests to verify failure**

Expected:
- FAIL

- [ ] **Step 3: Implement lightweight failure hints and immutability guards**

Requirements:
- every rejected drop gets user feedback
- rejected drops do not alter editor content

- [ ] **Step 4: Re-run the regression tests**

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/layout apps/web/components apps/web
git commit -m "fix(web): harden phase 68 file drop failure handling"
```

## Task 8: Codex Contract Audit, Collaboration Sync, And Release Handoff

**Owner:** Codex  
**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/releases/v0.6.xx.md` once Gemini and Windsurf land the implementation
- Reference: `docs/superpowers/specs/2026-04-25-phase68-file-archive-canvas-convergence-design.md`

- [ ] **Step 1: Review Gemini's backend diff against the frozen spec**

Verify:
- no new orchestration endpoint was introduced
- shared file payload fields are stable enough for Web normalization
- preview/detail routes remain compatible with the canvas card

- [ ] **Step 2: Review Windsurf's Web diff against the frozen spec**

Verify:
- one shared drag contract exists
- all three entry points normalize before canvas insertion
- insertion logic accepts only the normalized payload
- card rendering remains compact

- [ ] **Step 3: Run focused verification commands**

Backend:
```bash
cd apps/api && go test ./internal/handlers -run 'Test(FilePayload|FileSearch|MessageFiles)' -count=1
```

Web:
- run the narrowest repo-appropriate tests covering:
  - drag payload normalization
  - canvas drop insertion
  - persisted `file_ref` reload
  - preview navigation

- [ ] **Step 4: Update collaboration and release docs**

Update:
- `docs/AGENT-COLLAB.md` with completion notes and any follow-up asks
- `docs/releases/v0.6.xx.md` with a concise user-facing description

- [ ] **Step 5: Commit**

```bash
git add docs/AGENT-COLLAB.md docs/releases
git commit -m "docs: close out phase 68 file archive canvas convergence"
```

## Parallelization Guidance

Safe parallelism after Task 1 and Task 2 contract freeze:

- Windsurf Task 3 and Task 4 can run in parallel once the shared Web normalization helper from Task 2 exists.
- Windsurf Task 5 depends on Task 2 and should preferably land after Task 3/4 payload producers are stable.
- Windsurf Task 6 can start once Task 5 proves the persisted `file_ref` shape.
- Windsurf Task 7 should be the final hardening pass.
- Codex Task 8 happens after Gemini and Windsurf implementation slices are complete.

## Verification Checklist

- [ ] Files page drag emits frozen payload
- [ ] Message attachment drag emits frozen payload
- [ ] Search file result drag emits frozen payload
- [ ] Canvas accepts only normalized payloads
- [ ] Valid drop inserts exactly one persisted `file_ref` block
- [ ] Rejected drop leaves document unchanged
- [ ] Reload preserves and re-renders `file_ref`
- [ ] Card click opens existing file preview/detail UX
- [ ] No inline heavy file inspector appears inside canvas

## Non-Goals Reminder

Do not add in this phase:

- automatic file-to-text insertion
- AI summary-on-drop
- multi-file auto-layout
- canvas/file bidirectional content sync
- new dedicated backend endpoint for canvas insertion
