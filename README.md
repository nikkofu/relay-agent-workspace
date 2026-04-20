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

`v0.5.35` is the current release line and includes:

- Go + Gin API service under `apps/api`
- SQLite persistence via GORM
- seed data for org, team, user, agent, workspace, channel, and message
- realtime websocket endpoint for workspace event fanout
- REST endpoints for org, team, agent, workspace, channel, and message flows
- `AGENT-COLLAB.md` watcher and `agent_collab.sync` websocket broadcast for the `#agent-collab` dashboard
- provider-based LLM gateway with OpenAI, OpenAI-compatible, OpenRouter, and Gemini configuration
- `GET /api/v1/users`, thread-aware messages, and `POST /api/v1/ai/execute` SSE streaming
- local LLM config merge fixes and real provider validation
- `GET /api/v1/ai/config` for dynamic provider discovery
- `PATCH /api/v1/me/settings` for persisted AI preferences
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
- explicit AI command forwarding and stable new-canvas creation flow
- File asset upload and retrieval APIs for future attachment flows
- Presence heartbeat, status text, and last-seen metadata
- Workspace search API across channels, users, messages, and DM conversations
- Search suggestions and richer result snippets for command-palette style discovery
- Docked DM chat windows in the workspace shell
- Phase 10 Slack-parity foundation APIs for:
  - channel members
  - workspace invites
  - channel topic, purpose, and archive state
- Inbox and mentions APIs for notification-oriented workspace surfaces

See [CHANGELOG.md](./CHANGELOG.md) for shipped API details, and [docs/phase8-api-expansion.md](./docs/phase8-api-expansion.md) for the broader backend target.

## Tech Stack

- Next.js 15
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
