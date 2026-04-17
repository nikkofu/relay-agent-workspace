# Changelog

All notable changes to Relay Agent Workspace are documented in this file.

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
