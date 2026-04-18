# Changelog

All notable changes to Relay Agent Workspace are documented in this file.

## [0.5.8] - 2026-04-18

This release implements Phase 13: Presence and Typing indicators, bringing real-time social signals to the workspace.

### Added

- **Live Presence**: Real-time user status (Online, Away, Busy, Offline) across the entire UI.
- **Typing Indicators**: Visual feedback when teammates are typing in Channels, DMs, or Threads.
- **Presence Store**: New `presence-store.ts` for managing transient real-time states.
- **Websocket Expansion**: Integrated `presence.updated` and `typing.updated` events into the unified websocket hook.

### Fixed

- UI consistency: Status dots now appear correctly on all user avatars and navigation elements.
- Typing debounce: Efficiently manages typing broadcast state to minimize network overhead.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.6] - 2026-04-18


This release adds the first realtime presence layer for Relay by shipping presence APIs and typing broadcasts for channels, DMs, and threads.

### Added

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`

### Realtime

- `POST /api/v1/presence` broadcasts `presence.updated`
- `POST /api/v1/typing` broadcasts `typing.updated`

### Documentation

- added [docs/phases/phase-10-presence-typing.md](./docs/phases/phase-10-presence-typing.md)
- added [docs/releases/v0.5.7.md](./docs/releases/v0.5.7.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for presence and typing
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped realtime baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.6] - 2026-04-18

This release implements the Phase 12 Drafts integration, bringing autosave persistence to all composer surfaces.

### Added

- **Drafts Persistence**: Integrated `GET /api/v1/drafts` and `PUT /api/v1/drafts/:scope` for reliable message autosaving.
- **Rich Thread Replies**: Replaced the basic thread textarea with the rich `MessageComposer`, enabling bold, italic, and slash commands in threads.
- **Unified Draft Store**: Added `draft-store.ts` to manage draft state across Channels, DMs, and Threads.
- **Autosave & Restore**: Content is automatically saved as you type and restored when you return to a conversation.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.5] - 2026-04-18


This release starts the next Slack-parity backend wave by adding draft persistence APIs for channel, DM, and thread composers.

### Added

- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`

### Behavior

- drafts are returned for the current user only
- drafts are ordered by `updated_at desc`
- one draft is stored per `user_id + scope`
- recommended scope keys:
  - `channel:<channelId>`
  - `dm:<dmId>`
  - `thread:<messageId>`

### Persistence

- added `drafts`
- seeded one example draft for local development visibility

### Documentation

- added [docs/phases/phase-10-drafts.md](./docs/phases/phase-10-drafts.md)
- added [docs/releases/v0.5.5.md](./docs/releases/v0.5.5.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for drafts
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped drafts baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.4] - 2026-04-18

This release implements the Phase 11 notification UI wave, introducing a unified Activity, Inbox, and Mentions experience.

### Added

- **Inbox & Mentions UI**: Integrated new notification tabs in the Activity page.
- **Unified Activity Store**: Expanded `activity-store.ts` to support fetching Inbox and Mentions from backend endpoints.
- **Smart Navigation**: Activity items now support clicking to jump directly to the relevant channel or DM conversation.
- **Interactive Tabs**: Seamless switching between "All Activity", "Inbox", and "Mentions" views.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.3] - 2026-04-18


This release unifies Gemini's `v0.5.2` channel-management frontend work with the next Codex backend wave by adding inbox and mentions APIs.

### Added

- `GET /api/v1/inbox`
- `GET /api/v1/mentions`

### Behavior

- inbox returns aggregated collaboration signals for the current user:
  - direct mentions
  - thread replies
  - reactions on your messages
  - DM activity
- mentions returns the direct-mention subset only

### Documentation

- added [docs/phases/phase-10-inbox-mentions.md](./docs/phases/phase-10-inbox-mentions.md)
- added [docs/releases/v0.5.3.md](./docs/releases/v0.5.3.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for inbox and mentions
- updated `docs/phase8-api-expansion.md` shipped baseline to include the new endpoints

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.2] - 2026-04-18

This release completes the Phase 10 frontend integration for channel management, introducing member lists, metadata editing, and a unified channel info panel.

### Added

- `ChannelInfo` panel for managing channel metadata and membership
- support for editing channel `topic` and `purpose`
- member management UI (list, add, and remove members)
- extended `Channel` and `ChannelMember` types for robust metadata handling
- integrated `ChannelInfo` into the main message area header

### Fixed

- channel metadata sync: ensures `topic` and `purpose` are correctly updated across the UI
- member list hydration: automatically fetches and updates members when switching channels

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.1] - 2026-04-18

This release starts Phase 10 with the first Slack-parity backend wave and adds explicit stage documentation for ongoing Codex + Gemini collaboration.

### Added

- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`

### Data Model

- added `channel_members`
- added `workspace_invites`
- extended `channels` with:
  - `topic`
  - `purpose`
  - `is_archived`

### Documentation

- added [docs/phases/phase-10-slack-parity-foundation.md](./docs/phases/phase-10-slack-parity-foundation.md)
- added [docs/releases/v0.5.1.md](./docs/releases/v0.5.1.md)
- updated `docs/AGENT-COLLAB.md` with Gemini handoff for members, invites, and metadata editing
- updated `docs/phase8-api-expansion.md` shipped baseline to include the new endpoints

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.0] - 2026-04-18

This release synchronizes Gemini's docked DM overhaul with the backend contract, confirms the DM payload shape in handler tests, and documents the next Slack-parity API wave for Relay.

### Added

- DM conversation responses now formally include `user_ids`:
  - `GET /api/v1/dms`
  - `POST /api/v1/dms`
- handler coverage now verifies:
  - DM list payloads include `user_ids`
  - DM create/open payloads include `user_ids`

### Frontend Integration Sync

- `/workspace/dms` now redirects back into the unified workspace shell
- docked DM chat windows are mounted from the workspace layout
- DM store now maps `user_ids` and defensive `last_message_at` fallbacks
- message-store duplicate guards reduce repeated DM and thread inserts
- IME-safe Enter handling was added to the rich message composer

### Planning And Documentation

- updated `docs/phase8-api-expansion.md` with the next Slack-parity API layer:
  - channel members
  - invites
  - channel topic and purpose
  - stars and pins
  - inbox and mentions
  - drafts
- updated `docs/AGENT-COLLAB.md` with the `v0.5.0` handoff and next-phase objectives
- updated `README.md` to reflect the docked DM experience and the new planning track

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.4.1] - 2026-04-18

This release closes the DM realtime gap and expands the post-Phase-8 backend with Activity, Later, and Search APIs so Gemini can finish the remaining workspace pages against live data.

### Added

- `GET /api/v1/activity`
  - returns recent collaboration signals for the current user
  - includes mentions, thread replies, reactions on your messages, and DM activity
- `GET /api/v1/later`
  - returns the current user's saved messages with channel and sender context
- `GET /api/v1/search?q=...`
  - returns grouped results across:
    - `channels`
    - `users`
    - `messages`
    - `dms`

### Realtime And DM Improvements

- `POST /api/v1/dms/:id/messages` now broadcasts `message.created`
- DM websocket payloads now include:
  - `id`
  - `dm_id`
  - `user_id`
  - `content`
  - `created_at`
- `POST /api/v1/dms` now accepts either:
  - `user_id`
  - `user_ids`
  which keeps the backend compatible with the current frontend store shape

### Documentation Sync

- updated `README.md` current status to the `v0.4.1` release line
- updated `docs/phase8-api-expansion.md` so DM, Activity, Later, and Search are no longer listed as missing
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for page integration

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.9] - 2026-04-18

This release starts the first post-Phase-8 backend expansion by adding the minimum DM API surface needed for Gemini to replace the static DMs page.

### Added

- `GET /api/v1/dms`
  - returns recent DM conversations for the current user
  - includes the counterpart user plus last message preview
- `POST /api/v1/dms`
  - creates or reopens a 1:1 DM conversation
  - payload: `{ "user_id": string }`
- `GET /api/v1/dms/:id/messages`
  - returns ordered DM history
- `POST /api/v1/dms/:id/messages`
  - sends a DM message
  - payload:
    - `content`
    - `user_id`

### Data Model

- added:
  - `dm_conversations`
  - `dm_members`
  - `dm_messages`
- seeded one initial DM thread between `user-1` and `user-2`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.7] - 2026-04-18

This release adds a real backend collaboration insight path for user profiles and fixes `#agent-collab` so it renders meaningful content on first load.

### Added

- `GET /api/v1/agent-collab/snapshot`
  - returns:
    - `active_superpowers`
    - `task_board`
- dynamic `ai_insight` generation on existing user payloads:
  - `GET /api/v1/me`
  - `GET /api/v1/users`

### Fixed

- `#agent-collab` no longer depends on catching a startup websocket event
- frontend collab store now correctly maps backend lowercase JSON keys
- the dashboard now renders both:
  - agent cards
  - task board rows

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.4] - 2026-04-18

This release closes the audit loop on the original LLM/thread/user API delivery plan and synchronizes repository-facing docs with the shipped backend state.

### Audit Result

- `docs/superpowers/plans/2026-04-17-llm-thread-user-api.md` has been audited against the live codebase
- conclusion: the original plan scope is complete
- the plan file now reflects shipped status instead of stale unchecked tasks

### Documentation Sync

- updated `README.md` to point at the latest release line
- updated `docs/AGENT-COLLAB.md` with the audit result and Gemini handoff note

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.3.3] - 2026-04-18

This release aligns the backend with Gemini's latest realtime AI integration pass, adds reasoning-aware SSE support, and refreshes the repository docs for GitHub-facing consumption.

### Added

- `POST /api/v1/ai/execute` may now emit `event: reasoning` in addition to:
  - `start`
  - `chunk`
  - `done`
  - `error`

### Realtime Coverage

- reaction mutations now broadcast `reaction.updated`
- pin toggles now broadcast `message.updated`
- deletions now broadcast `message.deleted`

### Documentation Refresh

- updated `README.md` with current product positioning and shipped backend surface
- replaced broken local-only links with GitHub-safe relative links
- rewrote `docs/phase8-api-expansion.md` to reflect the current monorepo and broader backend target
- updated `docs/AGENT-COLLAB.md` with Gemini handoff notes for reasoning and websocket coverage

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.3.1] - 2026-04-18

This release completes the first persistent message interaction APIs for Gemini's channel message actions UI.

### Added

- `POST /api/v1/ai/feedback`
  - Payload: `{ "message_id": string, "is_good": boolean }`
  - Response: `{ "feedback": { ... } }`
- `POST /api/v1/messages/:id/reactions`
  - Payload: `{ "emoji": string }`
  - Toggle behavior for the current user
  - Response:
    - `message`
    - `added`
- `DELETE /api/v1/messages/:id`
  - Deletes the target message
  - Deletes child replies when the target is a thread parent
  - Response: `{ "deleted": true, "message_id": string }`
- `POST /api/v1/messages/:id/pin`
  - Toggle message pin state
  - Response:
    - `message`
    - `is_pinned`
- `POST /api/v1/messages/:id/later`
  - Toggle save-for-later state for the current user
  - Response:
    - `message_id`
    - `saved`
- `POST /api/v1/messages/:id/unread`
  - Marks the message as an unread checkpoint for the current user
  - Response:
    - `message_id`
    - `unread`

### Data Model And Behavior

- `Message` now includes:
  - `is_pinned`
- New persistence tables:
  - `message_reactions`
  - `saved_messages`
  - `unread_markers`
  - `ai_feedback`
- `GET /api/v1/messages` and `GET /api/v1/messages/:id/thread` now rebuild `metadata.reactions` from persisted reaction rows
- deleting a thread reply now recomputes the parent message:
  - `reply_count`
  - `last_reply_at`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.2.8] - 2026-04-17

This release removes the last hardcoded AI settings path for the frontend and adds persistence for user AI preferences plus stronger thread reply metadata.

### Added

- `GET /api/v1/ai/config`
  - Returns enabled providers and configured models
  - Response shape:
    - `default_provider`
    - `providers[]`
- `PATCH /api/v1/me/settings`
  - Persists AI preference fields on the current user:
    - `provider`
    - `model`
    - `mode`

### Thread Integrity Improvements

- reply creation now updates:
  - `reply_count`
  - `last_reply_at`
- `Message` model now stores:
  - `last_reply_at`
- `User` model now stores:
  - `ai_provider`
  - `ai_model`
  - `ai_mode`

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.2.5] - 2026-04-17

This release hardens local LLM configuration loading and records the first real upstream validation pass against the configured providers.

### Fixed

- layered LLM config merge now preserves provider defaults while applying local overrides
- local YAML parsing now tolerates tab characters in `llm.local.yaml`

### Verified

- local `gemini` provider configuration successfully streamed real SSE output through `POST /api/v1/ai/execute`
- local `openrouter` configuration loaded correctly, but the selected free model returned upstream `429` rate limiting during validation

### Notes

- with the current local config, Gemini is the verified working provider for frontend integration
- OpenRouter can still be used after switching to a non-rate-limited model or retrying later

## [0.2.4] - 2026-04-17

This release adds the first real LLM gateway architecture plus the backend APIs Gemini requested for user resolution, message threading, and AI streaming.

### Added

- Provider-based LLM gateway in `apps/api/internal/llm`
- Config loading in `apps/api/internal/config`
- Config files:
  - `apps/api/config/llm.base.yaml`
  - `apps/api/config/llm.example.yaml`
  - `apps/api/config/llm.local.yaml`
  - `apps/api/config/llm.secrets.local.yaml`
- Supported provider kinds:
  - `openai`
  - `openai-compatible`
  - `openrouter`
  - `gemini`

### API Surface Added In This Release

- `GET /api/v1/users`
  - Returns all users
  - Supports optional query param `id`
  - Response: `{ "users": [...] }`

- `GET /api/v1/messages/:id/thread`
  - Returns parent message plus replies
  - Response:
    - `parent`
    - `replies`

- `POST /api/v1/ai/execute`
  - Accepts:
    - `prompt`
    - `channel_id`
    - optional `provider`
    - optional `model`
  - Returns `text/event-stream`
  - Stream events currently emitted:
    - `start`
    - `chunk`
    - `done`
    - `error`

### Model And Handler Changes

- `Message` now includes:
  - `thread_id`
  - `reply_count`
- `GET /api/v1/messages` now returns top-level channel messages only
- `POST /api/v1/messages` accepts optional `thread_id`
- Reply creation increments parent `reply_count`

### LLM Notes

- OpenAI and OpenAI-compatible providers use configurable `api_style`
  - `responses`
  - `chat_completions`
- OpenRouter defaults to `responses`
- Gemini uses the official native Gemini streaming protocol over `streamGenerateContent?alt=sse`
- Env overrides are supported through:
  - `LLM_DEFAULT_PROVIDER`
  - `LLM_PROVIDER_<NAME>_API_KEY`
  - `LLM_PROVIDER_<NAME>_BASE_URL`
  - `LLM_PROVIDER_<NAME>_MODEL`
  - `LLM_PROVIDER_<NAME>_API_STYLE`
  - `LLM_PROVIDER_<NAME>_ENABLED`

### Verification Used For This Release

- `pnpm build`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.2.2] - 2026-04-16

This release adds the backend sync path for the `#agent-collab` workspace view and aligns the repository version with the latest cross-agent handoff.

### Added

- `docs/AGENT-COLLAB.md` file watcher in `apps/api`
- Markdown table parser for:
  - `Task Board`
  - `Active Superpowers`
- Realtime broadcast for `agent_collab.sync`

### Realtime Event Added

- `agent_collab.sync`
  - Broadcast through `GET /api/v1/realtime`
  - Fixed target channel id: `ch-collab`
  - Payload shape:
    - `active_superpowers`
    - `task_board`

Example payload:

```json
{
  "type": "agent_collab.sync",
  "channel_id": "ch-collab",
  "payload": {
    "active_superpowers": [],
    "task_board": []
  }
}
```

### Backend Notes

- Watch path defaults to `../../docs/AGENT-COLLAB.md` from `apps/api`
- Sync is pushed on service startup and on subsequent file `write/create` events
- Parsing currently targets the two collaboration tables only

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- Automated coverage added for:
  - Markdown table parsing
  - Direct sync broadcast
  - Watcher-triggered broadcast after file write

## [0.2.0] - 2026-04-16

This release turns Relay Agent Workspace into a monorepo with a working backend foundation. It is the first version suitable for frontend-to-backend integration across `apps/web` and `apps/api`.

### Added

- Monorepo structure with `apps/web` and `apps/api`
- Go API service built with Gin
- SQLite persistence using GORM auto-migrations
- Seed data for:
  - `Organization`
  - `Team`
  - `User`
  - `Agent`
  - `Workspace`
  - `Channel`
  - `Message`
- In-memory realtime WebSocket hub
- REST handlers for collaboration and organization workflows

### API Surface In This Release

Base URL:
- Web: `http://localhost:3000`
- API: `http://localhost:8080`

Health:
- `GET /ping`
  - Returns `{"message":"pong"}`

Current user:
- `GET /api/v1/me`
  - Returns the seeded current user
  - Response shape:
    - `user.id`
    - `user.org_id`
    - `user.name`
    - `user.email`
    - `user.avatar`
    - `user.status`

Organizations:
- `GET /api/v1/orgs`
  - Lists organizations visible to the current seeded user
  - Response: `{ "organizations": [...] }`

Teams:
- `GET /api/v1/orgs/:id/teams`
  - Lists teams under an organization
  - Example: `/api/v1/orgs/org_1/teams`
  - Response: `{ "teams": [...] }`

Agents:
- `POST /api/v1/orgs/:id/agents`
  - Creates an agent under an organization
  - Required JSON body:
    - `name`
    - `type`
    - `owner_id`
  - Response: `{ "agent": { ... } }`

Workspaces:
- `GET /api/v1/workspaces`
  - Lists workspaces from SQLite
  - Response: `{ "workspaces": [...] }`

Channels:
- `GET /api/v1/channels`
  - Supports query param:
    - `workspace_id`
  - Example: `/api/v1/channels?workspace_id=ws_1`
  - Response: `{ "channels": [...] }`

Messages:
- `GET /api/v1/messages`
  - Required query param:
    - `channel_id`
  - Example: `/api/v1/messages?channel_id=ch_1`
  - Returns channel messages ordered by `created_at asc`
  - Response: `{ "messages": [...] }`

- `POST /api/v1/messages`
  - Creates a new message and persists it to SQLite
  - Required JSON body:
    - `channel_id`
    - `content`
    - `user_id`
  - Response: `{ "message": { ... } }`
  - Side effect:
    - broadcasts a realtime `message.created` event to websocket clients

Realtime:
- `GET /api/v1/realtime`
  - WebSocket upgrade endpoint
  - Current shipped events:
    - `realtime.connected`
    - `message.created`
  - `message.created` event envelope includes:
    - `id`
    - `type`
    - `workspace_id`
    - `channel_id`
    - `entity_id`
    - `ts`
    - `payload`

### Notes For Frontend Integration

- Persistence uses `apps/api/db/relay.db`
- Database auto-migration runs on API startup
- Seed data is idempotent and backfills newly added records
- Realtime is currently in-memory and single-instance only
- Message pagination, presence, typing, AI execute, and SSE are not part of `0.2.0`

### Verification Used For This Release

- `pnpm build`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- Manual local verification of:
  - `GET /ping`
  - `GET /api/v1/orgs`
  - `GET /api/v1/orgs/org_1/teams`
  - `POST /api/v1/orgs/org_1/agents`
  - `GET /api/v1/workspaces`
  - `GET /api/v1/channels?workspace_id=ws_1`
  - `GET /api/v1/messages?channel_id=ch_1`
  - `POST /api/v1/messages`
  - `GET /api/v1/realtime` websocket upgrade and event receipt

### Internal Commits Included

- `e8e3bba` refactor: migrate to monorepo structure with Go backend boilerplate (Phase 8.1)
- `f628d06` feat: implement phase 8 core collaboration api
- `32202a8` chore: move api sqlite database into db directory
- `c752f9d` feat: add organization team and agent api endpoints
- `a1e8699` feat: add realtime websocket hub
