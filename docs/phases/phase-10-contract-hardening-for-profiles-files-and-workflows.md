# Phase 10: Contract Hardening For Profiles, Files, And Workflows

## Goal

Make Relay's existing operational shell more dependable by enriching the backend contracts that current pages already rely on:

- richer personal profile data
- workflow run detail that actually contains execution structure
- file payloads and audit logs aligned with the current UI expectations

## Scope

This phase delivers:

- extended `PATCH /api/v1/users/:id`
- richer `GET /api/v1/users/:id`
- richer `GET /api/v1/workflows/runs`
- richer `GET /api/v1/workflows/runs/:id`
- richer `GET /api/v1/files`
- richer `GET /api/v1/files/:id`
- richer `GET /api/v1/files/:id/audit`

## Why This Phase Matters

Relay already had the right pages and most of the right endpoints, but some payloads were still too thin:

- profile editing stopped at basic workplace identity
- workflow detail could open without meaningful step history
- files pages still depended on frontend-side assumptions about key names

This phase closes those gaps without forcing Gemini to keep adding adapter logic across stores.

## Backend Work

- added profile fields for:
  - `pronouns`
  - `location`
  - `phone`
  - `bio`
- added `workflow_run_steps`
- enriched workflow run responses with flat compatibility fields plus step history
- enriched file responses with compatibility aliases for existing web stores
- returned both `events` and `audit_history` from file audit for safer UI integration
- updated seed data so local dev immediately shows richer profiles and workflow detail

## Frontend Handoff

Gemini can now:

- expand profile edit dialogs and profile cards without waiting on new backend fields
- treat workflow run detail as a real execution object, not a placeholder shell
- consume file list/detail/audit payloads with less ad hoc mapping

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
