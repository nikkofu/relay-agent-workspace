# Phase 8: Backend Expansion Target

## 1. Current Reality

`Relay Agent Workspace` is no longer a pure mock-only UI. The project now ships a working monorepo with:

- `apps/web`: Next.js workspace UI
- `apps/api`: Go + Gin backend
- SQLite persistence through GORM
- WebSocket realtime fanout
- LLM gateway with:
  - `openai`
  - `openai-compatible`
  - `openrouter`
  - `gemini`

Already delivered backend capabilities include:

- current user and users list
- organizations, teams, workspaces, channels
- channel messages and thread replies
- AI SSE execution
- AI config and user AI settings
- persisted reactions, pinning, later, unread, and AI feedback
- realtime events for:
  - `message.created`
  - `reaction.updated`
  - `message.updated`
  - `message.deleted`
  - `agent_collab.sync`

So the purpose of this document is no longer “what do we need to start backend work at all”.  
It is now: **what still needs to be added to reach the broader Phase 8 dynamic product target**.

## 2. UI Surfaces Driving The Remaining API Work

The remaining backend work is still driven directly by the shipped UI in `apps/web`:

- Workspace shell:
  - `apps/web/app/workspace/layout.tsx`
  - `apps/web/app/workspace/page.tsx`
- Message experience:
  - `apps/web/components/layout/message-area.tsx`
  - `apps/web/components/layout/thread-panel.tsx`
  - `apps/web/components/message/message-composer.tsx`
  - `apps/web/components/message/message-item.tsx`
  - `apps/web/components/message/message-actions.tsx`
- AI surfaces:
  - `apps/web/components/ai-chat/ai-chat-panel.tsx`
  - `apps/web/components/ai-chat/ai-thread-summary.tsx`
  - `apps/web/components/ai-chat/ai-slash-command.tsx`
  - `apps/web/hooks/use-ai-chat.ts`
- Navigation and productivity views:
  - `apps/web/app/workspace/activity/page.tsx`
  - `apps/web/app/workspace/dms/page.tsx`
  - `apps/web/app/workspace/later/page.tsx`
  - `apps/web/components/search/search-dialog.tsx`
  - `apps/web/components/layout/canvas-panel.tsx`

## 3. Shipped API Baseline

Current backend surface already available:

- `GET /ping`
- `GET /api/v1/me`
- `GET /api/v1/home`
- `PATCH /api/v1/me/settings`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `PATCH /api/v1/users/:id`
- `PATCH /api/v1/users/:id/status`
- `GET /api/v1/user-groups/mentions`
- `GET /api/v1/user-groups`
- `POST /api/v1/user-groups`
- `GET /api/v1/user-groups/:id/members`
- `POST /api/v1/user-groups/:id/members`
- `DELETE /api/v1/user-groups/:id/members/:userId`
- `GET /api/v1/user-groups/:id`
- `PATCH /api/v1/user-groups/:id`
- `DELETE /api/v1/user-groups/:id`
- `GET /api/v1/workflows`
- `GET /api/v1/workflows/runs`
- `GET /api/v1/workflows/runs/:id/logs`
- `GET /api/v1/workflows/runs/:id`
- `DELETE /api/v1/workflows/runs/:id`
- `POST /api/v1/workflows/runs/:id/cancel`
- `POST /api/v1/workflows/runs/:id/retry`
- `POST /api/v1/workflows/:id/runs`
- `GET /api/v1/tools`
- `GET /api/v1/tools/runs`
- `GET /api/v1/tools/runs/:id`
- `POST /api/v1/tools/:id/execute`
- `GET /api/v1/orgs`
- `GET /api/v1/orgs/:id/teams`
- `POST /api/v1/orgs/:id/agents`
- `GET /api/v1/workspaces`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`
- `GET /api/v1/channels`
- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `POST /api/v1/channels`
  - validates `workspace_id`
  - rejects unknown workspace IDs
  - channel rows with legacy mock `workspace_id=ws_1` are repaired to `ws-1` on API startup
- `GET /api/v1/dms`
- `POST /api/v1/dms`
- `GET /api/v1/dms/:id/messages`
- `POST /api/v1/dms/:id/messages`
- `GET /api/v1/activity`
- `GET /api/v1/inbox`
- `GET /api/v1/mentions`
- `POST /api/v1/notifications/read`
- `GET /api/v1/notifications/preferences`
- `PATCH /api/v1/notifications/preferences`
- `GET /api/v1/later`
- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`
- `GET /api/v1/starred`
- `POST /api/v1/channels/:id/star`
- `GET /api/v1/pins`
- `GET /api/v1/pins?channel_id=...`
- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`
- `GET /api/v1/search`
- `GET /api/v1/search/messages/by-entity`
- `GET /api/v1/search/suggestions`
- `GET /api/v1/citations/lookup`
- `GET /api/v1/channels/:id/knowledge`
- `GET /api/v1/channels/:id/knowledge/summary`
- `GET /api/v1/channels/:id/knowledge/digest`
- `POST /api/v1/channels/:id/knowledge/digest/publish`
- `GET /api/v1/channels/:id/knowledge/digest/schedule`
- `PUT /api/v1/channels/:id/knowledge/digest/schedule`
- `DELETE /api/v1/channels/:id/knowledge/digest/schedule`
- `GET /api/v1/knowledge/inbox`
- `GET|PATCH /api/v1/workspace/settings`
- `GET /api/v1/users/me/knowledge/followed`
- `PATCH /api/v1/users/me/knowledge/followed/bulk`
- `GET /api/v1/users/me/knowledge/followed/stats`
- `GET /api/v1/knowledge/entities`
- `GET /api/v1/knowledge/trending`
- `GET /api/v1/knowledge/entities/suggest`
- `POST /api/v1/knowledge/entities/match-text`
- `POST /api/v1/knowledge/entities`
- `GET /api/v1/knowledge/entities/:id`
- `GET /api/v1/knowledge/entities/:id/activity`
- `GET /api/v1/knowledge/entities/:id/hover`
- `POST /api/v1/knowledge/entities/:id/share`
- `POST /api/v1/knowledge/entities/:id/follow`
- `DELETE /api/v1/knowledge/entities/:id/follow`
- `PATCH /api/v1/knowledge/entities/:id`
- `GET /api/v1/knowledge/entities/:id/refs`
- `POST /api/v1/knowledge/entities/:id/refs`
- `GET /api/v1/knowledge/entities/:id/timeline`
- `POST /api/v1/knowledge/entities/:id/events`
- `GET /api/v1/knowledge/entities/:id/links`
- `POST /api/v1/knowledge/links`
- `POST /api/v1/knowledge/events/ingest`
- `GET /api/v1/knowledge/entities/:id/graph`
- `GET /api/v1/lists`
- `POST /api/v1/lists`
- `GET /api/v1/lists/:id`
- `PATCH /api/v1/lists/:id`
- `DELETE /api/v1/lists/:id`
- `POST /api/v1/lists/:id/items`
- `PATCH /api/v1/lists/:id/items/:itemId`
- `DELETE /api/v1/lists/:id/items/:itemId`
- `GET /api/v1/messages`
- `GET /api/v1/messages/:id/thread`
- `POST /api/v1/messages`
- message metadata enrichment:
  - `message.metadata.entity_mentions`
  - `message.metadata.knowledge_digest`
- `DELETE /api/v1/messages/:id`
- `POST /api/v1/messages/:id/reactions`
- `POST /api/v1/messages/:id/pin`
- `POST /api/v1/messages/:id/later`
- `POST /api/v1/messages/:id/unread`
- `GET /api/v1/ai/config`
- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`
- `POST /api/v1/ai/execute`
- `POST /api/v1/ai/feedback`
- `GET /api/v1/realtime`

## 4. Remaining Backend Targets

### 4.1 Presence And Typing

Baseline support now exists for:

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`
- realtime events:
  - `presence.updated`
  - `typing.updated`

Likely follow-ups:

- debounce and timeout conventions for typing indicators
- DM-scoped presence list queries
- explicit presence preference/mute rules
- richer status preset catalogs and scheduled status windows

### 4.2 AI Conversation State

Baseline support now exists for:

- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`
- persisted AI history behind `POST /api/v1/ai/execute`
- `GET /api/v1/messages/:id/summary`
- `POST /api/v1/messages/:id/summary`
- `GET /api/v1/channels/:id/summary`
- `POST /api/v1/channels/:id/summary`

Likely follow-ups:

- command execution and tool orchestration history
- artifact outputs linked to AI conversations
- intelligent or semantic retrieval routed through persisted workspace knowledge

### 4.3 Canvas / Artifact Lifecycle

Baseline support now exists for:

- `GET /api/v1/artifacts`
- `GET /api/v1/artifacts/templates`
- `POST /api/v1/artifacts/from-template`
- `POST /api/v1/artifacts`
- `GET /api/v1/artifacts/:id`
- `PATCH /api/v1/artifacts/:id`
- `GET /api/v1/artifacts/:id/versions`
- `GET /api/v1/artifacts/:id/versions/:version`
- `GET /api/v1/artifacts/:id/diff/:from/:to`
- `POST /api/v1/artifacts/:id/restore/:version`
- `POST /api/v1/artifacts/:id/duplicate`
- `POST /api/v1/ai/canvas/generate`

Likely follow-ups:

- artifact restore UI integration and restore audit affordances
- richer structured diff blocks or semantic diff grouping for review UIs
- canvas collaboration cursors or presence
- deeper artifact backlinks such as thread-level and DM-level references
- artifact templates and generation presets
- richer template galleries grouped by workflow or team use case
- richer fork lineage metadata and compare-original affordances

### 4.4 Files

Baseline support now exists for:

- `POST /api/v1/files/upload`
- `GET /api/v1/files`
- `GET /api/v1/files/archive`
- `GET /api/v1/files/:id`
- `GET /api/v1/files/:id/audit`
- `GET /api/v1/files/:id/content`
- `GET /api/v1/files/:id/preview`
- `PATCH /api/v1/files/:id/archive`
- `PATCH /api/v1/files/:id/retention`
- `DELETE /api/v1/files/:id`

Likely follow-ups:

- image thumbnails and richer preview transforms beyond direct image/PDF URLs
- richer attachment previews in message surfaces
- richer archive filters such as date windows and channel grouping
- retention enforcement jobs and deleted-file governance surfaces
- file preview transforms and document snapshot generation

### 4.5 Slack Parity Layer

The current workspace is already functionally dynamic, but there is still a layer of classic Slack-style collaboration APIs that should be planned explicitly instead of being discovered ad hoc.

Recommended additions:

- channel membership and directory:
  - `GET /api/v1/channels/:id/members`
  - `POST /api/v1/channels/:id/members`
  - `DELETE /api/v1/channels/:id/members/:userId`
- channel metadata and admin actions:
  - `PATCH /api/v1/channels/:id`
  - `GET /api/v1/channels/:id/preferences`
  - `PATCH /api/v1/channels/:id/preferences`
  - `POST /api/v1/channels/:id/leave`
  - fields such as:
    - `topic`
    - `purpose`
    - `is_archived`
- invitations:
  - `POST /api/v1/workspaces/:id/invites`
  - `GET /api/v1/workspaces/:id/invites`
- sidebar and discovery surfaces:
  - shipped baseline:
    - `GET /api/v1/starred`
    - `POST /api/v1/channels/:id/star`
    - `GET /api/v1/pins`
  - likely follow-up:
    - pinned filters by channel or user
    - star sorting and manual ordering
- notification and inbox surfaces:
  - shipped baseline:
    - `GET /api/v1/inbox`
    - `GET /api/v1/mentions`
    - `POST /api/v1/notifications/read`
  - likely follow-up:
    - notification preferences
    - mute rules
    - richer read-state websocket reconciliation and per-surface badges
- drafts:
  - shipped baseline:
    - `GET /api/v1/drafts`
    - `PUT /api/v1/drafts/:scope`
    - `DELETE /api/v1/drafts/:scope`
  - likely follow-up:
    - draft cleanup triggers tied to send/discard flows

### 4.6 Home, Directory, And Profile Layer

Baseline support now exists for:

- `GET /api/v1/home`
- `GET /api/v1/users`
- `GET /api/v1/users/:id`
- `PATCH /api/v1/users/:id/status`
- `GET /api/v1/user-groups/mentions`
- `GET /api/v1/user-groups`
- `GET /api/v1/user-groups/:id/members`
- `GET /api/v1/user-groups/:id`

Likely follow-ups:

- deeper profile editing such as phone, bio, and avatar workflows
- dedicated home widgets for onboarding, workflows, unread work, and structured work summaries
- direct dashboard aliases such as `stats`, `recent_activity`, and top-level `recent_artifacts` for UI-oriented home hydration
- deeper directory facets such as title and working-hours filters
- user-group membership roles and permissions
- profile visibility/privacy controls

### 4.7 Workflow And Tool Registry

Baseline support now exists for:

- `GET /api/v1/workflows`
- `GET /api/v1/workflows/runs`
- `GET /api/v1/workflows/runs/:id`
- `GET /api/v1/workflows/runs/:id/logs`
- `DELETE /api/v1/workflows/runs/:id`
- `POST /api/v1/workflows/runs/:id/cancel`
- `POST /api/v1/workflows/runs/:id/retry`
- `POST /api/v1/workflows/:id/runs`
- `GET /api/v1/tools`
- `GET /api/v1/tools/runs`
- `GET /api/v1/tools/runs/:id`
- `POST /api/v1/tools/:id/execute`

Likely follow-ups:

- workflow templates
- tool execution audit history
- channel- or DM-scoped tool availability
- agent/tool routing metadata shared with the AI layer
- richer workflow telemetry such as token/tool cost and agent handoff spans
- workflow retention controls for automation history
- home and activity surfaces that foreground recent workflow and tool execution context

### 4.8 Structured Work Objects

Baseline support now exists for:

- `GET /api/v1/lists`
- `POST /api/v1/lists`
- `GET /api/v1/lists/:id`
- `PATCH /api/v1/lists/:id`
- `DELETE /api/v1/lists/:id`
- `POST /api/v1/lists/:id/items`
- `PATCH /api/v1/lists/:id/items/:itemId`
- `DELETE /api/v1/lists/:id/items/:itemId`
- virtual `GET /api/v1/artifacts/new-doc`
- `GET /api/v1/artifacts/templates`
- `POST /api/v1/artifacts/from-template`
- `POST /api/v1/artifacts/:id/duplicate`

Likely follow-ups:

- list-to-message linkage and deeper structured activity context beyond item completion
- assignee filters and due-date reminder jobs for list items
- richer workspace and channel home widgets backed by structured lists, tool runs, and files
- richer template bundles for product, design, engineering, and incident response
- fork lineage, ownership transfer, and review workflows for copied canvases

### 4.9 Notification Preferences And Mute Rules

Baseline support now exists for:

- `GET /api/v1/notifications/preferences`
- `PATCH /api/v1/notifications/preferences`

Likely follow-ups:

- mute rules for DMs and user groups
- schedule-based do-not-disturb windows
- per-surface badge suppression preferences
- treat string IDs as opaque identifiers; do not depend on timestamp-derived shapes in UI routing or stores

### 4.10 Knowledge And Wiki Layer

Baseline support now exists for:

- `GET /api/v1/citations/lookup`
- `GET /api/v1/channels/:id/knowledge`
- `GET /api/v1/channels/:id/knowledge/summary`
- `GET /api/v1/channels/:id/knowledge/digest`
- `POST /api/v1/channels/:id/knowledge/digest/publish`
- `GET /api/v1/channels/:id/knowledge/digest/schedule`
- `PUT /api/v1/channels/:id/knowledge/digest/schedule`
- `DELETE /api/v1/channels/:id/knowledge/digest/schedule`
- `GET /api/v1/knowledge/inbox`
- `GET|PATCH /api/v1/workspace/settings`
- `GET /api/v1/users/me/knowledge/followed`
- `PATCH /api/v1/users/me/knowledge/followed/bulk`
- `GET /api/v1/users/me/knowledge/followed/stats`
- `GET /api/v1/knowledge/entities`
- `GET /api/v1/knowledge/trending`
- `GET /api/v1/knowledge/entities/suggest`
- `POST /api/v1/knowledge/entities/match-text`
- `POST /api/v1/knowledge/entities`
- `GET /api/v1/knowledge/entities/:id`
- `GET /api/v1/knowledge/entities/:id/activity`
- `GET /api/v1/knowledge/entities/:id/hover`
- `POST /api/v1/knowledge/entities/:id/share`
- `POST /api/v1/knowledge/entities/:id/follow`
- `DELETE /api/v1/knowledge/entities/:id/follow`
- `PATCH /api/v1/knowledge/entities/:id`
- `GET /api/v1/knowledge/entities/:id/refs`
- `POST /api/v1/knowledge/entities/:id/refs`
- `GET /api/v1/knowledge/entities/:id/timeline`
- `POST /api/v1/knowledge/entities/:id/events`
- `GET /api/v1/knowledge/entities/:id/links`
- `POST /api/v1/knowledge/links`
- `POST /api/v1/knowledge/events/ingest`
- `GET /api/v1/knowledge/entities/:id/graph`
- message metadata enrichment:
  - `message.metadata.entity_mentions`
  - `message.metadata.knowledge_digest`
- channel summary alert metadata:
  - `summary.velocity.recent_window_days`
  - `summary.velocity.previous_ref_count`
  - `summary.velocity.recent_ref_count`
  - `summary.velocity.delta`
  - `summary.velocity.is_spiking`
- realtime events:
  - `knowledge.entity.created`
  - `knowledge.entity.updated`
  - `knowledge.entity.ref.created`
  - `knowledge.trending.changed`
  - `knowledge.event.created`
  - `knowledge.link.created`
  - `knowledge.digest.published`

Likely follow-ups:

- message rendering for first-class `@entity:` mentions after autocomplete selection
- entity-centric search UX and channel/header drilldowns using `/api/v1/search/messages/by-entity`
- richer hover-card UI that consumes `/api/v1/knowledge/entities/:id/hover`
- digest review/publish UX for weekly or daily channel summaries using `/api/v1/channels/:id/knowledge/digest`
- entity activity spike websockets for per-user knowledge watchlists
- passive composer reverse-lookup hints that convert plain-text entity matches into `@entity` mentions
- inbox detail/context APIs for digest drilldowns
- richer entity disambiguation and alias support beyond exact title matching
- entity-aware search ranking that blends citations, refs, files, and channel-local activity
- knowledge summaries that merge static wiki state with external live business signals
- deeper graph traversal, filtering, and graph-to-message/file backreferences
- explicit knowledge permissions and workspace-wide curation workflows
- file extraction review/confirmation flow before bulk entity auto-linking

## 5. Realtime Target State

Current websocket support is enough for message creation and interaction sync.  
The broader target should cover:

- `message.created`
- `message.updated`
- `message.deleted`
- `reaction.updated`
- `presence.updated`
- `typing.updated`
- `thread.updated`
- `ai.delta`
- `ai.reasoning`
- `agent_collab.sync`
- later agent runtime events such as:
  - `agent.run.updated`
  - `agent.step.updated`
- `artifact.updated`

## 6. Recommended Next Phases

Recommended sequence from here:

1. Expand realtime into presence, typing, and richer thread updates.
2. Add the Slack parity layer:
   - members
   - invites
   - channel metadata
   - inbox and mentions
   - drafts and stars
3. Persist AI conversation state and summaries.
4. Introduce artifact and file lifecycle APIs.
5. Deepen the knowledge layer with entity-aware mentions, summaries, graph views, and semantic retrieval.
   - Current delivered extensions now also include digest scheduling, digest inbox aggregation, inbox drill-down detail, entity follow, composer reverse lookup, and schedule dry-run preview.
   - The next delivered layer now also includes per-follow notification levels, realtime entity spike alerts, bulk follow operations, workspace-level spike tuning, entity activity timeseries, and a trending knowledge feed.
   - Phase 61 extends this into AI-native knowledge operations with entity briefs, weekly followed-knowledge briefs, historical activity backfill, followed-stats realtime events, and bulk presence hydration.
  - Phase 62 adds cache-first brief hydration, entity-brief realtime completion events, and atomic notification bulk-read for larger inbox workflows.
  - Phase 63A adds entity-scoped grounded Q&A, weekly brief snapshot sharing, and `knowledge.entity.brief.changed` invalidation broadcasts so the knowledge layer can move from cached reading into ask/share/refresh flows.
  - Phase 63B adds grounded `POST /api/v1/ai/compose` for channel/thread reply suggestions, connecting composer UX to recent conversation context and workspace knowledge refs.
6. Add richer search layers such as suggestions and semantic retrieval.
7. Move toward explicit agent runtime APIs once the collaboration foundation is stable.

## 7. Summary

Phase 8 is no longer about “starting a backend”. That phase is already underway and shipped.  
The broader target now is to evolve Relay from:

- a realtime AI-native workspace with foundational APIs

into:

- a full collaboration platform for humans and agents, with search, history, artifacts, agent execution, and richer realtime semantics.
