# Phase 71 AI Execution History Design

## Summary

Phase 71 makes AI-driven execution traceable across Relay:

`analysis snapshot -> next step -> execution target -> draft -> confirmed action -> created object`

After Phase 70A/70B/70C established draft-first execution for `list`, `workflow`, and `channel_message`, Phase 71 adds an independent execution history event model that records what happened, where it came from, and what object was created or failed.

## Product Definition

`Every AI-driven execution leaves an independent, queryable, projectable history event that links back to the originating analysis snapshot, next step, execution target, draft, and final object.`

## Scope

### In Scope

- add an independent execution history fact model
- record target action lifecycle events for:
  - `list`
  - `workflow`
  - `channel_message`
- cover at least these event types:
  - `draft_generated`
  - `confirmed`
  - `created`
  - `published`
  - `failed`
- expose query APIs for Canvas AI Dock to show per-step execution state
- project execution history into Activity
- project recent AI-driven execution status into Home
- persist failure events instead of leaving failures as temporary UI-only state

### Out of Scope

- full compliance/audit logging
- immutable regulatory-grade logs
- automatic retry orchestration
- cross-workspace analytics
- new execution target types
- graph visualization
- replacing existing Activity feed storage with this model

## User Experience

### Canvas AI Dock

When a user returns to an analysis result in Canvas AI Dock, Relay should show whether each relevant next step has already been acted on.

Examples:

- draft was generated
- user confirmed it
- list/workflow/message was created or published
- a failure happened during draft generation, confirmation, creation, or publish

The Dock should expose direct links to created objects when available.

### Activity

Activity should surface AI execution events as workspace activity, not as hidden metadata.

Example activity meanings:

- AI suggestion generated a workflow draft
- AI-driven plan created a list
- AI-driven message was published to a channel
- AI execution failed during publish

### Home

Home should summarize recent AI-driven execution progress and failed items for management visibility.

The first release should stay compact:

- recent AI-driven executions
- failed AI-driven executions needing attention
- links back to the source Canvas or created object when available

## System Design

### Event Model

Recommended conceptual model:

```json
{
  "id": "exec-hist-123",
  "event_type": "draft_generated|confirmed|created|published|failed",
  "status": "pending|success|failed",
  "actor_user_id": "user-123",
  "analysis_snapshot_id": "analysis-123",
  "next_step_id": "step-123",
  "step_index": 2,
  "execution_target_type": "list|workflow|channel_message",
  "draft_id": "draft-123",
  "draft_type": "list|workflow|channel_message",
  "created_object_id": "list-123",
  "created_object_type": "list|workflow|message",
  "failure_stage": "draft_generation|confirmation|creation|publish",
  "error_message": "string",
  "created_at": "2026-04-26T00:00:00Z"
}
```

### Append-Only Rule

Execution history events are append-only facts.

That means:

- write a new event for each lifecycle transition
- do not mutate old events to represent the latest state
- current execution state is computed by projections from event history
- this model is operational history, not regulatory audit logging

### Required Fields

Every event should include:

- `id`
- `event_type`
- `status`
- `actor_user_id`
- `analysis_snapshot_id`
- `execution_target_type`
- `created_at`

Events should include at least one step locator when available:

- `next_step_id`
- `step_index`

Draft-related events should include:

- `draft_id`
- `draft_type`

Object-related events should include:

- `created_object_id`
- `created_object_type`

Failure events should include:

- `failure_stage`
- concise `error_message`

### Per-Event Field Matrix

`draft_generated` requires:

- `id`
- `event_type=draft_generated`
- `status=success`
- `actor_user_id`
- `analysis_snapshot_id`
- `execution_target_type`
- `draft_id`
- `draft_type`
- `created_at`
- one step locator:
  - `next_step_id`, preferred when available
  - otherwise `step_index`

`confirmed` requires:

- `id`
- `event_type=confirmed`
- `status=success`
- `actor_user_id`
- `analysis_snapshot_id`
- `execution_target_type`
- `draft_id`
- `draft_type`
- `created_at`
- one step locator:
  - `next_step_id`, preferred when available
  - otherwise `step_index`

`created` requires:

- `id`
- `event_type=created`
- `status=success`
- `actor_user_id`
- `analysis_snapshot_id`
- `execution_target_type`
- `draft_id`
- `draft_type`
- `created_object_id`
- `created_object_type`
- `created_at`
- one step locator:
  - `next_step_id`, preferred when available
  - otherwise `step_index`

`published` requires:

- `id`
- `event_type=published`
- `status=success`
- `actor_user_id`
- `analysis_snapshot_id`
- `execution_target_type=channel_message`
- `draft_id`
- `draft_type=channel_message`
- `created_object_id`
- `created_object_type=message`
- `created_at`
- one step locator:
  - `next_step_id`, preferred when available
  - otherwise `step_index`

`failed` requires:

- `id`
- `event_type=failed`
- `status=failed`
- `actor_user_id`
- `analysis_snapshot_id`
- `execution_target_type`
- `failure_stage`
- `error_message`
- `created_at`
- one step locator when the failure is tied to a specific recommendation:
  - `next_step_id`, preferred when available
  - otherwise `step_index`

`failed` may include:

- `draft_id`
- `draft_type`
- `created_object_id`
- `created_object_type`

### Event Types

#### `draft_generated`

Written when Relay successfully creates a draft for an AI-driven execution.

Applies to:

- list draft
- workflow draft
- channel message draft

#### `confirmed`

Written when the user explicitly confirms the draft.

This records user intent before final object creation/publish completes.

#### `created`

Written when an object is created successfully.

Applies to:

- list
- workflow

#### `published`

Written when a message is published successfully.

Applies to:

- channel_message

#### `failed`

Written when any lifecycle stage fails.

The event should preserve enough context for UI recovery and support investigation without becoming a full audit log.

## Backend / API Boundary

Phase 71 backend work should make execution history the source of truth for AI-driven execution status.

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend responsibilities:

- add execution history event model and persistence
- write events from all three existing execution chains:
  - list draft/create
  - workflow draft/create
  - channel message draft/publish
- add query API for analysis-scoped execution history
- project execution history into Activity feed
- project recent/failed AI execution summary into Home
- add tests for:
  - event writes for success paths
  - event writes for failure paths
  - query by `analysis_snapshot_id`
  - Activity projection
  - Home projection

### Backend Constraints

- do not replace existing Activity feed storage wholesale
- do not introduce regulatory audit guarantees
- do not add automatic retry behavior
- do not add new execution target types

## Web / UI Boundary

Windsurf owns all Web delivery.

### Windsurf Scope

- query execution history for the active analysis snapshot in Canvas AI Dock
- render per-step execution status
- link created objects from the Dock when available
- render failure states from persisted history events
- consume Activity projection for AI execution events
- consume Home projection for recent/failed AI-driven executions

### Web Constraints

- do not rely on temporary Dock state as the source of truth after execution
- do not infer execution status from button state or local component state
- do not build a full audit viewer
- do not add retry UI unless a backend retry contract exists

## Projection Rules

### Canvas AI Dock Projection

Canvas should group events by:

- `analysis_snapshot_id`
- `next_step_id` or `step_index`
- `execution_target_type`

Deterministic state resolution:

- group by `analysis_snapshot_id + step locator + execution_target_type`
- sort events by `created_at` ascending, then `id` ascending as a tie-breaker
- current state is the last event in the group by that ordering
- visible states:
  - `draft_generated`
  - `confirmed`
  - `created`
  - `published`
  - `failed`
- if the current state is `failed`, show `failure_stage` and concise `error_message`
- if the current state has `created_object_id`, show an object link
- the first release should show at most the latest state plus an optional compact history disclosure per step

### Activity Projection

Activity should display execution events as concise workspace events.

Recommended event family:

- `ai_execution`

Recommended fields:

- actor/user
- event type
- target type
- source canvas/analysis link
- created object link when available

Deterministic Activity rules:

- event family: `ai_execution`
- include all `created`, `published`, and `failed` events
- do not include `draft_generated` or `confirmed` in the first-release Activity feed unless no later event exists for that same group
- order by `created_at` descending, then `id` descending
- dedupe by `analysis_snapshot_id + step locator + execution_target_type + event_type`
- default page size: 20 items

### Home Projection

Home should show a compact AI execution summary.

Recommended blocks:

- recent AI-driven executions
- failed AI executions needing attention

Deterministic Home rules:

- time window: last 7 days by `created_at`
- recent AI-driven executions:
  - latest 5 `created` or `published` events
  - order by `created_at` descending
- failed AI executions needing attention:
  - latest 5 `failed` events
  - order by `created_at` descending
- each row should include:
  - execution target type
  - event type/current state
  - source canvas/analysis link when available
  - created object link when available

## Error Handling

Required behaviors:

- if draft generation fails, write `failed` with `failure_stage=draft_generation`
- if confirmation fails, write `failed` with `failure_stage=confirmation`
- if list/workflow creation fails, write `failed` with `failure_stage=creation`
- if channel publish fails, write `failed` with `failure_stage=publish`
- UI must show persisted failure state after refresh
- missing optional object links should not break history rendering

## Acceptance Criteria

### Core

1. List, workflow, and channel_message execution chains all write execution history events.
2. Events cover target action lifecycle at least for:
   - `draft_generated`
   - `confirmed`
   - `created`
   - `published`
   - `failed`
3. Events link back to:
   - `analysis_snapshot_id`
   - `next_step_id` or `step_index`
   - `execution_target_type`
   - `draft_id`
   - `created_object_id` / `created_object_type` when available
4. Canvas AI Dock shows persisted execution state for each acted-on step.
5. Activity shows AI execution events.
6. Home summarizes recent and failed AI-driven execution.
7. Failure states survive refresh and are not UI-only.

### Architecture

1. Execution history is an independent fact model.
2. Activity and Home consume projections from that fact model.
3. Phase 71 adds traceability without introducing new execution target families.

## Testing Strategy

### Backend

- model persistence tests
- event write tests for list success/failure
- event write tests for workflow success/failure
- event write tests for channel_message success/failure
- analysis-scoped query tests
- Activity projection tests
- Home projection tests

### Web

- Canvas Dock execution-state rendering tests
- created-object link tests
- persisted failure-state rendering tests
- Activity AI execution event rendering tests
- Home summary rendering tests

## Release Notes Guidance

This release should be described as:

`AI-driven execution now leaves a traceable history from analysis suggestion to draft, confirmation, and created object, with status visible in Canvas, Activity, and Home.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: event taxonomy, projection boundary, collaboration docs, integration control, follow-up phase planning
