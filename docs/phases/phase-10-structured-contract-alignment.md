# Phase 10: Structured Contract Alignment

This phase does not introduce a new product surface. It hardens the contracts behind the newly integrated structured-work surfaces so the frontend can consume them without local inference or temporary shape adapters.

## Why This Phase Matters

After Phase 33, Gemini had already integrated:

- workspace lists
- tool-run history
- template-first canvas creation

The remaining issue was contract drift:

- list creation from channels did not provide `workspace_id`
- tool-run stores expected UI-friendly fields like `finished_at` and `duration_ms`
- artifact stores still looked for `user_id` on virtual/template artifact payloads

If left unresolved, this drift would create hidden regressions and ad hoc frontend workarounds across Home, channel, and canvas surfaces.

## Backend Hardening

- `POST /api/v1/lists` now derives `workspace_id` from `channel_id` when omitted
- list payloads now include:
  - `user_id`
- list item payloads now include:
  - `list_id`
  - `user_id`
- tool run payloads now include:
  - `user_id`
  - `channel_id`
  - `finished_at`
  - `duration_ms`
- `GET /api/v1/tools/runs` now supports `channel_id` filtering
- virtual/template artifact payloads now include:
  - `user_id`

## Gemini Handoff

Gemini can now treat Phase 33 as complete. Remaining optional cleanup:

1. Replace local `new-doc` stub hydration in the artifact store with `GET /api/v1/artifacts/new-doc?channel_id=...`
2. Remove temporary frontend shape assumptions where list/tool/artifact aliases now exist
3. Update `docs/AGENT-COLLAB.md` from `frontend-cleanup` to `idle` once that simplification pass lands

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
