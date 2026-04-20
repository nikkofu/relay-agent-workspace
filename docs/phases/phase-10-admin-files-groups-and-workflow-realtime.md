# Phase 10: Admin Profiles, Group CRUD, File Operations, and Workflow Realtime

## Goal

Push Relay's Slack-like operational shell from browse-only surfaces into editable/admin-capable flows:

- editable user profiles
- full user group CRUD
- richer file management
- workflow run realtime visibility

## Scope

This phase delivers:

- `PATCH /api/v1/users/:id`
- `POST /api/v1/user-groups`
- `PATCH /api/v1/user-groups/:id`
- `DELETE /api/v1/user-groups/:id`
- `DELETE /api/v1/files/:id`
- expanded file list/archive filters
- realtime `workflow.run.updated`

## Why This Phase Matters

Relay already had the read-side contracts for home, directory, workflows, notifications, and file archive. The next gap was control:

- profile information needed to be editable
- user groups needed lifecycle APIs
- files needed deletion and better filtering
- workflow runs needed realtime visibility for orchestration-style UI

## Backend Work

- added profile patch support for:
  - `title`
  - `department`
  - `timezone`
  - `working_hours`
- added create/update/delete flows for user groups and member replacement
- added file deletion
- added uploader/content-type/archive filters to file queries
- added realtime broadcast for workflow run creation

## Frontend Handoff

Gemini can now build:

- editable profile details
- user group creation/edit/delete flows
- richer file management and cleanup surfaces
- live workflow run indicators and activity surfaces

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
