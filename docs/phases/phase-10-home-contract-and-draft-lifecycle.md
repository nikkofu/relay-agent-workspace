# Phase 10 Home Contract And Draft Lifecycle

## Goal

Harden the home payload so the current dashboard can consume direct aliases without store-side guesswork, and add explicit draft deletion for cleaner composer lifecycle control.

## Delivered

- `GET /api/v1/home`
  now includes:
  - `stats`
  - `recent_activity`
  - top-level `recent_artifacts`
- `DELETE /api/v1/drafts/:scope`

## Backend Notes

- `stats.pending_actions` mirrors unread/pending work for the home hero and stat cards
- `stats.active_threads` summarizes current thread activity across the user’s channels
- `recent_activity` gives the home dashboard channel-aware cards with:
  - `id`
  - `channel_id`
  - `channel_name`
  - `last_message`
  - `occurred_at`
- draft deletion is scoped by current user plus exact draft scope, so channel, DM, and thread drafts remain isolated

## Gemini Handoff

1. use direct `home.stats` / `home.recent_activity` / `home.recent_artifacts` fields instead of relying on partial nested fallbacks
2. consider calling `DELETE /api/v1/drafts/:scope` on explicit composer discard or after successful send where UX expects a hard clear
3. keep backward compatibility if you want gradual cleanup, but the backend contract is now ready

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
