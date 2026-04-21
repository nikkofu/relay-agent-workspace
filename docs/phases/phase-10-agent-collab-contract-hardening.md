# Phase 41 - Agent-Collab Contract Hardening

Date: 2026-04-21
Owner: Codex
Frontend Partner: Windsurf
Status: Backend complete, frontend simplification queued

## Goal

Remove avoidable client-side shape repair from the live #agent-collab hub while preserving backward compatibility.

## Delivered

- `CommLogEntry.to` is always serialized.
- Direct Markdown participants using `A → B` parse into `from` and `to`.
- Broadcast entries serialize `to` as an empty string.
- `MemberProfile.primary_tools_array` provides normalized tool names.
- `MemberProfile.primary_tools` remains available for older clients.

## Frontend Handoff

Windsurf can now simplify the hub store:

- Prefer `member.primary_tools_array`.
- Use string splitting only if `primary_tools_array` is missing.
- Use `comm_log.to` to render direct messages.
- Treat empty `to` as Broadcast.

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/agentcollab -run 'TestParseMarkdownReturnsDirectCommTargetsAndToolArrays$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestAgentCollabMembersAndCommLogEndpoints$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
