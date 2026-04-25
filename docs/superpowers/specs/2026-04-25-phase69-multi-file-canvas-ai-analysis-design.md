# Phase 69 Multi-File Canvas AI Analysis Design

## Summary

Phase 69 extends Relay's file archive + canvas convergence into a more AI-native workflow:

`multiple persisted file_ref cards in one canvas -> Canvas AI Dock analyzes the file group -> AI returns a structured analysis preview -> user selectively inserts approved sections into canvas`

The first release focuses on AI understanding and proposal, not automatic execution. The Dock becomes an analysis and recommendation surface for a file group already assembled inside canvas.

## Product Definition

`Canvas is the working surface for a file group. Canvas AI Dock is the analysis and proposal layer. The user decides which AI conclusions become formal canvas content.`

## Scope

### In Scope

- detect when a canvas contains multiple persisted `file_ref` blocks
- let the user trigger AI analysis for the current file group from Canvas AI Dock
- generate a structured analysis result with three sections:
  - merged summary
  - key observations
  - next step suggestions
- keep results in the Dock first rather than auto-writing to canvas
- let the user insert any individual section into the canvas
- gracefully degrade when some files lack extraction data, previews, or complete content

### Out of Scope

- auto-writing analysis results into canvas without user confirmation
- direct creation of List items, Workflow instances, or Tool runs
- continuous background re-analysis whenever files change
- advanced file grouping UI, selection ranges, or multi-group management
- cross-canvas shared AI analysis sessions
- treating raw drag payloads as AI input

## User Experience

### Primary Flow

1. A user brings multiple files into one canvas as persisted `file_ref` cards.
2. Canvas AI Dock detects that the current canvas contains a file group.
3. The Dock presents an action such as `Analyze file group`.
4. AI returns a structured result preview in the Dock.
5. The user chooses whether to insert:
   - summary
   - observations
   - plan
6. Only user-approved sections are inserted into the canvas document.

### Why Dock-First

The first release intentionally keeps AI output in the Dock before insertion because:

- it preserves user control in multi-file contexts
- it feels more agentic than silently drafting into the document
- it supports future actions such as `Create list`, `Send to channel`, or `Run tool` without making the first release too heavy

### Output Shape

The first release uses one fixed output structure:

#### Summary

A short merged summary of the file group.

#### Key Observations

A compact set of notable findings, differences, risks, or themes.

#### Next Step Suggestions

Three to five natural-language suggestions with short rationale. These are human-readable recommendations first, but they should be structured enough internally that future phases can map them into List, Workflow, or Tool actions.

## AI Input Boundary

The AI input for this phase is not the original drag payload. It is the current canvas's persisted set of `file_ref` blocks.

This matters because:

- Phase 69 builds on Phase 68 persisted file-canvas convergence
- the canvas already represents the user's chosen working set
- the analysis should operate on durable file references, not ephemeral UI state

### First-Release File Group Definition

For the first release, `file group` means:

- all persisted `file_ref` blocks currently present in the active canvas

This avoids introducing a separate selection or grouping system in the same phase.

### Snapshot Rule

The file group must be snapshotted at trigger time.

That means:

- when the user clicks `Analyze file group`, Windsurf captures the current persisted `file_ref` set from the active canvas
- the backend request uses that captured snapshot, not a re-resolved later canvas state
- if the user edits the canvas file group after trigger, that does not mutate the in-flight analysis request
- a new analysis must be triggered to reflect a changed file group

## System Design

### Canonical Analysis Request

The Dock should send a request that is derived from the active canvas and its persisted file references.

Recommended conceptual request shape:

```json
{
  "artifact_id": "artifact-123",
  "file_refs": [
    { "file_id": "file-1" },
    { "file_id": "file-2" },
    { "file_id": "file-3" }
  ],
  "mode": "multi_file_analysis"
}
```

Authority rules:

- `artifact_id` identifies the canvas context
- `file_refs[].file_id` is the authoritative file-group input
- the submitted `file_refs[]` array is a trigger-time snapshot of the active canvas file group
- no raw drag payload should be accepted at this stage
- Windsurf may derive the request from current canvas document state
- Gemini defines the final backend request shape, but it must remain faithful to this conceptual contract

### Canonical Analysis Result

The AI result should be structured, not a single opaque blob.

Recommended conceptual result shape:

```json
{
  "analysis": {
    "summary": "string",
    "observations": [
      "string"
    ],
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

Authority rules:

- `summary` is a single merged summary string
- `observations` is an ordered array of concise findings
- `next_steps` is an ordered array with:
  - `text`
  - short `rationale`
  - lightweight `action_hint`
- the Web layer may render these sections richly, but it should not require inference from unstructured prose

`action_hint` is not a direct executable action in Phase 69. It is a lightweight machine-readable intent hint so later phases can map suggestions toward List, Workflow, or Tool flows without redefining the result shape.

### Canvas Insertion Model

Inserted content should be explicit and user-initiated.

Allowed insertion actions:

- insert summary
- insert observations
- insert next-step plan

The insertion target should follow one clear editor rule in implementation, such as:

- current cursor when available
- otherwise one editor-safe fallback insertion point

The AI Dock should not auto-insert after generation.

Preview lifecycle rules:

- Dock analysis result remains temporary preview state until the user explicitly inserts a section
- insertion copies from the frozen analysis result snapshot that was returned for that run; it must not re-serialize from mutable live UI state
- inserted canvas content becomes normal persisted document content and is no longer tied to continued Dock mutability
- the Dock may continue to show the analysis result after insertion, but it must remain visually distinct from persisted canvas content

## Backend / API Boundary

Phase 69 backend work should stay focused on analysis input aggregation and structured results.

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend responsibilities:

- define the final analysis request/response schema based on the conceptual contract above
- gather file-group inputs from persisted `file_ref` references
- resolve file extraction, preview, metadata, or content sources as available
- handle mixed-quality inputs where some files are partially indexed or unavailable
- return a structured analysis payload with:
  - summary
  - observations
  - next_steps
- add tests for:
  - valid multi-file request handling
  - partial file degradation behavior
  - structured result shape stability

### Backend Constraints

- do not create automatic task/list/workflow side effects in this phase
- do not use raw drag payloads as the primary analysis input
- do not require that every file have complete extracted content in order for analysis to proceed

## Web / UI Boundary

Windsurf owns all Web delivery for this phase.

### Windsurf Scope

- detect multiple persisted `file_ref` blocks in the active canvas
- expose a Dock action such as `Analyze file group`
- send the analysis request using the persisted canvas file group
- render the structured result preview in the Dock
- provide section-level insertion actions:
  - `Insert summary`
  - `Insert observations`
  - `Insert plan`
- clearly separate temporary Dock analysis state from already-inserted canvas content

### Web Constraints

- do not auto-write AI output into canvas
- do not add a complex file-group manager or selection UI in this phase
- do not force users into channel/List/Workflow actions before they have reviewed the AI result

## AI-Native Behavior

This phase should feel agentic because AI is operating on a user-assembled working set and proposing next actions, not because it is taking irreversible actions on its own.

The desired feeling is:

- the user assembles a set of source materials
- AI understands the combined context
- AI produces a preview of analysis and practical next steps
- the user chooses which parts become formal content

This leaves clean hooks for later phases:

- `Create list from plan`
- `Send plan to channel`
- `Run workflow from suggestion`
- `Summarize only selected file_refs`

## Error Handling And Degradation

Required degradation behaviors:

- if one or more files lack extracted text, continue with the remaining files and any fallback metadata available
- if file previews are missing, analysis should still proceed when text or metadata is available
- if the file group is too large, backend may truncate or summarize inputs internally, but must still return a valid structured result or a user-readable failure
- if analysis fails completely, the Dock should surface an error state without mutating canvas content

Failure rules:

- do not auto-insert partial output on failure
- do not blur temporary Dock state with persisted canvas content
- do not force the user to reassemble the file group after a recoverable analysis error

## Acceptance Criteria

### Core

1. When the active canvas contains multiple persisted `file_ref` blocks, Canvas AI Dock exposes a file-group analysis action.
2. Triggering analysis uses the active canvas's persisted file references as input rather than raw drag payloads or generic chat context.
3. The Dock returns a structured result with all three required sections:
   - summary
   - observations
   - next step suggestions
4. The user can insert any one section independently into the canvas.
5. Inserted content persists normally as part of the canvas document.
6. Missing extraction or incomplete file content on some files degrades the analysis rather than forcing total failure whenever reasonable.

### Architecture

1. Phase 69 builds on Phase 68 persisted `file_ref` blocks as the source of truth for the file group.
2. The Dock analysis result is structured enough that future phases can map suggestions to List/Workflow/Tool actions without re-defining the output model.
3. Temporary AI analysis state in the Dock remains distinct from already-persisted canvas content.

## Testing Strategy

### Backend

- request validation tests for multi-file analysis inputs
- file-group aggregation tests based on persisted `file_ref` references
- partial degradation tests when some files have weak extraction state
- structured response-shape tests

### Web

- detection test for canvases containing multiple `file_ref` blocks
- Dock action test for file-group analysis trigger
- rendering tests for:
  - summary
  - observations
  - next-step suggestions
- insertion tests for each section action
- state-separation test proving Dock preview state does not masquerade as persisted document content

## Release Notes Guidance

This release should be described as:

`Canvas AI can now analyze a group of files already assembled in your canvas, produce a merged summary plus next-step suggestions, and let you selectively insert the useful parts into your working document.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: product definition, contract freeze, collaboration docs, integration control, follow-up phase planning
