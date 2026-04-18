# Phase 10: Slack Parity Foundation

## Goal

Bring Relay Agent Workspace closer to a complete Slack-style collaboration baseline while preserving the AI-native direction.

This phase starts with the collaboration primitives that unlock the next UI surfaces:

- channel members
- workspace invites
- channel topic, purpose, and archive state

## Why This Phase Comes First

These APIs are the lowest-risk parity layer after the current workspace became fully dynamic.

They unlock the most recognizable Slack flows:

- who is in a channel
- inviting people into the workspace
- editing channel identity and intent

They also create the data model base for later phases:

- inbox and mentions
- drafts
- presence and typing
- channel info panels and admin actions

## Scope

### Backend

- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`

### Data Model

- `channel_members`
- `workspace_invites`
- channel fields:
  - `topic`
  - `purpose`
  - `is_archived`

### Frontend Follow-up For Gemini

- channel info / details panel
- members list UI
- invite member flow
- topic and purpose editing surface

## Roles

### Codex

- overall architecture and API contract
- data model design
- Go handler and route implementation
- release notes and repository docs
- handoff material for Gemini

### Gemini

- Web UI for members and invites
- channel metadata editing UX
- sidebar and channel detail integration
- state wiring and interaction polish

## Execution Steps

1. Add missing domain models and seed data.
2. Write and pass handler tests for members, invites, and channel metadata.
3. Wire routes in `apps/api/main.go`.
4. Update roadmap and release documents.
5. Hand off the concrete request and response shapes to Gemini.

## Delivered In This Phase So Far

### Implemented

- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`

### Added To Persistence

- `channel_members`
- `workspace_invites`
- channel `topic`
- channel `purpose`
- channel `is_archived`

## Verification

- `cd apps/api && go test ./internal/handlers`
- later release gate:
  - `cd apps/api && go test ./...`
  - `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
  - `pnpm build`

## Next Recommended Phase Order

1. `inbox + mentions`
2. `drafts`
3. `presence + typing`
4. richer channel admin surfaces
5. AI conversation persistence
