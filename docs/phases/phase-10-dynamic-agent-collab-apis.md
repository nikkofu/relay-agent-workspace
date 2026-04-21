# Phase 40 - Dynamic Agent-Collab APIs

Date: 2026-04-21
Owner: Codex
Frontend Partner: Windsurf
Status: Backend complete, frontend integration queued

## Goal

Make Windsurf's #agent-collab hub dynamic without introducing a second source of truth. `docs/AGENT-COLLAB.md` remains the collaboration record, while the API serves parsed JSON and writes new communication entries back to the Markdown file.

## Delivered APIs

### `GET /api/v1/agent-collab/members`

Returns member profiles parsed from the `Member Profiles` table.

### `POST /api/v1/agent-collab/comm-log`

Accepts:

```json
{
  "from": "Codex",
  "to": "Windsurf",
  "title": "Phase 40 Dynamic Agent-Collab API",
  "content": "The dynamic hub API is ready for frontend integration."
}
```

The handler inserts a new Markdown communication section, then broadcasts `agent_collab.sync`.

## Snapshot Changes

`GET /api/v1/agent-collab/snapshot` and websocket `agent_collab.sync` now include:

- `active_superpowers`
- `comm_log`
- `members`
- `task_board`

## Windsurf Tasks

- Replace static `MEMBERS` with `snapshot.members` or `GET /api/v1/agent-collab/members`.
- Replace static `COMM_SECTIONS` with `snapshot.comm_log`.
- Listen for websocket `agent_collab.sync` and refresh the hub state when `members`, `comm_log`, `task_board`, or `active_superpowers` changes.
- Keep static data as an offline fallback only.

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/agentcollab`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestAgentCollabMembersAndCommLogEndpoints$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
