# Channel Execution Hub Design

## Goal

Deliver a channel-centered execution layer for Relay so teams can turn channel context into tracked work, run tools against that context, and manage cross-channel execution from Home without breaking the existing Slack-style navigation model.

This phase intentionally focuses on one product slice:

1. make each channel a first-class execution hub for lists and tools
2. make Home a management cockpit for cross-channel execution visibility
3. add thin AI-native acceleration on top of the existing message/list/tool system

## Why This Phase Exists

Relay already has strong messaging, DM, file, artifact, activity, knowledge, workflow, and people surfaces. It also already has backend support for structured work objects such as lists and tool runs.

What is still missing is a coherent execution loop:

- channel discussion does not yet cleanly turn into tracked list work
- tool execution exists, but does not feel native to the channel context
- Home shows workspace activity, but not enough management-grade execution visibility
- AI features exist across the product, but not yet in the highest-value execution seam

This phase closes that gap with a narrow product definition:

`channel = execution context`
`Home = management cockpit`
`AI = suggestion and summarization layer`

## Scope

### Included

- add a channel-level execution surface for `Lists` and `Tools`
- allow users to create and manage channel-scoped lists from the channel experience
- allow users to execute tools from channel context
- allow users to turn a message into a list item with AI-assisted field suggestions
- allow tool results to be written back as a message or a list item
- extend Home with cross-channel execution summary blocks
- add channel execution summary data for management visibility
- define explicit implementation ownership:
  - all Web delivery goes to Windsurf
  - backend, API contracts, persistence, and aggregate semantics go to Codex

### Excluded

- complex workflow orchestration or multi-step automation chains
- nested project/task hierarchies
- cross-channel shared list editing models
- AI auto-closing tasks or auto-changing structured state
- a standalone tools marketplace
- a new execution engine separate from existing lists and tool runs

## Product Definition

### 1. Channel As Execution Hub

Each channel becomes the primary place where work is discussed and executed.

The channel experience gains an `Execution` surface with two fixed sections:

- `Lists`
- `Tools`

This surface should stay adjacent to message context rather than replacing the message view. The intended first implementation is a side panel or sibling panel in the existing channel page, not a separate product area.

### 2. Home As Management Cockpit

Home remains a workspace landing page, but gains three execution-oriented management blocks:

- `Open List Work`
- `Tool Runs Needing Attention`
- `Channel Execution Pulse`

Home is not the primary editing surface. It summarizes, filters, prioritizes, and deep-links back to the relevant channel and object.

### 3. AI As Thin Acceleration Layer

AI does not become the source of truth for structured work.

Instead, AI provides two narrow value-adds:

- `message -> list item` field suggestions
- `channel execution summary` generation

Structured facts still come from persisted messages, list items, and tool runs.

## Core User Flows

### A. Create A Channel List

From a channel execution panel, a user creates a list such as:

- release readiness
- customer follow-up
- weekly leadership actions

The list is channel-scoped by default and visibly tied to that channel.

Initial list-item fields stay intentionally small:

- title
- status
- assignee
- due date

### B. Turn A Message Into A List Item

From message actions, the user chooses `Add to List`.

The system:

1. lets the user choose a list in the current channel
2. proposes a task title and optional due-date/assignee suggestions using AI
3. stores a durable reference back to the source message

The user remains the final confirmer. AI suggestions must be editable before save.

### C. Run A Tool In Channel Context

From the channel execution tools section, the user chooses a tool to run.

Execution defaults to the current channel context and may optionally include:

- current thread
- selected message

The resulting tool run is visible in channel context and in Home summaries.

### D. Write Tool Results Back Into The Workspace

A successful tool run may write back to exactly one of these first-phase targets:

- message
- list item

This keeps tool outputs useful without introducing broad object sprawl.

### E. Manage Cross-Channel Execution From Home

Home users, especially managers and operators, can answer:

- which channels have the most open work
- which channels have overdue list items
- which tools are currently running or failed
- which channels need attention today

Each summary card deep-links back to the originating channel and object.

## UX Structure

## A. Channel Surface

Reuse the existing channel page and add an execution panel parallel to current secondary panels such as knowledge.

Recommended first structure:

- channel messages remain primary
- execution panel is toggled from channel context
- execution panel contains two tabs:
  - `Lists`
  - `Tools`

This preserves context density and prevents execution work from feeling detached from conversation.

## B. Home Surface

Extend the existing Home dashboard rather than replacing it.

Add three blocks:

### `Open List Work`

Shows cross-channel unfinished items with enough metadata to triage:

- item title
- channel
- assignee
- due date
- status

### `Tool Runs Needing Attention`

Shows runs that are:

- running
- failed
- recently completed but worth review

### `Channel Execution Pulse`

Shows per-channel execution signals such as:

- open item count
- overdue count
- recent tool activity
- recent execution summary

## Data Model And Contract Direction

This phase should avoid inventing a new structured-work model.

### Lists

Build on existing list/list-item APIs and stores.

Required additions:

- durable message reference on list items created from messages
- channel-friendly aggregation fields for Home and channel execution views

Recommended message-reference minimum shape:

- `message_id`
- `channel_id`
- source message snippet

### Tools

Build on existing tool and tool-run APIs.

Required additions:

- channel-native execution entry consumption
- explicit writeback target support

Recommended writeback target enum for this phase:

- `message`
- `list_item`

### Home Aggregation

Prefer extending the existing Home aggregation payload over creating multiple new front-end fetch dependencies.

Recommended additions:

- `open_list_work[]`
- `tool_runs_needing_attention[]`
- `channel_execution_pulse[]`

### Channel Execution Summary

Provide a lightweight summary source that combines:

- recent list-item creation/completion
- overdue state
- recent tool runs
- notable blockers or failures

This may be a dedicated endpoint or a Home/channel aggregate field, but it should remain lightweight and query-oriented.

## AI-Native Behavior

## A. Message To List Suggestions

Given a source message, AI may suggest:

- normalized task title
- due date candidate
- assignee candidate

Constraints:

- suggestions must be optional and editable
- a failed AI call must not block manual list-item creation
- the final persisted list item must clearly reflect user-confirmed values

## B. Channel Execution Summary

AI may summarize recent execution state for a channel using persisted structured facts and recent channel context.

The summary should help answer:

- what new work appeared
- what completed
- what is blocked
- what needs attention next

Constraints:

- summary is advisory, not authoritative
- summary failure must not break Home or channel execution rendering

## Ownership And Delivery Model

This project will be executed with explicit ownership boundaries.

### Windsurf Ownership

All Web delivery belongs to Windsurf, including:

- channel execution panel UI
- message action UI for `Add to List`
- tool execution UI in channel context
- Home dashboard execution blocks
- client-side state/store wiring needed for those surfaces

### Codex Ownership

Codex owns backend and contract work, including:

- API contract definition and compatibility
- list/message reference persistence
- tool writeback semantics
- Home aggregation changes
- channel execution summary aggregation
- backend tests and integration verification

This ownership split should be preserved in the implementation plan to support parallel work without overlap.

## Error Handling

Execution features must degrade cleanly.

- if AI suggestion generation fails, users can still create the list item manually
- if execution summary generation fails, Home and channel views still render structured facts
- if tool writeback fails, the tool run record still exists and should surface a clear failure state
- if a deep-link target is stale or missing, the UI should fail soft and keep the user in the nearest valid context

## Testing Strategy

This phase should land with focused backend and Web verification.

### Backend

- list item creation with source message reference persists correctly
- Home aggregation returns execution blocks with stable shapes
- tool execution supports allowed writeback targets
- channel execution summary aggregation handles empty, active, and failure-heavy channels
- AI-assist endpoints fail soft without blocking the underlying user action

### Web

- channel execution panel renders correctly for list/tool states
- message-to-list flow can be completed from channel context
- Home execution blocks deep-link to the correct channel/object
- empty/loading/error states stay usable

## Implementation Notes

This phase should be delivered as one coherent execution slice, but implemented in two coordinated tracks:

1. Codex ships backend contracts, persistence, and tests
2. Windsurf ships all Web consumption and interaction flows

The implementation plan should preserve that split and explicitly avoid overlapping file ownership across tracks.
