# Phase 10.5: Stars And Pins

## Goal

Add the next Slack-parity discovery layer so Relay can surface starred channels and pinned messages as first-class workspace destinations.

This phase separates two distinct concepts:

- channel starring for navigation preference
- message pinning for persistent shared references

## Scope

### Backend

- `GET /api/v1/starred`
- `POST /api/v1/channels/:id/star`
- `GET /api/v1/pins`

### Contract Shape

- starred response:
  - `{ "channels": [...] }`
- channel star toggle response:
  - `{ "channel": {...}, "is_starred": boolean }`
- pins response:
  - `{ "items": [...] }`

### Pins Item Shape

- `message`
- `channel`
- `user`

## Design Decision

This phase reuses existing persisted fields:

- `channels.is_starred`
- `messages.is_pinned`

Why:

- channel starring is already part of the channel model
- message pinning already exists in the interaction API layer
- the missing piece was discovery/list APIs, not new storage

## Roles

### Codex

- API contract
- list and toggle handlers
- release notes and repository docs
- Gemini handoff for starred and pinned surfaces

### Gemini

- starred section integration and any dedicated starred view
- pinned messages surface and navigation
- websocket polish for channel/message updates if needed

## Execution Steps

1. Add failing handler tests for starred list, star toggle, and pins list.
2. Reuse `channels.is_starred` and `messages.is_pinned`.
3. Register new routes in `apps/api/main.go`.
4. Seed one pinned message for local development visibility.
5. Update docs, changelog, and collaboration notes.

## Delivered

- `GET /api/v1/starred`
- `POST /api/v1/channels/:id/star`
- `GET /api/v1/pins`

## Next Recommended Phase Order

1. persistent notification read state
2. AI conversation persistence
3. richer search and artifact lifecycle
