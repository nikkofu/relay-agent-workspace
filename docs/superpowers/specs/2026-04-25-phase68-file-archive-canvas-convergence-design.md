# Phase 68 File Archive + Canvas Convergence Design

## Summary

Phase 68 connects Relay's file archive and canvas surfaces through one narrow first-class workflow:

`Files / message file attachments / search file results -> drag into an open canvas -> insert a persisted file reference card at the current cursor position`

The goal is not to turn files into editable canvas body text. The goal is to make files first-class source objects that can be organized inside canvas without losing their identity, preview affordances, or archive semantics.

## Product Definition

`File archive is the source object. Canvas is the organizing and collaboration surface.`

First release behavior:

- users can drag a file from three entry points:
  - Files page
  - message file attachment cards
  - global search file results
- all three entry points emit the same `file-to-canvas` drag payload
- dropping into an already-open canvas inserts one compact `file_ref` block
- the `file_ref` block renders as a compact file card
- clicking the card opens existing file preview/detail UX

## Scope

### In Scope

- one shared `file-to-canvas` drag contract across Files, message attachments, and search results
- canvas support for inserting a persisted `file_ref` block at the current cursor position
- compact file reference card rendering in canvas
- preview/detail navigation from the card into the existing file UX
- soft-failure handling when no canvas is open, cursor insertion is not possible, or payload validation fails

### Out of Scope

- inserting extracted file body content into canvas
- AI auto-summarization on drop
- bulk multi-file layout rules
- inline expanded file detail UI inside canvas
- bidirectional sync between file metadata edits and canvas block content beyond normal re-rendering
- new top-level routes for file-canvas workflows

## User Experience

### Primary Flow

1. A user opens an existing canvas.
2. The user drags a file from one of three entry points.
3. The canvas editor accepts the drop.
4. Relay inserts a compact file reference card at the current cursor position.
5. The user can continue writing around that card.
6. Clicking the card opens the existing file preview/detail experience.

### Entry Points

#### Files Page

This is the primary drag source and the reference implementation for the drag contract.

#### Message File Attachment Card

This reuses the same drag payload so files already shared in a conversation can be pulled into canvas without reopening the Files page.

#### Search File Result

This also reuses the same drag payload so a searched file can move directly into an open canvas.

### Card Density

The inserted card is intentionally compact:

- file name
- file type / mime hint
- file size
- preview/open action

The card should not expand into a second full file detail panel inside canvas.

## System Design

### Canonical Object

Canvas stores a block-level `file_ref` object rather than a copy of file body content.

Minimum persisted shape:

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

Optional future-friendly field:

```json
{
  "source_context": "files|message|search"
}
```

`source_context` is not required for the first release, but the design allows it if implementation cost is low.

### Contract Authority Rules

To avoid per-surface drift, this phase uses two explicit contract layers:

- drag-time contract: `file-to-canvas`
- persisted canvas contract: `file_ref`

Authority rules:

- the drag-time contract is the only accepted client payload for drop insertion
- the persisted `file_ref` block is the only stored canvas representation
- Windsurf owns client-side normalization from any UI surface into the drag-time contract
- Gemini only needs to guarantee that each surface can obtain enough file data to build that normalized contract
- Codex owns contract versioning and any schema change approval for this phase

Authoritative drag-time fields:

- `kind`
- `file.id`
- `file.title`
- `file.mime_type`
- `file.size`

Conditionally authoritative drag-time field:

- `file.preview_url`
  - use when present
  - if a surface lacks it, Windsurf may derive it from stable file routing only if that derivation matches the canonical file preview route already used elsewhere

Informational drag-time field:

- `source`

Required normalization rule:

- if Files, message attachments, and search results expose richer or differently named file objects, those shapes must be normalized into the same drag-time contract before canvas drop handling begins
- canvas drop handling must reject non-normalized source-specific shapes

### Drag Contract

All three drag sources should emit one normalized `file-to-canvas` payload. The payload should be light enough for Web delivery but stable enough that canvas insertion does not need per-source branching.

Recommended normalized payload:

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

Canvas should treat `source` as informational only. Insertion logic should depend on `kind` and the nested file object, not on which UI emitted it.

### Canvas Insertion

Drop target behavior:

- only active when a canvas editor is open
- inserts at the current cursor position when possible
- if selection state is missing, fall back to a safe insertion point determined by the editor, without corrupting content
- never clears or replaces existing document content during failure recovery

Persistence behavior:

- the inserted card must survive refresh
- reopening the same canvas must re-render the same `file_ref` block
- persisted `file_ref` fields must be populated from the normalized drag-time contract, not from source-specific UI state

Insertion pass/fail rule:

- if the editor has a valid current selection, insert exactly at that selection
- if the current selection is unavailable but the editor document is writable, insert once at the editor's safe fallback insertion point
- if neither condition holds, abort insertion and show a lightweight failure hint; do not mutate the document

### File Preview / Detail

The canvas card should link back into existing file preview/detail UX rather than inventing a second file inspector.

This keeps file archive as the source of truth for:

- metadata
- comments
- shares
- knowledge state
- extraction state
- audit trails

## Backend / API Boundary

Phase 68 should avoid redesigning the file system. Backend work should stay minimal.

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend deliverables:

- confirm that all three surfaces can obtain a stable lightweight file payload
- if needed, add or normalize a lightweight file shape suitable for drag insertion
- ensure file preview/detail links remain stable for canvas consumption
- add tests for any new serializer or lightweight file payload helper

Preferred approach:

- reuse existing file APIs and existing file objects where possible
- only add a new backend field or helper when the current responses are too inconsistent across Files, message attachments, and search results

Contract note:

- Gemini is not the owner of the client drag schema
- Gemini is responsible for making the source file data stable enough that Windsurf can normalize all three surfaces into the Codex-frozen drag-time contract

### Avoid in Phase 68

- no new dedicated `POST /canvas/file-ref` API
- no server-side file-to-canvas orchestration service
- no AI file transformation service in this phase

## Web / UI Boundary

Windsurf owns all Web delivery.

### Windsurf Scope

- define one shared client drag payload for file-to-canvas operations
- make Files page rows/cards draggable
- make message file attachment cards draggable
- make search file results draggable when the result is a file
- teach canvas editor drop handling to insert a persisted `file_ref` block
- render compact file reference cards in canvas
- wire card click/open behavior into existing file preview/detail UX
- implement soft-failure UI and editor-safe failure handling

Contract note:

- Windsurf owns the normalization adapter that converts each surface's file object into the frozen `file-to-canvas` payload
- Windsurf must not let canvas insertion consume raw per-surface objects directly

### Web Constraints

- do not fork three separate insertion paths
- do not introduce a second file detail system inside canvas
- do not expand the card into a heavy inspector in this phase

## AI-Native Boundary

Phase 68 is not an AI feature phase, but it should leave obvious extension hooks.

Do not build in this release:

- summarize file into canvas
- insert excerpt into canvas
- create list item from file
- auto-generate a canvas from dropped file content

Leave room for later actions on the card such as:

- `Summarize`
- `Insert excerpt`
- `Create task`

Those actions should be future card actions, not hidden automatic side effects on drop.

## Error Handling

Required soft failures:

- no open canvas
- invalid drag payload
- unsupported drop target
- editor insertion failure
- missing file preview target

Failure rules:

- do not corrupt existing canvas content
- do not silently replace content
- do not leave a broken half-inserted block
- show a lightweight user-facing failure hint for every rejected drop
- if drop insertion is rejected, the document JSON before and after the drop attempt must remain byte-for-byte equivalent in persistence tests

## Acceptance Criteria

### Core

1. Dragging a file from the Files page into an open canvas inserts exactly one compact `file_ref` block at the active editor selection; if no active selection exists, insertion occurs exactly once at the editor's safe fallback insertion point.
2. Dragging from a message file attachment card or a file search result produces the same persisted `file_ref` JSON shape for the same source file, ignoring the optional `source_context` field.
3. Refreshing or reopening the canvas preserves the inserted `file_ref` block and re-renders the same compact card.
4. Clicking the card resolves to an existing file preview/detail route for that file ID, not a dead link and not a raw identifier-only placeholder.
5. For rejected drops, the editor document content remains unchanged and the user receives a lightweight failure hint.

### Architecture

1. The three drag entry points normalize into one frozen drag-time payload with the same authoritative fields.
2. Canvas insertion logic accepts only the normalized payload and remains source-agnostic.
3. File archive remains the source of truth; canvas stores references, not duplicated file bodies.

## Testing Strategy

### Backend

- serializer / response-shape tests for any lightweight file payload normalization
- regression tests proving existing file preview/detail routes still satisfy canvas card needs

### Web

- drag source tests for Files, message attachments, and search file results
- drop handling test proving insertion at current cursor location
- persistence/reload test for `file_ref` blocks
- click-through test from canvas card to file preview/detail
- soft-failure tests for invalid payload or missing open canvas

## Release Notes Guidance

This release should be described as:

`Files can now be pulled directly into canvas as compact reference cards, connecting the file archive and canvas into one collaboration flow.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: product definition, contract freeze, contract versioning authority, collaboration docs, integration control, follow-up phase planning
