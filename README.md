# Relay Agent Workspace

Relay Agent Workspace is an AI-native collaboration product where humans and agents work in the same messaging surface. It combines team chat, AI execution, artifacts, and realtime coordination into a single workspace model.

## Product Positioning

Relay sits between a team messenger and an agent operating console:

- messaging is the shared coordination layer
- AI chat is part of the conversation flow, not a detached sidebar
- artifacts live beside channels and threads
- agent actions are designed to become visible, auditable workspace events

For product, design, and marketing, the short version is:

> `Relay` is a shared workspace for humans and agents to chat, think, execute, and hand off work in real time.

## Core Capabilities

- Slack-style workspace, channel, DM, activity, and later views
- Rich message composer with slash commands, mentions, and formatting
- AI assistant panel with provider-aware streaming chat
- Thread panel and AI thread summary surface
- Canvas / artifact panel for AI-generated collaborative outputs
- Realtime collaboration sync via WebSocket events
- Configurable LLM gateway across OpenAI, OpenAI-compatible, OpenRouter, and Gemini

## Current Status

`v0.6.30` is the current release line and includes:

- Go + Gin API service under `apps/api`
- SQLite persistence via GORM
- seed data for org, team, user, agent, workspace, channel, and message
- realtime websocket endpoint for workspace event fanout
- REST endpoints for org, team, agent, workspace, channel, and message flows
- hardened channel creation persistence so newly created channels survive refresh and reject unknown workspace IDs
- startup repair for legacy `ws_1` mock workspace channel rows created by earlier frontend fallback logic
- v0.5.94 UI bug fixes for Home scrolling, DM overlay behavior, composer draft clearing, AI avatar, and Agent-Collab statistics
- lint hotfix for `message-composer.tsx` after the v0.5.94 draft-restore change
- `AGENT-COLLAB.md` watcher and `agent_collab.sync` websocket broadcast for the `#agent-collab` dashboard
- dynamic Agent-Collab Hub APIs for members, comm-log persistence, and realtime hub refresh
- hardened Agent-Collab Hub payloads with direct-message `to` fields and normalized member tool arrays
- file collaboration APIs for comments, shares, stars, and knowledge-oriented metadata
- message-level rich file attachment payloads for inline channel/thread rendering
- `GET /api/v1/messages/:id/files` for message-scoped file card hydration
- file extraction lifecycle, chunk indexing, and file-content search/citation APIs
- unified AI citation lookup across file chunks, messages, threads, and artifact sections
- `apps/api/internal/knowledge/` evidence lookup layer for later wiki and graph phases
- first-class knowledge entity APIs for wiki-style entity pages, refs, timeline, links, and graph previews
- knowledge entity creation now defaults to the primary workspace when UI flows omit `workspace_id`, and preserves `tags[]` into entity metadata
- realtime knowledge entity/wiki websocket events for live entity, ref, event, and link refresh
- live knowledge event ingestion via `POST /api/v1/knowledge/events/ingest`
- deterministic knowledge entity auto-linking from newly created messages and uploaded files
- structured message `entity_mentions` metadata for explicit `@Entity Title` references
- structured message `knowledge_digest` metadata for published channel digest messages
- channel-scoped knowledge context via `GET /api/v1/channels/:id/knowledge`
- channel-scoped knowledge summary via `GET /api/v1/channels/:id/knowledge/summary`
- channel auto-summarize settings and execution via `GET|PUT|POST /api/v1/channels/:id/knowledge/auto-summarize`
- realtime channel summary refresh broadcasts via websocket `channel.summary.updated`
- channel-scoped knowledge digest preview via `GET /api/v1/channels/:id/knowledge/digest`
- channel-scoped knowledge digest publish flow via `POST /api/v1/channels/:id/knowledge/digest/publish`
- channel-scoped knowledge digest scheduling via `GET|PUT|DELETE /api/v1/channels/:id/knowledge/digest/schedule`
- digest schedule dry-run preview via `POST /api/v1/channels/:id/knowledge/digest/preview-schedule`
- scoped knowledge entity autocomplete via `GET /api/v1/knowledge/entities/suggest`
- entity hover enrichment via `GET /api/v1/knowledge/entities/:id/hover`
- entity-centric message discovery via `GET /api/v1/search/messages/by-entity`
- cross-channel knowledge inbox via `GET /api/v1/knowledge/inbox`
- knowledge inbox detail drill-down via `GET /api/v1/knowledge/inbox/:id`
- `GET /api/v1/me/settings` for cross-device settings hydration
- per-follow notification settings via `PATCH /api/v1/users/me/knowledge/followed/:id`
- bulk follow notification updates via `PATCH /api/v1/users/me/knowledge/followed/bulk`
- followed knowledge aggregate stats via `GET /api/v1/users/me/knowledge/followed/stats`
- workspace knowledge alert settings via `GET|PATCH /api/v1/workspace/settings`
- per-entity activity timeseries via `GET /api/v1/knowledge/entities/:id/activity`
- workspace trending knowledge entities via `GET /api/v1/knowledge/trending`
- realtime trending rerank broadcasts via websocket `knowledge.trending.changed`
- entity deeplink sharing via `POST /api/v1/knowledge/entities/:id/share`
- AI-generated knowledge entity briefs via `POST /api/v1/knowledge/entities/:id/brief`
- cached knowledge entity brief hydration via `GET /api/v1/knowledge/entities/:id/brief`
- per-user followed knowledge weekly briefs via `POST /api/v1/knowledge/weekly-brief`
- cached weekly followed-knowledge brief hydration via `GET /api/v1/knowledge/weekly-brief`
- entity-scoped grounded Q&A via `POST /api/v1/knowledge/entities/:id/ask`
- streaming entity-scoped Q&A via `POST /api/v1/knowledge/entities/:id/ask/stream`
- persisted entity Ask AI history via `GET /api/v1/knowledge/entities/:id/ask/history`
- grounded channel/thread/DM composer suggestions via `POST /api/v1/ai/compose`
- streaming grounded channel/thread/DM composer suggestions via `POST /api/v1/ai/compose/stream`
- composer intent variants for `reply`, `summarize`, `followup`, and `schedule`
- structured schedule intent slots via `compose.proposed_slots[]`
- realtime compose suggestion broadcasts via websocket `knowledge.compose.suggestion.generated`
- persisted compose suggestion activity via `GET /api/v1/ai/compose/activity`
- aggregated compose suggestion analytics via `GET /api/v1/ai/compose/activity/digest`
- per-suggestion composer feedback capture via `POST /api/v1/ai/compose/:id/feedback`
- composer feedback aggregation via `GET /api/v1/ai/compose/:id/feedback/summary`
- entity brief automation state via:
  - `GET /api/v1/knowledge/entities/:id/brief/automation`
  - `POST /api/v1/knowledge/entities/:id/brief/automation/run`
  - `POST /api/v1/knowledge/entities/:id/brief/automation/retry`
- background entity brief regeneration websocket events:
  - `knowledge.entity.brief.regen.queued`
  - `knowledge.entity.brief.regen.started`
  - `knowledge.entity.brief.regen.failed`
- interval-driven channel auto-summary worker that honors `min_new_messages` and emits websocket `channel.summary.updated` without manual clicks
- AI schedule booking APIs:
  - `POST /api/v1/ai/schedule/book`
  - `GET /api/v1/ai/schedule/bookings`
  - `GET /api/v1/ai/schedule/bookings/:id`
  - `POST /api/v1/ai/schedule/bookings/:id/cancel`
- AI schedule booking websocket events:
  - `schedule.event.booked`
  - `schedule.event.cancelled`
- workspace automation audit API via `GET /api/v1/ai/automation/jobs`
- recent cross-entity ask feed via `GET /api/v1/knowledge/ask/recent`
- recent cross-entity ask feed rows include `entity_title` and `entity_kind` for stable UI rendering
- realtime entity ask answer broadcasts via websocket `knowledge.entity.ask.answered`
- weekly brief snapshot sharing via `POST /api/v1/knowledge/weekly-brief/:id/share`
- historical knowledge activity backfill status and execution via `GET /api/v1/knowledge/entities/:id/activity/backfill-status` and `POST /api/v1/knowledge/entities/:id/activity/backfill`
- realtime followed-stats deltas via websocket `knowledge.followed.stats.changed`
- realtime entity-brief generation broadcasts via websocket `knowledge.entity.brief.generated`
- realtime entity-brief invalidation broadcasts via websocket `knowledge.entity.brief.changed`
- atomic notification bulk-read via `POST /api/v1/notifications/bulk-read`
- reconnect-friendly bulk presence hydration via `GET /api/v1/presence/bulk`
- realtime entity spike alerts via websocket `knowledge.entity.activity.spiked`
- knowledge summary velocity/anomaly fields for channel-header trend badges
- citation lookup entity hydration from canonical `KnowledgeEntityRef` message/file associations
- channel-aware entity ranking for `@entity:` composer autocomplete and knowledge side-panel summary cards
- richer knowledge graph payloads with edge weight, direction, role, and typed reference-node metadata
- real extraction support for `txt`, `md`, `pdf`, `docx`, `xlsx`, and `pptx`
- OCR provider abstraction for image files with a mock OCR implementation
- provider-based LLM gateway with OpenAI, OpenAI-compatible, OpenRouter, and Gemini configuration
- `GET /api/v1/users`, thread-aware messages, and `POST /api/v1/ai/execute` SSE streaming
- local LLM config merge fixes and real provider validation
- `GET /api/v1/ai/config` for dynamic provider discovery
- `GET|PATCH /api/v1/me/settings` for persisted AI and appearance preferences
- parent thread metadata updates including `reply_count` and `last_reply_at`
- persisted message interaction APIs for:
  - reactions
  - delete
  - pin
  - save for later
  - unread checkpoints
  - AI feedback
- reasoning-aware AI SSE streaming
- websocket sync for message reactions, pin updates, and deletions
- backend-generated `ai_insight` for user profile hover cards
- `GET /api/v1/agent-collab/snapshot` so `#agent-collab` renders on first load
- `GET /api/v1/agent-collab/members` and `POST /api/v1/agent-collab/comm-log` for Windsurf's dynamic collaboration hub
- DM APIs for:
  - listing conversations
  - creating/opening a DM
  - loading DM history
  - sending DM messages
- DM response parity for docked chat UX:
  - `user_ids` on conversation payloads
  - DM realtime sync through websocket `message.created`
- DM realtime sync for `message.created` payloads that include `dm_id`
- Activity API for mentions, reactions, thread replies, and DM signals
- Later API for saved message retrieval
- Presence and typing APIs with websocket event fanout
- Draft persistence APIs for channel, DM, and thread composer state
- Starred channel and pinned message discovery APIs
- Persistent notification read state for inbox and mentions
- Persistent AI conversation history for assistant chat
- Persistent AI thread and channel summaries
- Persistent artifact and AI canvas lifecycle APIs
- Persistent artifact version history APIs
- Artifact diff APIs for version-to-version comparison
- Artifact restore APIs and structured diff spans for richer canvas history workflows
- Message-level artifact references and file attachment hydration
- Search coverage for artifacts and files, plus typed suggestions for both
- Artifact backlink discovery APIs
- Intelligent ranked search for cross-object knowledge retrieval
- Realtime `notifications.read` sync for multi-client inbox and mentions coordination
- Home API for a Slack-style landing surface with activity, drafts, DMs, tools, workflows, and starred channels
- Home knowledge aggregation fields for unread digest counts and recent cross-channel digests
- Hydrated user profile detail API for richer personal profile and directory views
- Persisted status update API for custom status and realtime presence refresh
- User groups directory and detail APIs for shared people collections
- Workflow and tool registry APIs for future workflow builder and tool surfaces
- Directory filtering across people by search, department, status, timezone, and user group
- Notification preferences and mute-rule APIs
- File archive lifecycle APIs
- File delete and richer file filtering APIs
- Workflow run history and manual run trigger APIs
- User profile editing and user group CRUD APIs
- Workflow run realtime events
- User group membership and mention lookup APIs
- Workflow run detail, cancel, and retry APIs
- Workflow run execution log APIs and run deletion lifecycle
- Tool execution history APIs and on-demand tool run execution
- Phase 33 frontend integration for lists, tool history, and template-first canvases
- Phase 34 contract-alignment aliases for structured lists, tool runs, and canvas bootstrap payloads
- Home aggregation for recent lists, tool runs, and files
- Activity and inbox feeds that include structured list completion, tool execution, and file upload signals
- UUID-style prefixed string primary keys for newly created channels, lists, tool runs, workflow runs, files, artifacts, user groups, AI conversations/messages, DM conversations/messages, workspace invites, and agents
- Channel-scoped pin retrieval for `GET /api/v1/pins?channel_id=...`
- Home dashboard compatibility fields:
  - `stats`
  - `recent_activity`
  - top-level `recent_artifacts`
- `DELETE /api/v1/drafts/:scope` for explicit draft cleanup flows
- File retention policy and file audit trail APIs
- File preview metadata API for images, PDFs, and fallback downloads
- File extraction detail APIs:
  - `GET /api/v1/files/:id/extraction`
  - `POST /api/v1/files/:id/extraction/rebuild`
  - `GET /api/v1/files/:id/extracted-content`
  - `GET /api/v1/files/:id/chunks`
  - `GET /api/v1/files/:id/citations`
  - `GET /api/v1/search/files?q=...`
- Unified citation lookup API:
  - `GET /api/v1/citations/lookup?q=...`
- AI composer APIs:
  - `POST /api/v1/ai/compose`
  - `POST /api/v1/ai/compose/stream`
  - `GET /api/v1/ai/compose/activity`
  - `GET /api/v1/ai/compose/activity/digest`
  - `POST /api/v1/ai/compose/:id/feedback`
  - `GET /api/v1/ai/compose/:id/feedback/summary`
  - Compose scope accepts exactly one of `channel_id` or `dm_id`; `thread_id` remains optional for channel thread context.
  - Compose intent accepts `reply`, `summarize`, `followup`, or `schedule`.
- Knowledge entity/wiki APIs:
  - `GET /api/v1/channels/:id/knowledge`
  - `GET /api/v1/channels/:id/knowledge/summary`
  - `GET /api/v1/channels/:id/knowledge/auto-summarize`
  - `PUT /api/v1/channels/:id/knowledge/auto-summarize`
  - `POST /api/v1/channels/:id/knowledge/auto-summarize`
  - `GET /api/v1/channels/:id/knowledge/digest`
  - `POST /api/v1/channels/:id/knowledge/digest/publish`
  - `GET /api/v1/channels/:id/knowledge/digest/schedule`
  - `PUT /api/v1/channels/:id/knowledge/digest/schedule`
  - `POST /api/v1/channels/:id/knowledge/digest/preview-schedule`
  - `DELETE /api/v1/channels/:id/knowledge/digest/schedule`
  - `GET /api/v1/knowledge/inbox`
  - `GET /api/v1/knowledge/inbox/:id`
  - `GET /api/v1/me/settings`
  - `GET /api/v1/workspace/settings`
  - `PATCH /api/v1/workspace/settings`
  - `PATCH /api/v1/users/me/knowledge/followed/bulk`
  - `GET /api/v1/users/me/knowledge/followed/stats`
  - `PATCH /api/v1/users/me/knowledge/followed/:id`
  - `GET /api/v1/users/me/knowledge/followed`
  - `GET /api/v1/knowledge/trending`
  - `GET /api/v1/knowledge/entities`
  - `GET /api/v1/knowledge/entities/suggest`
  - `POST /api/v1/knowledge/entities/match-text`
  - `POST /api/v1/knowledge/entities`
  - `GET /api/v1/knowledge/entities/:id`
  - `GET /api/v1/knowledge/entities/:id/activity`
  - `GET /api/v1/knowledge/entities/:id/activity/backfill-status`
  - `POST /api/v1/knowledge/entities/:id/activity/backfill`
  - `GET /api/v1/knowledge/entities/:id/hover`
  - `GET /api/v1/knowledge/entities/:id/brief`
  - `POST /api/v1/knowledge/entities/:id/brief`
  - `GET /api/v1/knowledge/entities/:id/brief/automation`
  - `POST /api/v1/knowledge/entities/:id/brief/automation/run`
  - `POST /api/v1/knowledge/entities/:id/brief/automation/retry`
  - `POST /api/v1/knowledge/entities/:id/ask`
  - `POST /api/v1/knowledge/entities/:id/ask/stream`
  - `GET /api/v1/knowledge/entities/:id/ask/history`
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
  - `GET /api/v1/knowledge/weekly-brief`
  - `POST /api/v1/knowledge/weekly-brief`
  - `GET /api/v1/knowledge/entities/:id/graph`
- Message metadata enrichment:
  - `message.metadata.entity_mentions`
  - `message.metadata.knowledge_digest`
- Entity-centric message discovery:
  - `GET /api/v1/search/messages/by-entity?entity_id=...`
- Knowledge realtime websocket events:
  - `knowledge.entity.created`
  - `knowledge.entity.updated`
  - `knowledge.entity.ref.created`
  - `knowledge.event.created`
  - `knowledge.link.created`
  - `knowledge.digest.published`
  - `knowledge.entity.activity.spiked`
  - `knowledge.trending.changed`
  - `knowledge.followed.stats.changed`
  - `knowledge.entity.brief.generated`
  - `knowledge.entity.brief.regen.queued`
  - `knowledge.entity.brief.regen.started`
  - `knowledge.entity.brief.regen.failed`
- AI schedule booking APIs:
  - `POST /api/v1/ai/schedule/book`
  - `GET /api/v1/ai/schedule/bookings`
  - `GET /api/v1/ai/schedule/bookings/:id`
  - `POST /api/v1/ai/schedule/bookings/:id/cancel`
- Channel notification preferences and self-service leave-channel API
- Structured workspace list APIs for shared checklists and operational tracking
- Artifact template APIs and virtual `new-doc` bootstrap support for canvas-first creation
- Artifact duplicate/fork API for copying an existing canvas into the same or another channel
- Custom status emoji and status expiration support
- Extended profile fields for pronouns, location, phone, and bio
- Workflow run step history, flat compatibility fields, and richer run detail hydration
- File response compatibility fields for `type`, `size`, `userId`, `channelId`, and `createdAt`
- File audit payload aliases for current governance UI consumption
- explicit AI command forwarding and stable new-canvas creation flow
- File asset upload and retrieval APIs for future attachment flows
- Presence heartbeat, status text, and last-seen metadata
- Workspace search API across channels, users, messages, and DM conversations
- Search suggestions and richer result snippets for command-palette style discovery
- Realtime file extraction updates through websocket `file.extraction.updated`
- Docked DM chat windows in the workspace shell
- Phase 10 Slack-parity foundation APIs for:
  - channel members
  - workspace invites
  - channel topic, purpose, and archive state
- Inbox and mentions APIs for notification-oriented workspace surfaces

See [CHANGELOG.md](./CHANGELOG.md) for shipped API details, and [docs/phase8-api-expansion.md](./docs/phase8-api-expansion.md) for the broader backend target.

Local verification note: in this environment, `pnpm build` still reaches `Creating an optimized production build ...` and does not exit. API verification and frontend lint continue to pass; production build investigation remains an active frontend follow-up.

## Tech Stack

- Next.js 16
- React 19
- TypeScript
- Tailwind CSS v4
- Zustand
- Tiptap
- Go
- Gin
- GORM
- SQLite

## Local Development

```bash
pnpm install
pnpm dev
```

Open `http://localhost:3000`.

To start the API locally:

```bash
make api-dev
```

API server runs on `http://localhost:8080`.

LLM configuration lives under `apps/api/config/`:

```txt
llm.base.yaml
llm.example.yaml
llm.local.yaml
llm.secrets.local.yaml
```

## Roadmap

- Expand realtime into presence, typing, thread deltas, and agent execution events
- Add Slack parity APIs for stars, pins surfaces, and richer composer state controls
- Evolve the LLM gateway into tool-calling and agent-runtime orchestration
- Add database-backed search, artifact lifecycle APIs, and execution history
- Strengthen product packaging across onboarding, docs, and public-facing positioning

## Repository Structure

```txt
apps/web/             Next.js app, UI components, hooks, stores, and mock flows
apps/api/             Go API, SQLite integration, REST handlers, and realtime hub
docs/                 Implementation, positioning, migration, and backend notes
```

## Repository Metadata

- Name: `relay-agent-workspace`
- Short product name: `Relay`
- Suggested GitHub description:
  - `AI-native collaboration workspace for humans and agents, combining messaging, artifacts, and orchestration.`
