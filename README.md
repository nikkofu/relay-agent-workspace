# Relay Agent Workspace

AI-native collaboration workspace for humans and agents, combining messaging, artifacts, and orchestration.

## Overview

Relay Agent Workspace is a TypeScript-first UI foundation for a shared workspace where people and agents can chat, think, coordinate, and execute together. The current repository focuses on the frontend experience: workspace navigation, channel messaging, threads, AI chat, search, and canvas-style artifact surfaces.

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

This repository is currently a frontend-first implementation with mock data and simulated AI flows. The next major step is backend integration:

- workspace, channel, message, and thread APIs
- realtime messaging and presence
- external LLM integration through a unified AI gateway
- database-backed state and search
- agent runtime and artifact persistence

See [docs/phase8-api-expansion.md](/Users/admin/Documents/WORK/ai/acim-ui/docs/phase8-api-expansion.md:1) for the current backend API expansion analysis.

## Tech Stack

- Next.js 15
- React 19
- TypeScript
- Tailwind CSS v4
- shadcn/ui
- Zustand
- Tiptap

## Local Development

```bash
pnpm install
pnpm dev
```

Open `http://localhost:3000`.

## Roadmap

- Connect the UI to real workspace and messaging APIs
- Add realtime transport for messages, threads, presence, and AI streaming
- Introduce an AI gateway for external LLM providers
- Add database-backed search, artifacts, and agent execution history
- Migrate the repository to `relay-agent-workspace`

## Repository Structure

```txt
app/                  Next.js app router pages and layouts
components/           UI, layout, message, AI, search, and canvas components
hooks/                Client-side hooks, including simulated AI chat
stores/               Zustand state stores
lib/                  Mock data and utilities
types/                Shared TypeScript types
docs/                 Implementation, positioning, and migration documents
```

## Repository Metadata

- Name: `relay-agent-workspace`
- Short product name: `Relay`
- Suggested GitHub description:
  - `AI-native collaboration workspace for humans and agents, combining messaging, artifacts, and orchestration.`

