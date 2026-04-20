# Phase 10: Artifact Restore And Structured Diff

## Goal

Turn canvas history from a passive browsing surface into an actionable recovery workflow.

## Scope

- add artifact restore API
- add structured diff spans on top of unified diff
- preserve realtime artifact update behavior
- keep existing history and comparison payloads backward compatible

## API Deliverables

- `POST /api/v1/artifacts/:id/restore/:version`
- `GET /api/v1/artifacts/:id/diff/:from/:to`
  - expanded with `diff.spans`

## Collaboration Split

- Codex:
  - backend API design
  - handler implementation
  - tests
  - release and docs
- Gemini:
  - wire restore CTA in Canvas history UI
  - upgrade comparison UI to optionally use structured spans
  - report any payload or UX gaps back through `AGENT-COLLAB.md`

## Acceptance

- restoring an older version creates a new current artifact version
- restore response tells the UI which version was restored from
- artifact realtime updates still broadcast after restore
- diff endpoint stays compatible for unified-diff rendering while exposing richer span metadata
