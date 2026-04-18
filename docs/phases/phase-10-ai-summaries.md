# Phase 10: AI Summaries

## Goal

Ship persistent AI-generated summaries for threads and channels so Relay can surface fast context recovery instead of forcing users to read every message manually.

## Scope

- `GET /api/v1/messages/:id/summary`
- `POST /api/v1/messages/:id/summary`
- `GET /api/v1/channels/:id/summary`
- `POST /api/v1/channels/:id/summary`
- persisted `ai_summaries` storage

## Backend Tasks

- add a shared summary persistence model keyed by `scope_type + scope_id`
- summarize a thread from its parent message plus replies
- summarize a channel from its recent messages
- persist provider, model, reasoning, message count, and last message timestamp

## Frontend Handoff

- thread panel should replace the current mock summary card with `GET/POST /api/v1/messages/:id/summary`
- channel summary can be added as a second surface without changing the backend contract
- POST accepts optional `provider` and `model`
- GET returns `{ "summary": null }` when no summary exists yet

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`
