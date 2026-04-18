# Phase 10.7: AI Conversation Persistence

## Goal

Add persistent AI conversation history so Relay's assistant panel is not just a streaming surface, but a recoverable workspace memory layer.

This phase keeps the existing `POST /api/v1/ai/execute` interaction and adds persistence beneath it.

## Scope

### Backend

- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`
- `POST /api/v1/ai/execute`

### Behavior

- `POST /api/v1/ai/execute` now persists:
  - the user prompt
  - the assistant response
  - reasoning text
  - provider/model metadata
- `conversation_id` can be provided to continue an existing AI thread
- if `conversation_id` is omitted, a new one is created and returned in the stream

## Contract Shape

- list response:
  - `{ "conversations": [...] }`
- detail response:
  - `{ "conversation": {...}, "messages": [...] }`

## Design Decision

Persistence is added to the existing execute path instead of introducing a separate create-conversation endpoint first.

Why:

- avoids duplicating the AI entrypoint
- keeps Gemini's current integration path stable
- lets the frontend adopt history incrementally

## Roles

### Codex

- conversation and message persistence model
- streaming-path persistence logic
- list/detail read APIs
- docs, release notes, Gemini handoff

### Gemini

- conversation history sidebar or restore surface
- loading past AI messages into the assistant panel
- choosing when to continue an existing conversation versus start a new one

## Execution Steps

1. Add failing handler tests for persisted execute flow and conversation reads.
2. Add `ai_conversations` and `ai_conversation_messages`.
3. Persist AI output in `POST /api/v1/ai/execute`.
4. Add list/detail APIs.
5. Update docs and Gemini handoff notes.

## Delivered

- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`
- persisted AI conversations and messages behind `POST /api/v1/ai/execute`

## Next Recommended Phase Order

1. AI thread and channel summaries
2. artifact lifecycle
3. richer search and semantic retrieval
