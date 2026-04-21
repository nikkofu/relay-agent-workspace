# Phase 10: Operational Maturity, Governance, and Workflow Control

## Goal

Push Relay's Slack-style workspace from editable admin surfaces into governed, durable operational flows:

- richer custom status behavior
- user-group membership management and mentions
- file retention and audit visibility
- workflow run detail, cancel, and retry

## Scope

This phase delivers:

- expanded `PATCH /api/v1/users/:id/status`
- `GET /api/v1/user-groups/mentions`
- `GET /api/v1/user-groups/:id/members`
- `POST /api/v1/user-groups/:id/members`
- `DELETE /api/v1/user-groups/:id/members/:userId`
- `GET /api/v1/workflows/runs/:id`
- `POST /api/v1/workflows/runs/:id/cancel`
- `POST /api/v1/workflows/runs/:id/retry`
- `PATCH /api/v1/files/:id/retention`
- `GET /api/v1/files/:id/audit`

## Why This Phase Matters

Relay already had the visible shell of profiles, groups, files, and workflows. The next gap was operating them with the kind of controls users expect from a serious Slack-like workspace:

- statuses need more than plain text
- groups need membership-specific APIs for mentions and management
- files need retention and audit signals, not just upload/delete
- workflow runs need lifecycle control, not just a launch button

## Backend Work

- added custom status emoji and optional expiration windows
- added dedicated membership endpoints for user groups
- added mention-oriented user-group lookup for composer and search surfaces
- added file retention policy updates and file audit event history
- added workflow run detail lookup plus cancel and retry actions
- expanded realtime `workflow.run.updated` coverage to include cancel and retry transitions

## Frontend Handoff

Gemini can now build:

- richer status picker UI with optional expiry presets
- explicit group membership editors and `@group` lookup UX
- file governance surfaces for retention and audit history
- workflow run detail drawers with cancel/retry controls

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm build` still hangs in this environment after `Creating an optimized production build ...`
