# Phase 38 - Artifact Duplicate/Fork APIs

Date: 2026-04-21
Owner: Codex
Status: Backend complete, frontend integration queued

## Goal

Close the canvas workflow gap where users can create artifacts from templates and AI generation, but cannot copy an existing canvas into a new working draft.

## Delivered API

### `POST /api/v1/artifacts/:id/duplicate`

Creates a new artifact from an existing artifact.

Request body:

```json
{
  "channel_id": "optional target channel id",
  "title": "optional copied artifact title"
}
```

Response body:

```json
{
  "artifact": {
    "id": "artifact_<uuid>",
    "channel_id": "ch_...",
    "title": "Launch Notes Fork",
    "version": 1,
    "source": "duplicate",
    "status": "draft"
  }
}
```

Behavior:

- Creates a new prefixed UUID artifact ID.
- Preserves source `type`, `content`, `template_id`, `provider`, and `model`.
- Uses the requested `channel_id`, or the source artifact channel when omitted.
- Uses the requested `title`, or creates a copy title when omitted.
- Resets `status` to `draft`, `version` to `1`, and `source` to `duplicate`.
- Writes an initial `ArtifactVersion` snapshot.
- Broadcasts `artifact.updated` for realtime canvas refresh.

## Gemini Handoff

Frontend integration should add Duplicate/Fork entry points to:

- Canvas panel action menu.
- Artifact history/detail sidebar.
- Artifact cards or template-first canvas menus where it fits the UX.

On success, consume the returned `artifact` directly and open/select it immediately.

## Verification

- `cd apps/api && go test ./internal/handlers -run 'TestArtifactCRUDAndAI_generate$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
