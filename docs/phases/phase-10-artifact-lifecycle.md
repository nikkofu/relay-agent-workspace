# Phase 10: Artifact Lifecycle

## Goal

Turn the canvas surface into a real backend-backed collaboration layer by shipping artifact CRUD and AI-generated canvas creation.

## Scope

- `GET /api/v1/artifacts`
- `POST /api/v1/artifacts`
- `GET /api/v1/artifacts/:id`
- `PATCH /api/v1/artifacts/:id`
- `POST /api/v1/ai/canvas/generate`
- realtime `artifact.updated`

## Backend Tasks

- persist artifacts by channel
- support manual creation and update flows
- support AI-generated canvas content through the existing LLM gateway
- broadcast artifact updates over websocket
- harden activity/inbox/mentions item IDs so reaction events stay unique

## Frontend Handoff

- `CanvasPanel` can hydrate from `GET /api/v1/artifacts?channel_id=...`
- manual edit/save flows should use `POST /api/v1/artifacts` and `PATCH /api/v1/artifacts/:id`
- `/canvas` and AI canvas actions can call `POST /api/v1/ai/canvas/generate`
- websocket consumers can react to `artifact.updated`

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`
