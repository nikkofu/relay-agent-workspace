# Phase 63G Compose Activity Persistence Design

## Goal

Persist AI compose suggestion activity so co-drafting observability survives page refresh and can feed agent-collab or activity surfaces.

## Scope

This phase adds:

- `AIComposeActivity` persistence
- `GET /api/v1/ai/compose/activity`
- websocket payload enrichment for `knowledge.compose.suggestion.generated`

It does not add a new frontend page, server-side analytics dashboard, or calendar booking.

## Data Model

`AIComposeActivity` stores one row per finalized compose request:

- `id`
- `compose_id`
- `workspace_id`
- `channel_id`
- `dm_id`
- `thread_id`
- `intent`
- `suggestion_count`
- `provider`
- `model`
- `created_at`

`compose_id` is unique to keep sync/stream duplicate finalization idempotent.

## API

`GET /api/v1/ai/compose/activity`

Filters:

- `channel_id`
- `dm_id`
- `workspace_id`
- `intent`
- `limit`, capped at 100

Response:

```json
{ "items": [] }
```

## Realtime

`knowledge.compose.suggestion.generated` keeps `compose` and adds `activity`.

## Testing

`TestPhase63GComposeActivityPersistsAndLists` covers:

- activity persistence after compose
- channel-scoped filtering
- newest-first list envelope
- key metadata fields
