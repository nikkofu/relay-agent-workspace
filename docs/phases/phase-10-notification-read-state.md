# Phase 10.6: Notification Read State

## Goal

Add persistent read state for inbox and mentions so Relay notifications are not only visible, but also manageable.

This phase keeps the existing feed-generation model and adds read tracking on top.

## Scope

### Backend

- `POST /api/v1/notifications/read`

### Feed Changes

- `GET /api/v1/inbox`
- `GET /api/v1/mentions`

Both now include:

- `is_read`

## Contract Shape

- mark read request:
  - `{ "item_ids": string[] }`
- mark read response:
  - `{ "read": true, "item_ids": [...] }`

## Design Decision

Read state is tracked by `user_id + item_id` in a dedicated table.

Why:

- inbox and mentions items are synthesized from collaboration signals
- the feed itself should remain derived, not duplicated into a notification table yet
- a small read-state table keeps the current architecture incremental and reversible

## Roles

### Codex

- read state persistence model
- feed enrichment with `is_read`
- batch read API
- docs, release notes, handoff

### Gemini

- unread/read visual treatment in inbox and mentions
- mark-as-read on click or explicit action
- optional mark-all behavior by sending visible `item_ids`

## Execution Steps

1. Add failing tests for unread feed items and mark-read persistence.
2. Add a `notification_reads` model keyed by `user_id + item_id`.
3. Enrich activity-derived feed items with `is_read`.
4. Add `POST /api/v1/notifications/read`.
5. Update docs and Gemini handoff notes.

## Delivered

- `POST /api/v1/notifications/read`
- `is_read` on inbox and mentions items
- `notification_reads` persistence table

## Next Recommended Phase Order

1. AI conversation persistence
2. richer search and artifact lifecycle
3. notification preferences and mute rules
