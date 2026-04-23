# Phase 63E Entity Ask Stream And History Design

## Goal

Turn entity Ask AI from a one-shot response into a streamable, persistent knowledge interaction surface.

## Scope

- Keep existing `POST /api/v1/knowledge/entities/:id/ask` compatible.
- Add `POST /api/v1/knowledge/entities/:id/ask/stream` with SSE events.
- Add `GET /api/v1/knowledge/entities/:id/ask/history` for current-user Q&A history.
- Persist both sync and streaming entity Ask answers.

## Architecture

Reuse the existing entity ask prompt builder so sync and streaming answers share the same grounding, citations, provider, and model behavior. Add a small `KnowledgeEntityAskAnswer` domain model with entity/user/question/answer/reasoning/provider/model/citation-count timestamps. The persisted row is an interaction history record, not a cache, so every ask creates a new row.

The streaming endpoint follows the composer stream pattern: emit `start`, one or more `answer.delta`, `answer.done`, and `done`. If the upstream fails after headers, emit `error`. On successful completion, persist the full answer and include it in `answer.done` with citations.

History is scoped to the current user and entity, ordered newest first, and capped by `limit` with a safe default. It returns rows plus hydrated entity context so the UI can restore an Ask AI panel after refresh without invoking the LLM.

## Error Handling

- Missing AI gateway: `503`.
- Empty question: `400`.
- Unknown entity: `404`.
- Stream unsupported: `500`.
- Upstream stream errors before headers: standard JSON error.
- Upstream stream errors after headers: SSE `error`.

## Testing

- Sync ask persists a history row with citations count.
- History returns current-user rows for the entity in newest-first order.
- Stream ask emits `start`, `answer.delta`, `answer.done`, and `done`.
- Stream ask persists the final answer.
- Missing gateway and empty question remain covered by handler behavior.
