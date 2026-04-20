# Phase 10: Artifact Diff API

## Goal

Add a version-to-version diff API so Relay artifacts can support visual comparison views instead of only browsing snapshots one at a time.

## Scope

This phase adds:

- `GET /api/v1/artifacts/:id/diff/:from/:to`
- unified text diff output
- line-count summary for added and removed lines
- raw `from_content` and `to_content` in the same payload for flexible UI rendering

This phase does not add:

- structured token-level diff spans
- three-way merge
- side-by-side rendering in the backend

## API Contract

Example response:

```json
{
  "diff": {
    "artifact_id": "artifact-1",
    "from_version": 1,
    "to_version": 2,
    "from_content": "Initial outline",
    "to_content": "Revised outline",
    "unified_diff": "--- v1\n+++ v2\n@@\n-Initial outline\n+Revised outline",
    "summary": {
      "added_lines": 1,
      "removed_lines": 1
    }
  }
}
```

## Gemini Handoff

Gemini can now build a compare view from one request:

- `GET /api/v1/artifacts/:id/diff/:from/:to`

Recommended UI order:

1. choose two versions from the history list
2. request the diff payload
3. render either:
   - unified diff view from `unified_diff`
   - side-by-side comparison from `from_content` and `to_content`
4. show change stats from `summary`

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`
