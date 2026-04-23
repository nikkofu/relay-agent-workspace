# Phase 63F Always-On Knowledge Automation Design

## Goal

Make channel summaries and AI compose suggestions observable workspace events, while keeping the implementation small enough for the current Phase 63 UI loop.

## Scope

This phase delivers:

- channel auto-summarize settings and manual run API
- websocket `channel.summary.updated`
- websocket `knowledge.compose.suggestion.generated`
- additive `compose.proposed_slots[]` for schedule intent

It intentionally does not add a background entity-brief worker yet. That should be a separate debounced worker slice because it needs throttling, retry policy, and per-entity caps.

## API Design

### Channel Auto-Summarize

- `GET /api/v1/channels/:id/knowledge/auto-summarize`
- `PUT /api/v1/channels/:id/knowledge/auto-summarize`
- `POST /api/v1/channels/:id/knowledge/auto-summarize`

Settings persist in `ChannelAutoSummarySetting` with:

- channel/workspace ownership
- `is_enabled`
- `window_hours`
- `message_limit`
- `min_new_messages`
- optional provider/model override
- `last_run_at`
- `last_message_at`

`POST` reuses the existing `AISummary` channel cache instead of adding a parallel summary table.

### Realtime

`POST /channels/:id/knowledge/auto-summarize` emits:

- `channel.summary.updated`

Compose APIs emit after final suggestions are built:

- `knowledge.compose.suggestion.generated`

### Schedule Intent

`schedule` compose responses add:

- `proposed_slots[]`

The field is additive and does not break existing suggestion cards.

## Testing

The behavior is covered by `TestPhase63FAutoSummarizeAndComposeRealtimeContracts`, which asserts:

- setting persistence
- summary execution
- `channel.summary.updated`
- `knowledge.compose.suggestion.generated`
- `proposed_slots[]`
