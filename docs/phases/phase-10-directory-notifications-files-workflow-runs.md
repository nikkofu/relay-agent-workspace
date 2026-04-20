# Phase 10: Directory Filters, Notification Preferences, File Archive, and Workflow Runs

## Goal

Extend Relay's Slack-like workspace foundation into the operational surfaces that sit around messaging:

- richer people directory filtering
- explicit notification preferences and mute rules
- file archive lifecycle
- workflow run history and manual run creation

## Scope

This phase delivers:

- `GET /api/v1/users` filtering by:
  - `q`
  - `department`
  - `status`
  - `timezone`
  - `user_group_id`
- `GET /api/v1/notifications/preferences`
- `PATCH /api/v1/notifications/preferences`
- `GET /api/v1/files/archive`
- `PATCH /api/v1/files/:id/archive`
- `GET /api/v1/workflows/runs`
- `POST /api/v1/workflows/:id/runs`

## Why This Phase Matters

Relay already had dynamic channels, DMs, AI, artifacts, search, home, and groups. The next missing layer was the operational model around that collaboration:

- users need a real directory instead of raw people lists
- notification behavior needs backend preferences and mute rules
- files need a lifecycle beyond upload-only
- workflows need observable runs, not just static definitions

## Backend Work

- added `workflow_runs`
- added `notification_preferences`
- added `notification_mute_rules`
- extended `file_assets` with archive state
- extended `GET /api/v1/users` with directory filters
- added workflow run hydration with workflow and starter user objects
- added notification preference persistence with scoped mute rules

## Frontend Handoff

Gemini can now:

- build filtered people directory and group-based people views
- create notification settings and mute surfaces
- add archived file views and archive toggles
- expose workflow run history and manual workflow trigger UI

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
