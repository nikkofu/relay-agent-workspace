# Phase 65A User Mention Semantics Design

## Goal

Deliver a durable, queryable, and unified user-mention backend model for Relay so that message hydration, unified activity feed, mentions surfaces, and inbox semantics stop depending on ad-hoc text scanning.

This phase is the first Phase 65 slice and is intentionally narrow:

1. persist explicit `@user` mentions when messages are created
2. make user mentions first-class feed items in the unified activity stream
3. align `GET /api/v1/mentions` and the mention portion of `GET /api/v1/inbox` to the same persisted semantics

## Why This Phase Exists

Phase 64C completed the unified activity feed but deliberately used the most deterministic persisted mention source already available: `message.metadata.entity_mentions`.

That solved the feed shape problem, but it did not solve user mentions as a durable product primitive:

- current mention logic in legacy activity surfaces still scans message text against the current user's display name
- this is vulnerable to renames, collisions, and false positives
- activity feed, mentions, and inbox do not share a single persisted user-mention source
- Windsurf's latest `v0.6.35` UI pass explicitly calls for real `@user` mention events in the unified feed

Phase 65A closes that gap by introducing a lightweight mention index and using it consistently across the relevant APIs.

## Scope

Phase 65A covers one product seam: **user mention semantics**.

### Included

- parse explicit `@user` mentions on channel and DM message creation
- persist mention render data into `message.metadata.user_mentions[]`
- persist a lightweight query index in a new `MessageMention` table
- make unified activity feed `mention` rows come from persisted user mentions
- align `GET /api/v1/mentions` to the new mention source
- align the mention branch of `GET /api/v1/inbox` to the same source
- emit realtime `mention.created`

### Excluded

- `@group` mention persistence
- historical backfill over all old messages
- mention mute/read-state redesign
- proactive AI nudges based on mentions
- generalized notification routing overhaul

Those remain later follow-up work.

## Architecture

### 1. Dual Persistence: Metadata + Index

Use two coordinated representations for user mentions:

1. `message.metadata.user_mentions[]`
   - optimized for message hydration and rendering
2. `MessageMention`
   - optimized for query, filtering, pagination, and later notification/read-state expansion

This avoids repeated JSON scans for mention-centric APIs while keeping message payloads self-describing for the frontend.

### 2. Parse Only Explicit Mentions

Phase 65A only recognizes deliberate `@Name` mentions that resolve to a real workspace user.

Do not use fuzzy natural-language matching.
Do not recreate the legacy "substring of current user's display name" behavior.

### 3. Keep Message Creation Resilient

Message creation remains the primary user action. Mention parsing and indexing should enrich the message, not endanger delivery.

If mention persistence fails:

- the message still succeeds
- mention metadata/index rows may be partially absent
- the backend should log the failure
- the API must not invent successful mention results it did not persist

## Data Model

Add `MessageMention`:

- `id`
- `message_id`
- `workspace_id`
- `channel_id`
- `dm_id`
- `mentioned_user_id`
- `mentioned_by_user_id`
- `mention_text`
- `mention_kind` (`user`)
- `created_at`

Recommended indexes:

- `mentioned_user_id, created_at`
- `workspace_id, created_at`
- `channel_id, created_at`
- `dm_id, created_at`
- unique-ish protection on `message_id + mentioned_user_id + mention_text` to avoid duplicate rows from the same parsed mention

`message.metadata.user_mentions[]` minimum shape:

- `user_id`
- `name`
- `mention_text`

This is separate from existing `entity_mentions[]` and should coexist in the same metadata payload.

## Parsing Rules

Phase 65A parsing behavior:

- only parse explicit `@Name`
- only match users within the message scope workspace
- longest name wins to avoid `@Ann` matching `@Anna`
- preserve the original matched text as `mention_text`
- dedupe repeated identical mentions to the same user within one message

Scope behavior:

- channel message:
  - `workspace_id` comes from the channel
  - `channel_id` is set
  - `dm_id` is empty
- DM message:
  - `dm_id` is set
  - `workspace_id` should be derived from the participating users' organization/workspace context already used by the DM UI
  - `channel_id` is empty

Self mention behavior:

- may be stored in metadata and index for fidelity
- should not appear in that same user's mentions feed or mention-based inbox branch

## Message Creation Flow

For channel message create:

1. persist the message
2. derive workspace scope
3. parse explicit user mentions from the final message content
4. update `message.metadata.user_mentions[]`
5. insert `MessageMention` rows
6. broadcast existing message realtime event
7. broadcast `mention.created` for each persisted mention

For DM message create:

1. persist the DM message
2. derive workspace scope for the DM participants
3. parse explicit user mentions from the final message content
4. update any DM message metadata representation used by hydration
5. insert `MessageMention` rows with `dm_id`
6. broadcast existing DM realtime event
7. broadcast `mention.created`

Phase 65A may keep DM mention hydration lighter than channel message hydration if the current DM message model does not already persist a metadata column. In that case:

- index persistence is still required
- feed/mentions/inbox semantics should still work
- message-level `user_mentions[]` hydration for DMs can be documented as partial if needed

## API Behavior

## A. Message Create APIs

Channel and DM message create handlers must:

- parse user mentions on write
- persist mention rows
- persist renderable user mention metadata when supported by the message model

## B. `GET /api/v1/activity/feed`

`mention` items should now be backed by `MessageMention`, not by ad-hoc user-name scanning.

Mapping:

- `event_type`: `mention`
- `workspace_id`
- `actor_id` / `actor_name`: `mentioned_by_user_id`
- `channel_id` or `dm_id`
- `title`: mention summary
- `body`: source message snippet
- `link`: message deep link
- `occurred_at`: mention `created_at`
- `meta`:
  - `message_id`
  - `mention_kind=user`
  - `mentioned_user_id`

Entity mentions remain valid feed rows from their current source. This phase requires `mention_kind` to differentiate:

- `user`
- `entity`

## C. `GET /api/v1/mentions`

This endpoint should stop using legacy display-name scanning.

It should return rows derived from `MessageMention` where:

- `mentioned_user_id = current_user`
- exclude self-mentions from the current user
- support stable newest-first ordering by mention creation time

## D. `GET /api/v1/inbox`

This phase does not require a total inbox rewrite.

It does require that the mention branch inside inbox use the same persisted `MessageMention` semantics as:

- `GET /api/v1/activity/feed`
- `GET /api/v1/mentions`

Other inbox categories may continue using existing sources in this phase.

## Realtime

Add websocket event:

- `mention.created`

Payload should be close to a unified feed row so Windsurf can reuse it with minimal transform:

- mention identity
- message id
- workspace / channel / dm scope
- actor metadata
- mentioned user id
- title/body/link
- occurred_at
- `mention_kind=user`

## Error Handling

Mention enrichment must not break message delivery.

- if parsing fails, create the message and skip mention persistence
- if metadata update fails after message creation, keep the message and log
- if mention index insertion fails, keep the message and log
- if one mention target is invalid, skip that target and continue processing other valid ones

The system should prefer missing mention enrichment over incorrect mention data.

## Testing Strategy

Phase 65A should land with three test groups.

### 1. Persistence Tests

- channel message creation writes `message.metadata.user_mentions[]`
- channel message creation writes `MessageMention`
- DM message creation writes `MessageMention`
- duplicate mention targets within one message are deduped

### 2. Query Tests

- `GET /api/v1/mentions` returns only the current user's persisted mentions
- self-mentions are excluded from the current user's mentions feed
- `GET /api/v1/activity/feed` returns `mention` with `mention_kind=user`
- entity mentions still remain available as `mention_kind=entity`

### 3. Inbox/Realtime Tests

- mention branch in `GET /api/v1/inbox` aligns with `MessageMention`
- `mention.created` broadcasts on successful persistence
- message creation still succeeds when mention persistence is intentionally made to fail

## Release Strategy

Ship as one coherent Phase 65A release, but implement in three internal cuts:

1. `65A-A`: mention persistence on message create
2. `65A-B`: unified feed + mentions endpoint migration
3. `65A-C`: inbox mention alignment + realtime + docs

## Windsurf Handoff Target

After this phase, Windsurf should be able to:

- render true `@user` mention rows from the unified feed
- distinguish `mention_kind=user` and `mention_kind=entity`
- use `mention.created` to append mention activity live
- trust `GET /api/v1/mentions` as a real persisted surface instead of legacy text heuristics
