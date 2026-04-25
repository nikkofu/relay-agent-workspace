# Phase 70A Create List From Analysis Snapshot Design

## Summary

Phase 70A turns the output of Phase 69 into the first execution object in Relay's AI-native workflow:

`multi-file canvas analysis snapshot -> Dock action -> List draft preview -> user confirms -> one channel-scoped list with multiple items is created`

The first release does not infer tasks from arbitrary text. It only converts a frozen structured analysis snapshot that already exists in Canvas AI Dock.

## Product Definition

`Canvas AI Dock does not just analyze a file group. It can convert a frozen analysis snapshot into one List draft for the current channel, and the user confirms before creation.`

## Scope

### In Scope

- expose `Create list from plan` only when the Dock contains a valid structured analysis snapshot
- generate one List draft from that snapshot
- show draft preview in the Dock
- default the draft to the current canvas's channel
- create one list with multiple items only after explicit user confirmation
- return the user to either:
  - the created list
  - the prior analysis view

### Out of Scope

- generating lists from arbitrary canvas text
- parsing already-inserted prose back into tasks
- automatic list creation without confirmation
- multiple-list splitting
- automatic assignee, due date, or priority inference
- direct workflow/tool execution after creation

## User Experience

### Primary Flow

1. A user has a valid structured analysis result in Canvas AI Dock.
2. The Dock exposes `Create list from plan`.
3. The user clicks it.
4. Relay converts the frozen analysis snapshot into a draft preview:
   - list title
   - list items
   - target channel
5. The user clicks `Confirm create`.
6. Relay creates:
   - one list
   - multiple items
7. The Dock shows success actions:
   - `Open list`
   - `Back to analysis`

### Why Dock-First

This keeps the whole chain inside one continuous AI-native workflow:

- analyze source materials
- review AI recommendations
- convert those recommendations into an execution object
- confirm creation

That is more agentic than forcing the user to jump into a separate modal or manually reconstruct tasks from prose.

## Input Boundary

The only valid source for Phase 70A is a frozen structured analysis snapshot from Phase 69.

This means:

- valid source:
  - Dock-held structured analysis result tied to a specific analysis run
- invalid sources:
  - arbitrary canvas text
  - manually edited inserted plan prose
  - raw AI chat text without structured `summary / observations / next_steps`

### Snapshot Rule

List draft generation must use the exact frozen analysis snapshot that was approved for conversion.

That means:

- every valid source snapshot must carry an immutable `analysis_snapshot_id`
- no re-running AI during draft generation
- no reparsing user-edited prose
- no silently switching to newer Dock state if the analysis changes after the draft is opened

## System Design

### Canonical Input

The conceptual source object for this phase is:

```json
{
  "artifact_id": "artifact-123",
  "channel_id": "channel-123",
  "analysis_snapshot_id": "analysis-123",
  "analysis_snapshot": {
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

Authority rules:

- `artifact_id` identifies the canvas context
- `channel_id` defaults from the current canvas context and is the default list destination
- `analysis_snapshot_id` is the immutable identity of the exact structured analysis run being converted
- `analysis_snapshot.next_steps[]` is the primary source for draft item generation
- `summary` and `observations` may influence list title/description generation if needed, but they do not replace `next_steps[]` as the execution source

Snapshot authority rule:

- Windsurf must invoke draft generation from one specific `analysis_snapshot_id`
- Gemini must treat that snapshot identity as the canonical conversion source for the draft
- current Dock text or mutable canvas state must not override the referenced snapshot

### Canonical Draft Output

The draft should be structured, not an opaque blob.

Recommended conceptual output:

```json
{
  "draft": {
    "draft_id": "draft-123",
    "channel_id": "channel-123",
    "title": "Launch readiness follow-up",
    "items": [
      {
        "title": "Confirm missing launch dependencies"
      }
    ]
  }
}
```

Authority rules:

- one `draft`
- one immutable `draft_id`
- one `title`
- one `channel_id`
- ordered `items[]`
- each item must have a `title`
- first release item schema is title-only

First-release transformation rule:

- generate exactly one draft item per `analysis_snapshot.next_steps[]` entry
- preserve the source ordering from `next_steps[]`
- do not enrich first-release items with assignee, due date, priority, or freeform descriptions
- draft title may be derived from the snapshot context, but item generation must come only from `next_steps[]`

### Canonical Create Action

Creation should be explicit and separate from draft generation.

Recommended conceptual create flow:

1. request draft from the frozen analysis snapshot
2. render draft preview in Dock
3. user confirms creation against the returned `draft_id`
4. backend creates:
   - one list
   - multiple list items

The system must not treat preview generation as actual list creation.

Create authority rule:

- confirm-create must consume the previously returned `draft_id`
- the backend must not silently recompute the draft from mutable Dock state at confirm time
- if the draft is stale, missing, or invalid, create must fail clearly rather than improvising a new conversion

## Backend / API Boundary

Phase 70A backend work should focus on deterministic draft generation from a frozen structured snapshot plus explicit creation.

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend responsibilities:

- define the final draft-generation contract
- define the final create-from-draft contract
- accept only valid structured analysis snapshots
- require immutable `analysis_snapshot_id` for draft generation
- reject unstructured or prose-only input
- generate:
  - default list title
  - exactly one ordered item per `next_steps[]` entry
- create one list in the requested/default channel only after explicit confirmation
- require `draft_id` on confirm-create
- add tests for:
  - valid structured snapshot -> draft
  - invalid/unstructured input rejection
  - missing/invalid `analysis_snapshot_id` rejection
  - confirm-create requires valid `draft_id`
  - confirm-create -> one list + many items
  - failure path does not pretend creation succeeded

### Backend Constraints

- do not auto-create on draft generation
- do not re-run AI inference when the snapshot already exists
- do not parse arbitrary canvas prose into tasks
- do not create multiple lists in the first release

## Web / UI Boundary

Windsurf owns all Web delivery for this phase.

### Windsurf Scope

- expose `Create list from plan` only when a valid structured analysis snapshot exists
- keep the action tied to the frozen `analysis_snapshot_id` currently visible in the Dock
- render Dock-local draft preview
- show:
  - title
  - items
  - target channel
- default target channel from the current canvas
- call explicit confirm-create action
- render success state with:
  - `Open list`
  - `Back to analysis`
- preserve failure states without losing the source analysis snapshot

### Web Constraints

- do not offer conversion from plain inserted canvas text
- do not auto-confirm creation
- do not split one analysis into multiple lists in the first release
- do not expose heavyweight configuration UI for assignees/dates/priorities in this phase

## AI-Native Behavior

This phase should feel agentic because Relay is turning AI proposals into execution objects in a controlled, reviewable way.

The intended product feeling is:

- AI studies a file group
- AI produces a structured plan
- Relay converts that plan into an execution draft
- the user approves creation in one continuous flow

This leaves clean follow-up hooks for later phases:

- `Create workflow from plan`
- `Send plan to channel`
- `Map action_hint to execution type`
- richer list item metadata generation

## Error Handling

Required behaviors:

- if no valid structured analysis snapshot exists, do not show `Create list from plan`
- if draft generation fails, keep the analysis snapshot visible and show a Dock-local error
- if confirm-create fails, keep the draft preview tied to its `draft_id` visible and clearly uncreated
- do not show success state until actual list creation succeeds
- do not lose the analysis snapshot just because draft/create failed

## Acceptance Criteria

### Core

1. `Create list from plan` is available only when the Dock contains a valid structured analysis snapshot.
2. Draft generation uses the frozen structured analysis snapshot identified by `analysis_snapshot_id` rather than re-running AI or reparsing prose.
3. Draft preview appears inside the Dock and includes:
   - list title
   - list items
   - target channel
4. The target channel defaults to the current canvas's channel.
5. Confirming creation with the returned `draft_id` creates exactly one list with multiple items in that channel.
6. Creation failure keeps the snapshot/draft state available and does not show false success.

### Architecture

1. Phase 70A consumes structured analysis output from Phase 69 rather than inventing a new AI source.
2. Draft generation and final creation are distinct steps.
3. The first release produces one list container, not multiple lists.

## Testing Strategy

### Backend

- request validation tests for structured snapshot input
- request validation tests for `analysis_snapshot_id`
- rejection tests for prose-only or malformed input
- draft generation shape tests
- draft generation tests proving one ordered item per `next_steps[]`
- creation tests proving one list + many items
- creation tests proving valid `draft_id` is required
- failure-path tests for create errors

### Web

- visibility test for `Create list from plan`
- Dock draft-preview rendering test
- confirm-create success flow test
- failure-state retention test
- navigation/action test for `Open list` and `Back to analysis`

## Release Notes Guidance

This release should be described as:

`Canvas AI can now turn a structured file-group analysis into a reviewable List draft and create one channel-scoped execution list after you confirm it.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: product definition, contract freeze, collaboration docs, integration control, follow-up phase planning
