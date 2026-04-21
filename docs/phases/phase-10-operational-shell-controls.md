# Phase 32: Operational Shell Controls

Date: 2026-04-21

Owner: Codex

## Goal

Close the next set of Slack-parity operational gaps around workflows, files, and channel administration while preserving Relay Agent Workspace's AI-native audit trail direction.

## Backend Scope

- Workflow execution observability:
  - `GET /api/v1/workflows/runs/:id/logs`
  - `DELETE /api/v1/workflows/runs/:id`
  - persisted `workflow_run_logs`
  - realtime `workflow.run.deleted`
- File preview metadata:
  - `GET /api/v1/files/:id/preview`
  - direct preview/download URLs for image/PDF-capable files
  - fallback download metadata for non-previewable files
- Channel controls:
  - `GET /api/v1/channels/:id/preferences`
  - `PATCH /api/v1/channels/:id/preferences`
  - `POST /api/v1/channels/:id/leave`
  - persisted per-user channel notification preferences

## Gemini Handoff

Gemini owns frontend integration after this backend release:

- Wire Workflows page `Delete Log` to `DELETE /api/v1/workflows/runs/:id`.
- Add workflow run log viewing via `GET /api/v1/workflows/runs/:id/logs`.
- Use `GET /api/v1/files/:id/preview` for file preview dialog/card metadata.
- Connect ChannelInfo notification and leave rows to the new channel APIs.

## Verification

- Add contract tests for each new endpoint.
- Run the API package tests, API build, and workspace web lint before release.
