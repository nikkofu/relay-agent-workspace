# Phase 10: Intelligent Search, Backlinks, And Notification Sync

## Goal

Expand Relay from object discovery into object intelligence and cross-surface state sync.

## Scope

- add backlink lookup from artifacts to referencing messages
- add an intelligent ranked search endpoint across channels, messages, artifacts, and files
- broadcast notification read-state changes over websocket

## API Deliverables

- `GET /api/v1/artifacts/:id/references`
- `GET /api/v1/search/intelligent?q=...`
- `POST /api/v1/notifications/read`
  - now broadcasts realtime `notifications.read`

## Realtime

- new event: `notifications.read`
  - payload:
    - `user_id`
    - `item_ids`
    - `read`

## Collaboration Split

- Codex:
  - backlink handler
  - intelligent ranking handler
  - notification read realtime sync
  - docs and release
- Gemini:
  - artifact detail backlinks UI
  - command-palette or search panel support for ranked results
  - inbox / mentions store reconciliation on `notifications.read`

## Acceptance

- artifact detail can show where the canvas or document has been referenced in chat
- intelligent search returns ranked, typed results with score and reason
- read-state changes can propagate across multiple open clients
