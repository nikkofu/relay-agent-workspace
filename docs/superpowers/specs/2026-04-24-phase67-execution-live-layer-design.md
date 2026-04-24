# Phase 67 Execution Live Layer Design

## Goal

Upgrade Relay's newly shipped Channel Execution Hub from "usable" to "live, traceable, and management-ready" by adding realtime execution updates, durable source-message jump behavior, lighter mention badge hydration, and richer execution pulse signals on Home.

This phase intentionally treats those capabilities as one coordinated product slice:

1. make channel execution surfaces feel live without manual refresh
2. make message-linked work objects jump back to source context cleanly
3. decouple mention badge hydration from the full Home payload
4. strengthen Home execution pulse with directional management signals

## Why This Phase Exists

Phase 66 completed the first end-to-end version of the Channel Execution Hub:

- channels can now host lists and tool runs
- messages can become list items
- Home can summarize execution work
- `/ask` gives channels a native AI entry point

That closed the functional gap, but not yet the interaction-quality gap.

Current shortcomings:

- the execution panel does not yet behave with the same realtime confidence as message, knowledge, and activity surfaces
- message-linked tasks route back toward context, but do not yet reliably scroll to and highlight the exact source message
- the primary navigation mention badge is still heavier than it should be
- Home execution pulse shows state, but not enough momentum or direction

Phase 67 addresses those seams without introducing a new product area or a new work model.

## Scope

### Included

- realtime consumption of structured execution events for lists and tool runs
- stable channel execution panel live updates
- Home execution block live refresh or lightweight refresh coordination
- source-message jump and flash-highlight behavior for list items created from messages
- a lighter unread-mention count path than full `GET /api/v1/home`
- additive trend/momentum signals for `channel_execution_pulse[]`
- explicit execution split:
  - Gemini owns backend/API/tests
  - Windsurf owns Web/UI consumption
  - Codex owns planning, contract review, integration control

### Excluded

- new execution object types
- workflow-builder redesign
- generalized notification system rewrite
- complex analytics dashboards
- historical replay/backfill over all legacy execution events
- autonomous AI workflow chains

## Product Definition

### 1. Execution Should Feel Live

The execution panel should behave like a first-class collaborative surface.

When a user:

- creates a list
- adds a list item
- updates or completes a list item
- runs a tool
- receives a tool outcome

the relevant UI should update without depending on manual refresh.

The target experience is parity with Relay's existing live messaging and knowledge surfaces, not a polling-heavy approximation.

### 2. Message-Linked Work Should Return To Context

If a list item was created from a message, the user should be able to return directly to the source context.

The jump should:

- route to the right channel
- locate the right message
- visibly highlight it for a short period

This makes "message -> task" feel like one continuous workflow instead of two disconnected objects.

### 3. Mention Badges Should Be Cheap

The primary navigation mention badge is a lightweight awareness signal and should not require full Home hydration to stay accurate.

Home remains the management cockpit.
Badge hydration becomes a smaller, cheaper concern.

### 4. Home Should Show Direction, Not Just State

Execution pulse on Home should answer not just:

- how much work exists

but also:

- whether the channel is improving or degrading
- whether overdue work is rising
- whether tool failures are becoming a pattern

The first phase of that signal can remain lightweight and textual or badge-based.

## Core User Flows

### A. Live Execution In Channel

Two users are in the same channel.

One user:

- adds a list item from a message
- completes an item
- runs a tool with `message` or `list_item` writeback

The other user should see the relevant execution state update in the channel execution panel without refreshing.

### B. Return To Source Message

From a channel list row containing a `From msg` chip, the user clicks the chip.

The app should:

1. navigate to the source channel if needed
2. scroll the target message into view
3. apply a temporary highlight treatment
4. clear the highlight naturally

If the message cannot be found, the app should fail soft rather than leaving the user with a broken interaction.

### C. Mention Badge Refresh

The user receives or reads mentions.

The primary nav badge should refresh through:

- realtime updates when available
- a lighter refresh path when a fresh count is needed

without requiring a full Home dashboard fetch.

### D. Manage Execution Trend On Home

The user opens Home and checks `Channel Execution Pulse`.

They should quickly infer:

- which channels are accumulating open work
- which channels have rising overdue pressure
- which channels recently had tool failures

without needing a large analytics view.

## Architecture

## A. Realtime Execution Event Layer

Phase 67 should standardize a minimal execution event set for Web consumption.

Recommended event types:

- `list.item.created`
- `list.item.updated`
- `list.item.deleted`
- `tool.run.started`
- `tool.run.updated`
- `tool.run.completed`

If `tool.run.updated` already exists in practice, preserve compatibility and add the missing lifecycle events around it rather than replacing the contract.

Payloads should be shaped for direct store consumption:

- enough identifiers to locate channel/list/run scope
- the updated item/run payload
- timestamps for ordering/last-write-wins handling

### Design principle

Do not force Windsurf to re-fetch the full execution panel after every event unless an edge case requires it.

The first path should be append/update/remove against existing stores.

## B. Source Message Jump Layer

This phase does not require new backend persistence.

It does require a stable frontend convention for message anchors and highlight state.

Recommended approach:

- message rows expose stable DOM ids such as `msg-<message_id>`
- source-message links route to `/workspace?c=<channel_id>#msg-<message_id>`
- message list watches location/hash changes
- once the target message is present, it scrolls into view and receives a transient highlight class/state

If the current list view does not contain the message yet, Windsurf may need a minimal fetch or channel hydration path, but this should remain targeted and additive.

## C. Lightweight Mention Count Path

Provide a small API surface for unread mention count.

Freeze the contract as a dedicated endpoint:

- `GET /api/v1/me/unread-counts`

Minimum response:

```json
{
  "counts": {
    "unread_mention_count": 3
  }
}
```

This path should:

- return unread mention count only or as part of a very small payload
- be safe for frequent badge refresh
- stay consistent with durable mention/inbox semantics from Phase 65C

The primary-nav badge should prefer this endpoint over `GET /api/v1/home`.

## E. Home Refresh Coordination

Home execution blocks do not need per-event row-level websocket reconciliation in this phase.

Instead, refresh should be coordinated by lightweight invalidation triggers:

- on receipt of `list.item.created`
- on receipt of `list.item.updated`
- on receipt of `list.item.deleted`
- on receipt of `tool.run.started`
- on receipt of `tool.run.updated`
- on receipt of `tool.run.completed`

Windsurf may debounce these into a scoped Home refresh path, but should not infer a heavier strategy than:

- mark Home execution data stale
- refetch the Home execution aggregates once per short debounce window when the Home surface is mounted

## D. Execution Pulse Trend Layer

Extend `channel_execution_pulse[]` with trend-oriented fields that remain cheap to compute.

Recommended first-phase fields:

- `open_item_delta_7d`
- `overdue_delta_7d`
- `recent_tool_failure_count`
- optional lightweight sparkline array if it can be supplied cheaply

The frontend should not need to infer deltas by comparing multiple fetches.

## Data Contract Direction

### Realtime List Events

Minimum payload direction:

- `channel_id`
- `list_id`
- `item`
- `event_ts`

Where `item` includes the source-message fields already frozen in Phase 66:

- `source_message_id`
- `source_channel_id`
- `source_snippet`

### Realtime Tool Events

Minimum payload direction:

- `channel_id`
- `run`
- `event_ts`

Where `run` preserves Phase 66 writeback semantics:

- `writeback_target`
- `writeback`

### Lightweight Mention Count

Endpoint:

- `GET /api/v1/me/unread-counts`

Minimum payload:

- `counts.unread_mention_count`

### Home Pulse Trend

Each `channel_execution_pulse[]` row should keep existing identifiers and summary fields, then add trend fields additively.

## AI-Native Behavior

This phase is not about adding new autonomous AI workflows.

It is about strengthening the operability of existing AI-native surfaces:

- AI-created list suggestions should participate in live execution updates
- `/ask` channel usage should coexist with the lighter mention/navigation refresh path
- Home execution pulse can include recent AI tool failure signals if they are already part of tool-run outcomes

AI remains advisory and execution-adjacent, not a new source of truth.

## Error Handling

- if an execution websocket event is missed or malformed, Web may fall back to a scoped refetch for that channel surface
- if source-message jump target is absent, route to the correct channel and show a soft failure rather than a dead link
- if lightweight mention count fetch fails, keep the last known badge state or fall back to existing Home-derived count
- if trend fields are unavailable, Home should still render the existing pulse row without directional embellishment

## Testing Strategy

Phase 67 should land with four focused verification groups.

### 1. Realtime Execution Tests

- list item create/update/delete emits stable websocket payloads
- tool run lifecycle emits stable websocket payloads
- channel-scoped execution consumers can update without full refetch

### 2. Source Jump Tests

- message-linked list item routes to correct channel hash
- target message is scrolled into view
- target message receives temporary highlight state
- missing target degrades cleanly

### 3. Mention Badge Tests

- lightweight unread mention count endpoint returns durable count
- badge refresh does not require full Home fetch
- badge remains consistent with mention read flows

### 4. Execution Pulse Trend Tests

- pulse rows include trend fields for active channels
- empty channels do not show misleading trend output
- failure-heavy channels reflect recent tool-failure pressure

## Ownership And Delivery Model

### Gemini Ownership

Gemini owns backend/API/test work for Phase 67, including:

- websocket execution event contracts
- lightweight mention count path
- execution pulse trend fields
- backend tests

### Windsurf Ownership

Windsurf owns Web/UI work for Phase 67, including:

- realtime execution panel consumption
- source-message jump and highlight behavior
- primary-nav badge hydration changes
- Home pulse trend rendering

### Codex Ownership

Codex owns:

- phase planning
- contract freezing
- task decomposition
- collaboration document updates
- merge/integration control

## Release Intent

Phase 67 should ship as one coherent "execution quality" release, even if Gemini and Windsurf land it in multiple implementation slices.

The release story should be:

`Execution is now live, traceable, and easier to manage at a glance.`
