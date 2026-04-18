# Phase 10.2: Inbox And Mentions

## Goal

Add the next Slack-parity notification layer after channel management:

- inbox
- mentions

This gives Gemini a stable backend target for the next notification-oriented UI surfaces without waiting for a full persistent notification center.

## Scope

### Backend

- `GET /api/v1/inbox`
- `GET /api/v1/mentions`

### Contract Shape

- inbox response:
  - `{ "items": [...] }`
- mentions response:
  - `{ "items": [...] }`
- item shape is aligned with the activity feed:
  - `id`
  - `type`
  - `user`
  - optional `channel`
  - optional `message`
  - optional `target`
  - `summary`
  - `occurred_at`

## Design Decision

This phase intentionally reuses existing collaboration signals instead of introducing a persisted notification table immediately.

Signals currently included:

- direct mentions
- thread replies to your messages
- reactions on your messages
- incoming DM activity

Why:

- keeps the implementation incremental
- matches the current workspace data already in SQLite
- gives Gemini enough stable data to build the UI now
- leaves room to add notification read state in a later phase

## Roles

### Codex

- API contract
- feed aggregation logic
- tests, docs, release packaging

### Gemini

- inbox page / panel
- mentions surface
- list rendering and navigation
- notification UI feedback

## Execution Steps

1. Add failing handler tests for inbox and mentions.
2. Reuse the activity aggregation logic behind a common feed builder.
3. Add routes in `apps/api/main.go`.
4. Update docs and handoff notes.
5. Release a unified version with the latest Gemini frontend state.

## Delivered

- `GET /api/v1/inbox`
- `GET /api/v1/mentions`

## Next Recommended Phase Order

1. `drafts`
2. `presence + typing`
3. persistent notification read state
4. AI conversation persistence
