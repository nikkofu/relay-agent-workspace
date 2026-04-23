# Phase 63B AI Compose Design

## Goal

Add a grounded AI compose endpoint for the Slack-like message flow so channel and thread composers can request reply suggestions backed by recent conversation context and workspace knowledge.

## Scope

This phase only covers:

- channel composer
- thread composer inside a channel
- synchronous request/response compose suggestions

This phase does not cover:

- DM compose
- SSE or token streaming
- persistence of compose sessions/history
- automatic sending or draft mutation
- `summarize` or `rewrite` compose intents

## Why This Slice

The current product already has:

- channel and thread messages
- persisted AI execution
- knowledge entity refs and citations
- entity match-text for draft detection
- entity ask/brief surfaces

The missing piece is direct AI assistance at the point of writing. `POST /api/v1/ai/compose` is the smallest high-value backend slice that connects the existing AI-native knowledge system to the core Slack-style composer.

## API Contract

### Endpoint

- `POST /api/v1/ai/compose`

### Request

```json
{
  "channel_id": "ch-1",
  "thread_id": "msg-parent",
  "draft": "Can we confirm the launch owner and timeline?",
  "intent": "reply",
  "limit": 3
}
```

### Request Rules

- `channel_id` is required
- `thread_id` is optional
- `draft` is optional but recommended
- `intent` is optional and this phase only supports `reply`
- `limit` is optional
  - default `3`
  - values `< 1` normalize to `3`
  - values `> 5` clamp to `5`

### Response

```json
{
  "compose": {
    "channel_id": "ch-1",
    "thread_id": "msg-parent",
    "intent": "reply",
    "suggestions": [
      {
        "id": "cmp-1",
        "text": "I can take the owner update. Current launch target still looks aligned with the Friday review.",
        "tone": "clear",
        "kind": "reply"
      }
    ],
    "citations": [
      {
        "id": "kref-1",
        "evidence_kind": "message",
        "source_kind": "message",
        "source_ref": "msg-4",
        "ref_kind": "message",
        "snippet": "Launch Program kickoff has clear owners",
        "title": "Launch Program kickoff has clear owners",
        "score": 2,
        "entity_id": "entity-1",
        "entity_title": "Launch Program"
      }
    ],
    "context_entities": [
      {
        "id": "entity-1",
        "title": "Launch Program",
        "kind": "project"
      }
    ],
    "provider": "openrouter",
    "model": "nvidia/nemotron-3-super-120b-a12b:free"
  }
}
```

## Grounding Strategy

The endpoint should assemble a compact prompt from:

1. recent thread replies when `thread_id` is provided
2. otherwise recent channel messages
3. recent knowledge refs for the channel
4. entity matches detected from `draft`
5. optional entity snippets tied to the matched refs

The response should expose:

- `suggestions[]` for direct UI use
- `citations[]` for trust and traceability
- `context_entities[]` so the UI can explain what the compose engine grounded against

## Prompt Shape

The prompt should instruct the model to:

- act as an AI writing assistant for a Slack-like workspace
- stay grounded in the provided context
- avoid fabricating owners, dates, or decisions
- produce a small set of distinct suggestions
- keep suggestions concise and send-ready

Suggestion tone should be practical, short, and collaboration-native.

## Error Handling

- `400` when `channel_id` is missing or `intent` is invalid
- `400` when `limit` is non-numeric or malformed at the JSON layer
- `404` when `channel_id` does not exist
- `404` when `thread_id` is provided but the parent thread message does not exist
- `503` when the AI gateway is not configured
- `502` when upstream provider generation fails

## Testing

Add focused coverage for:

- channel compose suggestions
- thread compose suggestions
- `draft` entity matching feeding `context_entities`
- invalid channel/thread handling
- AI gateway missing behavior

## Windsurf Handoff

Once shipped, Windsurf should be able to:

- add an AI suggestions rail or popover in `message-composer`
- request compose suggestions for channel and thread scopes
- display citations/context entities beside each suggestion
- let the user inject a suggestion into the draft without auto-send

## Follow-On Phases

Natural next steps after this phase:

- DM compose
- streaming compose
- persisted compose history
- intent-specific suggestion templates
- channel auto-summarize and auto-regen scheduling
