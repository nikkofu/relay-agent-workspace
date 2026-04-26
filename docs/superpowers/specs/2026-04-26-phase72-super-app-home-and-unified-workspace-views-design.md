# Phase 72 Super App Home And Unified Workspace Views Design

## Summary

Phase 72 shifts Relay from a collection of Slack-like collaboration surfaces toward a Super App Workspace foundation.

The first release focuses on two connected outcomes:

- make Home a Slack-like daily work entry point
- define a lightweight `WorkspaceView` registry that future Agents/tools can create, query, update, and link to channels

This phase does not build a full app runtime. It gives Relay a better Home and a common model for future list, calendar, search, report, form, and channel-message views.

## Product Definition

`Home keeps Workspace Overview at the top, and the rest becomes a Slack-like workbench centered on Today and My Work. A lightweight WorkspaceView registry prepares Relay for a Super App Workspace without trying to build every app surface at once.`

## Scope

### In Scope

- preserve the existing top `Workspace Overview` section
- redesign the lower Home experience around:
  - Today
  - My Work
  - Recent Channels
  - AI Suggestions
  - Apps & Tools
  - Activity
- make `Today` and `My Work` the primary below-overview sections
- use real Home/API data instead of static placeholders
- define a lightweight `WorkspaceView` registry model
- support first-release view types:
  - `list`
  - `calendar`
  - `search`
  - `report`
  - `form`
  - `channel_messages`
- expose backend APIs for creating/querying WorkspaceViews
- show lightweight view entry points in Home, especially through `Apps & Tools`

### Out of Scope

- full Super App runtime
- full app builder
- complex permission model
- complete UI for every view type
- BI/reporting engine
- workflow automation engine rewrite
- replacing existing channel/canvas/activity/dm/files/workflows pages

## Home Information Architecture

### Preserved Section

#### Workspace Overview

The current top overview remains the anchor.

It should continue to show high-level workspace state and should not be replaced by a new visual system in this phase.

### New Below-Overview Workbench

### Section Rules

Home section ownership must be deterministic.

Global rules:

- no item should appear in more than one of `Today`, `My Work`, and `Activity` unless the backend explicitly returns distinct item IDs for distinct meanings
- if an item qualifies for multiple sections, priority is:
  - Today
  - My Work
  - Activity
- each section should return backend-ranked items so Web does not invent ranking rules
- first-release section limits:
  - Today: 8 items
  - My Work: 8 items
  - Recent Channels: 6 channels
  - AI Suggestions: 5 suggestions
  - Activity: 10 events

#### Today

Primary purpose:

- answer what needs attention now

Candidate content:

- unread mentions
- failed AI-driven executions
- due or recently changed list items
- active workflow runs
- important channel updates
- recent AI-driven execution outcomes

Each item should be actionable and routeable.

Hard inclusion rule:

- Today is for time-sensitive or attention-needed work:
  - unread mentions
  - failed executions
  - due-today list items
  - active workflow runs
  - channel updates marked important by backend ranking

#### My Work

Primary purpose:

- aggregate execution objects connected to the current user

Candidate content:

- assigned list items
- user-confirmed AI-driven lists/workflows/messages
- workflow runs involving the user
- failed executions involving the user
- drafts involving the user

Hard inclusion rule:

- My Work is for user-owned or user-involved work that is not already in Today:
  - assigned list items not due today
  - user-confirmed AI-driven executions
  - user-owned drafts
  - workflow runs involving the user but not active/urgent

#### Recent Channels

Primary purpose:

- show channels with recent work changes, not just a static channel list

Candidate signals:

- new messages
- new files
- new list items
- tool runs
- AI execution events
- workflow activity

Hard inclusion rule:

- Recent Channels is channel-scoped and should return channels, not individual work objects

#### AI Suggestions

Primary purpose:

- show actionable AI suggestions grounded in workspace state

Examples:

- failed execution needs attention
- analysis result has a plan that has not been converted
- channel has unresolved mentions or blocked work

This section should not be generic tips.

Hard inclusion rule:

- AI Suggestions must be generated from concrete workspace state and include a source link or source object ID

#### Apps & Tools

Primary purpose:

- provide compact entry points into the Super App surfaces

First-release entries:

- Lists
- Calendar
- Search
- Reports
- Forms
- Workflows
- Files
- Tools

#### Activity

Primary purpose:

- summarize recent workspace activity

Priority event families:

- AI execution
- messages
- files
- workflows
- lists
- tool runs

Hard inclusion rule:

- Activity is a recent-event summary and should not duplicate items already promoted into Today/My Work unless the backend returns a different event ID

## WorkspaceView Registry

### Purpose

`WorkspaceView` is a lightweight registry object for future Super App views.

It gives Agents/tools a consistent target:

- create a view
- query a view
- update filters
- associate a view with a channel
- summarize a view

It is not a full app runtime in Phase 72.

### Conceptual Model

```json
{
  "id": "view-123",
  "title": "Customer Follow-up This Week",
  "view_type": "list|calendar|search|report|form|channel_messages",
  "source": "manual|agent|tool|system",
  "primary_channel_id": "channel-123",
  "filters": {
    "channel_id": "channel-123",
    "status": "open"
  },
  "actions": [
    { "type": "open" }
  ],
  "created_by": "user-123",
  "updated_at": "2026-04-26T00:00:00Z"
}
```

### View Types

#### `list`

Represents a structured list-oriented view.

#### `calendar`

Represents date/time-oriented data.

#### `search`

Represents a saved or dynamic search view.

#### `report`

Represents a statistics/reporting view.

#### `form`

Represents a data-entry or collection view.

#### `channel_messages`

Represents a channel-linked message data view.

### First-Release Contract Depth

The first release should keep `WorkspaceView` shallow:

- required:
  - `id`
  - `title`
  - `view_type`
  - `source`
  - `updated_at`
- optional:
  - `primary_channel_id`
  - `filters`
  - `actions`
  - `created_by`

The schema should allow future depth without forcing every view type to become fully functional immediately.

### Registry API Contract

First-release APIs:

- `GET /api/v1/workspace/views`
- `POST /api/v1/workspace/views`
- `GET /api/v1/workspace/views/:id`
- `PATCH /api/v1/workspace/views/:id`

List query parameters:

- `view_type`
- `primary_channel_id`
- `source`
- `limit`
- `cursor`

List response shape:

```json
{
  "views": [],
  "next_cursor": "string"
}
```

Default list limit:

- 20

Maximum list limit:

- 50

Validation:

- `view_type` must be one of the frozen first-release values
- `title` must be non-empty
- `filters` must be a shallow JSON object
- `actions` must be an array of shallow action descriptors
- unknown `actions[].type` values must be preserved but not executed

First-release action types:

- `open`
- `summarize`
- `update_filters`
- `open_channel`

First-release action boundary:

- actions are metadata/descriptors only
- actions do not execute workflows, reports, forms, or automations in Phase 72

## Agent / Tool Model

Phase 72 should define the target contract for future tool use.

Examples:

- create a list view for this week's customer follow-up
- create a report view from this channel's execution history
- update a saved search view's filters
- summarize a calendar view
- open the related channel for a view

The first release should support API-level creation and querying, with only lightweight UI exposure.

Hard boundary:

- Phase 72 only creates and displays registry entries
- view actions are descriptors for future tools
- no per-view runtime execution is required in this phase

## Backend / API Boundary

### Gemini Scope

Gemini owns backend/API/tests only.

Expected backend responsibilities:

- extend Home aggregation for the new below-overview workbench sections
- include enough real data for:
  - Today
  - My Work
  - Recent Channels
  - AI Suggestions
  - Apps & Tools
  - Activity
- add lightweight `WorkspaceView` model
- add WorkspaceView create/list/detail/update APIs
- validate allowed `view_type` values
- enforce list pagination and shallow `filters/actions` validation
- expose enough fields for Home to show view entries
- add tests for:
  - Home aggregation shape
  - Today/My Work data presence
  - WorkspaceView CRUD
  - `view_type` validation

### Backend Constraints

- do not build a full app runtime
- do not implement complete UI-specific logic for every view type
- do not replace existing Home APIs wholesale if extension is sufficient
- do not break existing Home consumers
- do not execute WorkspaceView `actions` in Phase 72

## Web / UI Boundary

### Windsurf Scope

Windsurf owns all Web/UI delivery.

Expected Web responsibilities:

- keep Workspace Overview at the top of Home
- replace the lower Home layout with the new workbench structure
- make Today and My Work the dominant sections
- render Recent Channels, AI Suggestions, Apps & Tools, and Activity as secondary sections
- use backend-provided Home data
- show lightweight WorkspaceView entry points
- preserve existing navigation to:
  - channels
  - lists
  - workflows
  - files
  - activity
  - tools

First-release Apps & Tools behavior:

- entries are navigation or registry entry points
- Calendar, Reports, and Forms may show shells or registered views
- they must not pretend to be full products in Phase 72

### Web Constraints

- do not redesign the whole workspace shell in this phase
- do not build a full app marketplace
- do not add full calendar/report/form products yet
- do not execute WorkspaceView actions from Home in Phase 72
- do not use static placeholder content where backend data exists
- do not break existing Channel, Canvas, Activity, DM, Files, Workflows pages

## AI-Native Behavior

This phase supports AI-native workflows by giving Agents/tools a consistent workspace object to target.

Near-term AI-native actions enabled by the contract:

- create a WorkspaceView from a conversation
- link a WorkspaceView to a channel
- summarize a view
- update view filters
- recommend a view from Home

Home should increasingly show AI-driven work outcomes and suggestions, but the first release should stay grounded in real existing data.

Phase 72 boundary:

- Agent/tool integration is registry-level only
- Agents may create or update WorkspaceView records through API contracts
- Agents do not run full view-specific products through this contract yet

## Error Handling

Required behaviors:

- if a Home section has no data, render a compact empty state without looking broken
- if WorkspaceView APIs fail, Home should still render existing non-view sections
- invalid `view_type` must be rejected server-side
- unknown future `view_type` values should not break existing clients
- route links should degrade safely if a target object no longer exists

## Acceptance Criteria

### Core

1. Home keeps the top Workspace Overview section.
2. Home below the overview is reorganized into a Slack-like workbench.
3. Today and My Work are the primary below-overview sections.
4. Today/My Work consume real backend data rather than static placeholders.
5. Apps & Tools includes entries for Lists, Calendar, Search, Reports, Forms, Workflows, Files, and Tools.
6. Backend exposes a lightweight WorkspaceView registry with allowed view types:
   - `list`
   - `calendar`
   - `search`
   - `report`
   - `form`
   - `channel_messages`
7. WorkspaceView CRUD/query exists at API level even if each view type is not fully built out.
8. Existing Channel, Canvas, Activity, DM, Files, and Workflows pages continue to work.
9. Home section limits, ordering, and dedupe rules are backend-driven and testable.
10. WorkspaceView actions are stored as metadata only and are not executed in Phase 72.

### Architecture

1. Phase 72 creates a Super App foundation without building a full app runtime.
2. Home becomes the main daily work entry point.
3. WorkspaceView gives Agents/tools a consistent target for future view creation and query.

## Testing Strategy

### Backend

- Home aggregation tests for new section fields
- Today/My Work data shape tests
- WorkspaceView model/API tests
- `view_type` validation tests
- backwards-compatibility tests for existing Home response fields

### Web

- Home layout tests for section presence and ordering
- Today/My Work rendering tests with real API-shaped data
- Apps & Tools entry tests
- WorkspaceView entry rendering tests
- navigation/link tests for existing destinations
- regression checks for Channel/Canvas/Activity/DM/Files/Workflows routes where feasible

## Release Notes Guidance

This release should be described as:

`Home has been reorganized into a Slack-like workbench centered on Today and My Work, while Relay gains a lightweight WorkspaceView registry for future list, calendar, search, report, form, and channel-linked views.`

## Ownership

- Gemini: backend/API/test support only
- Windsurf: all Web/UI implementation
- Codex: Home information architecture, WorkspaceView contract, collaboration docs, integration control, follow-up phase planning
