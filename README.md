# Relay Agent Workspace

AI-native collaboration workspace for humans and agents, combining messaging, artifacts, and orchestration.

## Overview

Relay Agent Workspace is an AI-native collaboration workspace where people and agents can chat, think, coordinate, and execute together. The repository now includes both the Next.js app and an initial Go API for organizations, workspaces, channels, messages, and realtime events.

It started as an internal Slack-inspired UI exploration and is now being repositioned as `Relay`: a broader collaboration product centered on realtime messaging plus agent-native workflows.

## Core Capabilities

- Slack-style workspace, channel, DM, activity, and later views
- Rich message composer with slash commands, mentions, and formatting
- AI assistant panel with simulated streaming chat
- Thread panel and AI thread summary surface
- Canvas / artifact panel for AI-generated collaborative outputs
- Search dialog and workspace navigation patterns for future dynamic data integration

## Why Relay

Traditional team chat tools treat AI as a sidebar feature. Relay treats agents as first-class collaborators inside the workspace:

- messaging is the shared coordination layer
- AI chat is part of the conversation flow
- artifacts live alongside messages
- agent execution can become a visible, auditable workspace primitive

## Current Status

`v0.2.5` is the latest backend-enabled release line and includes:

- Go + Gin API service under `apps/api`
- SQLite persistence via GORM
- seed data for org, team, user, agent, workspace, channel, and message
- realtime websocket endpoint for workspace event fanout
- REST endpoints for org, team, agent, workspace, channel, and message flows
- `AGENT-COLLAB.md` watcher and `agent_collab.sync` websocket broadcast for the `#agent-collab` dashboard
- provider-based LLM gateway with OpenAI, OpenAI-compatible, OpenRouter, and Gemini configuration
- `GET /api/v1/users`, thread-aware messages, and `POST /api/v1/ai/execute` SSE streaming
- local LLM config merge fixes and real provider validation

See [CHANGELOG.md](/Users/admin/Documents/WORK/ai/relay-agent-workspace/CHANGELOG.md:1) for the detailed API inventory in this release, and [docs/phase8-api-expansion.md](/Users/admin/Documents/WORK/ai/relay-agent-workspace/docs/phase8-api-expansion.md:1) for the broader backend target.

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

- Connect the UI to the shipped org/workspace/message APIs
- Expand realtime beyond `message.created` into presence, typing, and thread events
- Evolve the shipped LLM gateway into tool-calling and agent-runtime orchestration
- Add database-backed search, artifacts, and agent execution history
- Keep strengthening the Relay brand across docs, metadata, and product surfaces

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
