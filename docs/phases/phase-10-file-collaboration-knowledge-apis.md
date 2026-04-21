# Phase 42 - File Collaboration And Knowledge Metadata APIs

Date: 2026-04-21
Owner: Codex
Frontend Partner: Windsurf
Status: Backend complete, frontend integration queued

## Goal

Move files from passive uploads into collaborative workspace knowledge objects that can be discussed, shared into conversation, starred, and prepared for future LLM/wiki retrieval.

## Delivered

- File comments:
  - `GET /api/v1/files/:id/comments`
  - `POST /api/v1/files/:id/comments`
- File shares:
  - `GET /api/v1/files/:id/shares`
  - `POST /api/v1/files/:id/share`
- File stars:
  - `POST /api/v1/files/:id/star`
  - `GET /api/v1/files/starred`
- File knowledge metadata:
  - `PATCH /api/v1/files/:id/knowledge`

## Notes

- Shares currently target channels/threads by creating a real `Message` and `MessageFileAttachment`.
- DM-targeted file shares are intentionally deferred to the next slice.
- Knowledge fields are stored now so wiki and LLM retrieval layers can consume them later without another model migration.

## Windsurf Handoff

- Integrate comments and shares into file detail first.
- Add star toggles in Files list/cards.
- Surface `source_kind`, `knowledge_state`, `summary`, and `tags` in file detail/edit UX.
- Prefer “knowledge file” affordances when `source_kind = wiki` and `knowledge_state = ready`.

## Verification

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestFile(CollaborationCommentsStarsAndKnowledgeMetadata|ShareCreatesAttachmentMessageAndAuditEvent)$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
