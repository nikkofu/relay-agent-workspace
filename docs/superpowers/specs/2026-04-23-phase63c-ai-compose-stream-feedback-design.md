# Phase 63C AI Compose Stream And Feedback Design

## Goal

Extend the existing grounded composer assistant from one-shot synchronous suggestions into a stream-capable reply assistant with explicit per-suggestion feedback.

## Scope

This phase adds:

- `POST /api/v1/ai/compose/stream`
- `POST /api/v1/ai/compose/:id/feedback`

This phase does not add:

- DM compose parity
- new compose intents beyond `reply`
- online re-ranking or model fine-tuning
- persistence of full compose sessions

## Why This Slice

`POST /api/v1/ai/compose` is already wired into the shared channel/thread composer UI. The next highest-value backend step is to make long suggestions stream progressively and to capture explicit user signal on the suggestions the assistant generates.

That keeps the work on the Slack-style primary loop:

- read context
- draft reply
- get grounded AI help
- accept or reject the help

## API Design

### `POST /api/v1/ai/compose/stream`

Request body:

```json
{
  "channel_id": "ch-1",
  "thread_id": "msg-parent",
  "draft": "Can we confirm owner and deadline?",
  "intent": "reply",
  "limit": 3
}
```

Rules:

- `channel_id` is required
- `thread_id` is optional
- `intent` defaults to `reply`
- only `reply` is accepted in this phase
- `limit` is normalized exactly like `POST /api/v1/ai/compose`
  - `< 1 => 3`
  - `> 5 => 5`

Response type:

- `text/event-stream`

SSE event contract:

- `start`
  - payload: `{ provider, model, request_id, channel_id, thread_id?, intent, limit }`
- `suggestion.delta`
  - payload: `{ suggestion_id, index, text_delta }`
- `suggestion.done`
  - payload: `{ suggestion, citations, context_entities }`
- `done`
  - payload: `{ request_id, suggestion_count, provider, model }`
- `error`
  - payload: `{ message }`

Behavior:

- reuse the exact grounding pipeline from `POST /api/v1/ai/compose`
- stream suggestion text progressively from the LLM stream
- emit exactly one `suggestion.done` per parsed suggestion
- if the provider output cannot be parsed into structured compose JSON, fall back to one synthesized reply suggestion based on grounded context
- no DB persistence in this endpoint

ID rules:

- `request_id` identifies one stream request
- `suggestion_id` identifies one suggestion inside that request
- `POST /api/v1/ai/compose/:id/feedback` uses that same `suggestion_id` as `:id`
- server-normalized suggestion ids must be stable in both sync and stream compose responses
- in this phase, `suggestion.delta` streams only the first provisional suggestion (`index = 0`) while the final parsed payload may still emit up to `limit` `suggestion.done` events

### `POST /api/v1/ai/compose/:id/feedback`

Request body:

```json
{
  "channel_id": "ch-1",
  "thread_id": "msg-parent",
  "intent": "reply",
  "feedback": "edited",
  "suggestion_text": "Let's confirm the owner and keep Friday as the target.",
  "provider": "openrouter",
  "model": "nvidia/nemotron-3-super-120b-a12b:free"
}
```

Rules:

- `:id` is the compose suggestion id returned by sync or stream compose
- `feedback` must be one of:
  - `up`
  - `down`
  - `edited`
- `channel_id` is required
- `thread_id` is optional
- `intent` defaults to `reply`
- `suggestion_text` is optional but recommended so edited flows can be analyzed later
- `provider` and `model` are optional passthrough fields from the client; when absent they persist as empty strings in this phase

Response body:

```json
{
  "feedback": {
    "id": "compose-feedback-...",
    "compose_id": "compose-suggestion-...",
    "user_id": "user-1",
    "channel_id": "ch-1",
    "thread_id": "msg-parent",
    "intent": "reply",
    "feedback": "edited",
    "suggestion_text": "Let's confirm the owner and keep Friday as the target.",
    "provider": "openrouter",
    "model": "nvidia/nemotron-3-super-120b-a12b:free",
    "created_at": "...",
    "updated_at": "..."
  }
}
```

Behavior:

- store one row per `(compose_id, user_id)`
- later feedback updates the same row
- provider/model should be persisted exactly as supplied by the client; no server-side compose-session lookup is introduced in this phase

## Persistence

Add a dedicated `AIComposeFeedback` model instead of overloading `AIFeedback`.

Reason:

- current `AIFeedback` is keyed to message ids and thumbs-up/down on message-level AI usage
- compose feedback is pre-send and suggestion-scoped
- keeping them separate avoids muddying semantics and unique indexes

Suggested fields:

- `ID`
- `ComposeID`
- `UserID`
- `ChannelID`
- `ThreadID`
- `Intent`
- `Feedback`
- `SuggestionText`
- `Provider`
- `Model`
- `CreatedAt`
- `UpdatedAt`

## Implementation Approach

### Shared compose preparation

Extract the synchronous compose request preparation into reusable helpers:

- validate scope
- load channel and optional thread parent
- load recent messages
- load channel knowledge context
- match entities from draft/context
- build prompt

That keeps sync and stream behavior aligned.

### Stream parsing

We do not need true token-level structured JSON assembly from the model. For this phase:

- accumulate streamed `chunk` text
- forward chunks as `suggestion.delta` against a single active suggestion id
- when the upstream stream ends, parse the full content using the same compose parser
- emit final `suggestion.done` events for the parsed suggestions

This gives Windsurf a progressive UI without forcing a much larger parser architecture today.

## Error Handling

- `400` for invalid request body or unsupported `intent`
- `404` for unknown channel or thread parent
- `503` when the AI gateway is not configured
- SSE `error` event for provider failure after headers are sent
- if `http.Flusher` is unavailable, return `500`

## Testing

Handler tests should cover:

- compose stream contract for channel scope
- compose stream contract for thread scope
- compose stream returns `Content-Type: text/event-stream`
- compose stream emits `error` when the gateway stream fails after headers are written
- compose stream returns `503` without gateway
- compose feedback creates a row
- compose feedback upserts on repeated submission by same user for same compose id
- compose feedback rejects invalid feedback values

## Windsurf Handoff

Once shipped, Windsurf can:

- progressively render suggestion text in the composer popover
- add thumbs-up / thumbs-down / edited actions per suggestion
- keep the existing sync endpoint as fallback when SSE is unavailable
