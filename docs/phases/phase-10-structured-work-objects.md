# Phase 10: Structured Work Objects

This phase expands Relay beyond chat and automation logs into lightweight operational objects that teams can share inside the workspace:

- structured lists for launch checklists, tracking, and handoff work
- executable tool runs with history and detail payloads
- artifact templates and virtual `new-doc` bootstrap flows for canvas-first creation

## Added Backend APIs

- `GET /api/v1/lists`
- `POST /api/v1/lists`
- `GET /api/v1/lists/:id`
- `PATCH /api/v1/lists/:id`
- `DELETE /api/v1/lists/:id`
- `POST /api/v1/lists/:id/items`
- `PATCH /api/v1/lists/:id/items/:itemId`
- `DELETE /api/v1/lists/:id/items/:itemId`
- `GET /api/v1/tools/runs`
- `GET /api/v1/tools/runs/:id`
- `POST /api/v1/tools/:id/execute`
- `GET /api/v1/artifacts/templates`
- `POST /api/v1/artifacts/from-template`
- virtual `GET /api/v1/artifacts/new-doc`

## Why This Phase Matters

Relay already had messages, files, artifacts, workflows, and AI conversations. What was still missing was the layer between a message and a full automation:

- a shared checklist object for operational work
- a visible execution history for user-triggered tools
- a faster path into structured canvases than blank manual creation

Without these objects, Home, channel sidebars, and future project surfaces cannot become true operational workspaces. With them, Gemini can build list panels, tool-run inspectors, and template pickers using stable contracts.

## Backend Notes

- new persistence:
  - `workspace_lists`
  - `workspace_list_items`
  - `tool_runs`
  - `tool_run_logs`
- `artifacts` and `artifact_versions` now persist `template_id`
- seed data now includes:
  - a launch checklist list
  - a sample tool run with logs
  - template-based artifact creation support

## Gemini Handoff

Gemini can now integrate the following UI surfaces:

1. Home or channel checklist widgets using:
   - `GET /api/v1/lists?channel_id=...`
   - `GET /api/v1/lists/:id`
2. List editing UX using:
   - `POST /api/v1/lists`
   - `PATCH /api/v1/lists/:id`
   - `POST /api/v1/lists/:id/items`
   - `PATCH /api/v1/lists/:id/items/:itemId`
   - `DELETE /api/v1/lists/:id/items/:itemId`
3. Tool execution panels using:
   - `GET /api/v1/tools/runs`
   - `GET /api/v1/tools/runs/:id`
   - `POST /api/v1/tools/:id/execute`
4. Canvas/template pickers using:
   - `GET /api/v1/artifacts/templates`
   - `POST /api/v1/artifacts/from-template`
   - `GET /api/v1/artifacts/new-doc?channel_id=...`

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
