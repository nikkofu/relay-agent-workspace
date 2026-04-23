# Phase 64: Slack Core Convergence

## Goal

Phase 64 moves Relay Agent Workspace out of the long Phase 63 AI automation thread and back into the broader product goal: a Slack-like workspace where AI-native events, human collaboration, files, workflows, activity, profiles, and knowledge all share one coherent operating model.

Phase 63 is closed except for hotfixes.

## Product Direction

Relay should feel like a team messenger and an agent operations console at the same time:

- humans, agents, files, tools, workflows, and knowledge events appear in the same workspace timeline
- activity is not fragmented across one-off panels
- user status, presence, profile, groups, channels, DMs, files, and artifacts are treated as first-class Slack-like primitives
- AI output is observable, auditable, and connected to channel/message context

## Phase 64 Workstreams

### 1. Unified Activity Feed

Backend target:

- `GET /api/v1/activity/feed?workspace_id=...`
- optional filters: `channel_id`, `dm_id`, `actor_id`, `type`, `limit`, `cursor`
- event sources: messages, mentions, reactions, files, artifacts, schedule bookings, compose activity, entity asks, automation jobs, workflow/tool runs

Frontend target:

- replace fragmented Activity/Home panels with one reusable Slack-like activity rail
- support filter chips and entity/channel deep links

### 2. Profiles, Status, Presence

Backend target:

- finish profile/status payload consistency across `GET /users`, `GET /users/:id`, `GET /presence`, and `GET /presence/bulk`
- expose recent activity/status metadata without mock fallbacks

Frontend target:

- user profile drawer, status update affordance, and team presence details

### 3. Files, Archives, Artifacts

Backend target:

- harden file archive/star/share/comment flows
- unify artifact/file references in activity and search
- make file-created and artifact-updated events available to the unified feed

Frontend target:

- file archive view, artifact references, and channel file surfaces should all hydrate from API data after refresh

### 4. Channels, DMs, Groups

Backend target:

- close remaining lifecycle gaps for channel membership, user groups, DMs, channel preferences, and notification state
- ensure all string primary IDs use UUID-style IDs, not timestamp-like IDs

Frontend target:

- Slack-like create/edit/invite flows should persist after refresh and never rely on local mock state

### 5. AI-Native Workspace Observability

Backend target:

- compose digest push or websocket delta
- external calendar adapter status on `AIScheduleBooking`
- automation/job events folded into the unified activity feed

Frontend target:

- #agent-collab becomes an observability cockpit, not a separate one-off dashboard

## Recommended Phase 64A

Start with the unified activity feed because it gives every later Slack-like surface a common backbone.

Phase 64A proposed scope:

- backend: `GET /api/v1/activity/feed`
- backend: include at least message, file, artifact, booking, compose activity, knowledge ask, and automation job items
- frontend: Activity/Home/Agent-Collab consume the same feed shape
- websocket: keep existing event-specific messages for now; do not add a new generic feed event until the REST feed is stable

## Phase 64C Completion Record

Phase 64C is the first feed-completion slice after the initial REST contract and UI integration.

Delivered in `v0.6.34`:

- `artifact_updated`
- `tool_run`
- `reply`
- `dm_message`
- `mention`
- `reaction`

Remembered next continuations from the user decision:

- `2`: continue strengthening the unified feed itself
- `3`: then broaden into larger Slack-like capability work outside the feed slice

## Phase 65A Continuation Record

Delivered in `v0.6.36` as the first concrete continuation of remembered follow-up `2`:

- durable `@user` mention persistence via `MessageMention`
- unified feed `mention_kind=user|entity`
- `GET /api/v1/mentions` and inbox mention alignment on the same persisted source
- realtime `mention.created`

This keeps the product on the chosen path:

- first, keep hardening the shared feed/inbox/message semantics
- next, continue into the broader Slack-like surface expansion tracked as remembered follow-up `3`
