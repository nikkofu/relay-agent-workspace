# Phase 10: Artifact Version History

## Goal

Move Relay artifacts from simple editable records to auditable collaborative assets with stored version snapshots.

## Scope

This phase adds:

- persisted `artifact_versions`
- automatic snapshot creation on:
  - `POST /api/v1/artifacts`
  - `PATCH /api/v1/artifacts/:id`
  - `POST /api/v1/ai/canvas/generate`
- `GET /api/v1/artifacts/:id/versions`
- `GET /api/v1/artifacts/:id/versions/:version`

This phase does not add:

- rollback or restore
- version diff rendering
- collaborative cursors inside the canvas

## API Contract

### `GET /api/v1/artifacts/:id/versions`

Returns:

```json
{
  "versions": [
    {
      "artifact_id": "artifact-1",
      "version": 2,
      "title": "Launch Notes",
      "type": "document",
      "status": "live",
      "content": "Revised outline",
      "updated_by": "user-1",
      "updated_by_user": {
        "id": "user-1",
        "name": "Nikko Fu"
      },
      "created_at": "2026-04-20T02:02:13Z"
    }
  ]
}
```

### `GET /api/v1/artifacts/:id/versions/:version`

Returns:

```json
{
  "version": {
    "artifact_id": "artifact-1",
    "version": 1,
    "content": "Initial outline"
  }
}
```

## Data Model

- `artifacts.version` stores the latest version number
- `artifact_versions` stores immutable historical snapshots keyed by:
  - `artifact_id`
  - `version`

## Gemini Handoff

Gemini can now replace the placeholder `History` affordance in `CanvasPanel` with a real history sheet or drawer using:

- `GET /api/v1/artifacts/:id/versions`
- `GET /api/v1/artifacts/:id/versions/:version`

Recommended UI order:

1. show version list newest first
2. show editor user and timestamp for each entry
3. fetch detail only when a version row is opened

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`
