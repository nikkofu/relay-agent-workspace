# Phase 70C Draft-First Workflow And Channel Publish Targets Design

## Summary

Phase 70C deepens Relay's execution taxonomy by making two more target families truly executable:

`analysis snapshot -> typed target -> draft preview -> user confirm -> create/publish`

After Phase 70A established `list` as a real execution path and Phase 70B established typed execution targets, Phase 70C extends the same draft-first execution model to:

- `workflow`
- `channel_message`

## Product Definition

`Workflow and channel_message targets become draft-first execution paths. AI produces typed drafts, the user reviews them in Canvas AI Dock, and only then does Relay create or publish.`

## Scope

### In Scope

- define minimal typed payload depth for:
  - `workflow`
  - `channel_message`
- keep the existing execution-target inheritance model:
  - step override
  - analysis default
- generate workflow drafts inside Canvas AI Dock
- generate channel-message drafts inside Canvas AI Dock
- require explicit user confirmation before creation/publish
- keep the protocol globally available across AI reply surfaces, while landing real execution UX in Canvas first

### Out of Scope

- full workflow builder UX
- advanced workflow parameter binding
- multi-channel message distribution
- auto-publish or auto-execute
- full execution UX parity across DM, `/ask`, and Canvas
- prose re-parsing from arbitrary user-edited text

## User Experience

### Draft-First Execution Grammar

Phase 70C extends a consistent execution model:

- `list`
  - analysis snapshot -> list draft -> confirm create
- `workflow`
  - analysis snapshot -> workflow draft -> confirm create
- `channel_message`
  - analysis snapshot -> message draft -> confirm publish

The key product principle is consistency: AI does not jump straight to execution. It drafts first, then the user confirms.

### Canvas-First Interaction

Real execution entry points for `workflow` and `channel_message` land first in Canvas AI Dock.

That means:

- Canvas AI receives the first full UX
- `/ask` and AI DM may carry the same typed target contract
- other surfaces are not forced into heavy execution UX in this phase

### Workflow Draft UX

The first-release workflow draft should be lightweight and reviewable.

Minimum preview content:

- title
- goal/summary
- ordered steps

Optional but still shallow:

- lightweight recommended trigger type

### Channel Message Draft UX

The first-release message draft should also be lightweight.

Minimum preview content:

- target channel
- message body

Optional but shallow:

- short lead-in/title line

## System Design

### Shared Execution Context

Phase 70C continues to depend on typed execution targets from Phase 70B and frozen analysis snapshots from earlier phases.

The conceptual source remains:

- a valid structured analysis snapshot
- an execution target resolved through:
  - step target if present
  - otherwise analysis default target

### Workflow Target Shape

First-release workflow targets should stay typed but shallow.

Recommended conceptual shape:

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

Authority rules:

- `type` is required
- `workflow_draft.title` is required for preview/create
- `workflow_draft.goal` is required in first release
- `workflow_draft.steps[]` is required and ordered
- first-release step schema is title-only

### Channel Message Target Shape

Recommended conceptual shape:

```json
{
  "type": "channel_message",
  "message_draft": {
    "channel_id": "channel-123",
    "body": "string"
  }
}
```

Authority rules:

- `type` is required
- `message_draft.channel_id` defaults to the current canvas channel in first release
- `message_draft.body` is required

### Draft Identity Rule

Like the list flow, both new execution families should be draft-first and immutable through confirmation.

That means:

- workflow draft generation produces a stable workflow draft identity before create
- channel-message draft generation produces a stable message draft identity before publish
- confirm actions consume those immutable draft identities, not recomputed live Dock text

### Execution Resolution Rule

Phase 70C does not change Phase 70B's resolution logic:

1. if `next_steps[i].execution_target` exists and is valid, use it
2. otherwise use `analysis.default_execution_target` if valid
3. otherwise there is no execution route

## Backend / API Boundary

Phase 70C backend work should focus on typed payload depth plus explicit draft-generation and confirm-execute contracts.

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend responsibilities:

- define typed workflow target payload support
- define typed channel-message target payload support
- generate workflow drafts from frozen analysis snapshot + resolved target
- generate message drafts from frozen analysis snapshot + resolved target
- return immutable draft identities for both families
- confirm-create workflow from immutable workflow draft ID
- confirm-publish message from immutable message draft ID
- add tests for:
  - valid workflow draft generation
  - valid message draft generation
  - confirm-create workflow
  - confirm-publish message
  - failure paths that keep draft-first semantics intact

### Backend Constraints

- do not auto-create workflow on draft generation
- do not auto-publish channel message on draft generation
- do not re-run AI when creating/publishing from a valid generated draft
- do not widen first-release payloads into full builder schemas

## Web / UI Boundary

Windsurf owns all Web delivery.

### Windsurf Scope

- extend the shared execution-target consumption layer for:
  - `workflow`
  - `channel_message`
- surface draft-generation actions inside Canvas AI Dock
- render workflow draft preview
- render message draft preview
- confirm-create workflow using immutable workflow draft ID
- confirm-publish message using immutable message draft ID
- show success/failure states while preserving the source analysis context
- keep `/ask` and AI DM contract-compatible even if their UX stays lighter

### Web Constraints

- do not build a full workflow builder
- do not turn channel publish into a multi-target composer
- do not auto-execute or auto-publish
- do not infer draft payloads from arbitrary prose

## AI-Native Behavior

This phase is AI-native because it upgrades the remaining execution target families into controlled execution flows rather than leaving them as decorative labels.

The intended feeling is:

- AI recommendations already know whether they are:
  - list-like
  - workflow-like
  - channel-publish-like
- Relay can then materialize the right kind of draft
- the user approves it before execution

This keeps the system agentic without becoming reckless.

## Error Handling

Required behaviors:

- if a workflow or channel_message target lacks the minimum required draft payload, no execution path should be shown
- if draft generation fails, keep the source analysis visible and show a Dock-local failure state
- if confirm-create/publish fails, keep the draft visible and clearly unexecuted
- do not show false success
- do not lose analysis context because a draft or confirm step failed

## Acceptance Criteria

### Core

1. Canvas AI Dock can generate a workflow draft from a valid `workflow` target.
2. Canvas AI Dock can generate a message draft from a valid `channel_message` target.
3. Both draft families render preview-first inside the Dock.
4. Workflow creation and channel publish happen only after explicit confirmation.
5. Confirm actions use immutable draft identities rather than recomputed live UI text.
6. Failure states keep draft-first semantics and preserve analysis context.

### Architecture

1. Phase 70C keeps the global execution-target contract from Phase 70B.
2. Canvas remains the first surface with full execution UX.
3. `workflow` and `channel_message` now follow the same draft-first execution grammar already established for `list`.

## Testing Strategy

### Backend

- workflow target payload validation tests
- channel_message payload validation tests
- workflow draft generation tests
- channel message draft generation tests
- immutable draft-id confirm tests
- failure-path tests

### Web

- Canvas draft-entry visibility tests
- workflow draft preview tests
- channel message draft preview tests
- confirm-create/publish tests
- context-preservation and failure-state tests
- contract-compatibility tests for `/ask` and AI DM

## Release Notes Guidance

This release should be described as:

`Canvas AI can now turn workflow-shaped and channel-publish-shaped recommendations into reviewable drafts before you create or publish them, extending Relay's draft-first AI execution model beyond lists.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: contract freeze, draft-first execution consistency, collaboration docs, integration control, follow-up phase planning
