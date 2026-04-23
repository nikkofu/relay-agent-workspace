# Phase 63D AI Compose DM And Intent Expansion Design

## Goal

Extend the AI composer from channel/thread-only reply suggestions into a shared composer substrate for channel, thread, and DM scopes, while adding intent variants and feedback aggregation for Windsurf's next UI pass.

## Scope

- `POST /api/v1/ai/compose` accepts either `channel_id` or `dm_id`.
- `POST /api/v1/ai/compose/stream` accepts either `channel_id` or `dm_id`.
- Supported intents become `reply`, `summarize`, `followup`, and `schedule`.
- `POST /api/v1/ai/compose/:id/feedback` accepts channel or DM scoped feedback.
- `GET /api/v1/ai/compose/:id/feedback/summary` returns aggregate feedback counts and recent rows.

## Architecture

Keep the existing compose flow intact and introduce scope-aware preparation helpers. Channel scopes keep using message, thread, and knowledge context. DM scopes load recent `DMMessage` rows and build a private-message prompt without channel knowledge refs. The response shape stays backward-compatible: existing `channel_id`, `thread_id`, `suggestions[]`, `citations[]`, `context_entities[]`, `provider`, and `model` remain, and `dm_id` is added when the scope is a DM.

Intent handling is deliberately prompt-level and fallback-level in this phase. The LLM still returns the same structured JSON, but the prompt instructions and deterministic fallback text change by intent. This keeps Windsurf's existing suggestion card renderer reusable and avoids separate UI contracts per intent.

Feedback summary reads the existing `AIComposeFeedback` table. It is scoped by `compose_id` and returns totals for `up`, `down`, and `edited`, plus recent feedback metadata. This provides a learning-signal API without introducing ranking changes yet.

## Error Handling

- Missing both `channel_id` and `dm_id`: `400`.
- Supplying both `channel_id` and `dm_id`: `400`.
- Unsupported intent: `400`.
- Unknown channel/thread/DM: `404`.
- Missing AI gateway: `503`.
- Upstream stream failures after SSE headers: `event: error`.

## Testing

Use handler-level tests to lock the public contracts:

- DM compose returns `dm_id`, intent, suggestions, provider, and model.
- Stream start/done payloads include `dm_id` and the requested intent.
- Each supported intent is accepted and unsupported intents are rejected.
- DM feedback persists with `dm_id`.
- Feedback summary aggregates `up`, `down`, and `edited` counts.
