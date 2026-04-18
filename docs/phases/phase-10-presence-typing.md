# Phase 10.4: Presence And Typing

## Goal

Add the next realtime Slack-parity layer so Relay can show live user availability and typing intent across channels, DMs, and threads.

This phase is intentionally lightweight: persistent presence, event-driven typing.

## Scope

### Backend

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`

### Realtime Events

- `presence.updated`
- `typing.updated`

### Contract Shape

- presence list response:
  - `{ "users": [...] }`
- presence update response:
  - `{ "user": {...} }`
- typing update response:
  - `{ "typing": {...} }`

### Typing Payload

- `user_id`
- optional `channel_id`
- optional `thread_id`
- optional `dm_id`
- `is_typing`

## Design Decision

Presence is stored on the existing `users.status` field. Typing is not persisted and is broadcast as a websocket event.

Why:

- presence needs a stable current value for initial page load
- typing is transient and should not create database churn
- this keeps the API simple while remaining compatible with future Redis-backed realtime fanout

## Roles

### Codex

- API contract
- realtime event emission
- tests, docs, release packaging

### Gemini

- user status hydration in avatar/profile surfaces
- typing indicator UI in channel, DM, and thread composers
- websocket event consumption and local timeout behavior

## Execution Steps

1. Add failing tests for presence list/update and typing event broadcast.
2. Reuse `users.status` for persisted presence.
3. Add typing event endpoint and websocket broadcast.
4. Update docs, changelog, and Gemini handoff notes.

## Delivered

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`
- websocket events:
  - `presence.updated`
  - `typing.updated`

## Next Recommended Phase Order

1. stars and pinned surfaces
2. persistent notification read state
3. AI conversation persistence
