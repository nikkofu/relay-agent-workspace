# Phase 10: File Assets

## Goal

Add real file asset APIs so Relay can move from mock attachments toward durable uploads that can be referenced by messages and artifacts.

## Scope

- `POST /api/v1/files/upload`
- `GET /api/v1/files`
- `GET /api/v1/files/:id`
- `GET /api/v1/files/:id/content`
- artifact responses now include hydrated editor user data

## Backend Tasks

- persist uploaded file metadata
- store file binaries under local upload storage for dev
- return uploader metadata and fetchable content URLs
- hydrate `created_by_user` and `updated_by_user` on artifact responses

## Frontend Handoff

- file picker flows can upload with multipart `file` plus optional `channel_id`
- list files via `GET /api/v1/files?channel_id=...`
- use `GET /api/v1/files/:id` for metadata and `url` for direct fetch/download
- artifact surfaces now receive `created_by_user` and `updated_by_user`

## Verification

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`
