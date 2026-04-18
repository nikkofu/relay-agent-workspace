# Phase 10.3: Drafts

## Goal

Add Slack-style draft persistence so Relay can preserve unfinished composer state across channels, DMs, and threads.

This gives Gemini a stable backend target for autosave and draft restore without waiting for richer notification or presence work.

## Scope

### Backend

- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`

### Contract Shape

- drafts list response:
  - `{ "drafts": [...] }`
- draft upsert response:
  - `{ "draft": {...} }`
- draft fields:
  - `id`
  - `user_id`
  - `scope`
  - `content`
  - `created_at`
  - `updated_at`

### Scope Convention

Drafts are keyed by `scope`, not by route path. Current recommended scope shapes are:

- `channel:<channelId>`
- `dm:<dmId>`
- `thread:<messageId>`

This keeps the backend generic while letting Gemini map drafts to whichever composer surface is active.

## Design Decision

This phase stores one draft per `user + scope`.

Why:

- matches how Slack-style drafts behave in practice
- avoids over-modeling draft history too early
- lets the frontend overwrite draft content safely on autosave
- keeps the contract compatible with future surfaces like AI-assisted composer drafts

## Roles

### Codex

- draft persistence model
- API contract and handler tests
- release notes and repository docs
- Gemini handoff for scope usage

### Gemini

- composer autosave integration
- draft restore on channel / DM / thread switch
- local debounce and optimistic UI polish
- feedback on any missing delete/clear workflow

## Execution Steps

1. Add failing handler tests for listing and upserting drafts.
2. Add a persistence model keyed by `user_id + scope`.
3. Register routes in `apps/api/main.go`.
4. Seed one example draft for local dev visibility.
5. Update collaboration docs and release notes.

## Delivered

- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`
- `drafts` persistence table keyed by `user_id + scope`

## Next Recommended Phase Order

1. `presence + typing`
2. stars and pinned surfaces
3. persistent notification read state
4. AI conversation persistence
