# Phase 64C Unified Activity Feed Completion Design

## Goal

Deliver the next Slack-like backend convergence slice for Relay by completing the unified activity feed and tightening the persistence behind the newly-added activity event types.

This phase intentionally stays narrow:

1. complete the feed coverage that Windsurf already exposed in the UI
2. harden only the persistence and linking behavior required for those feed events
3. avoid opening unrelated new product surfaces in the same release

## Why This Phase Exists

Phase 64A and 64B established the feed foundation:

- Windsurf shipped the shared activity rail and started consuming one backend feed contract
- Codex shipped `GET /api/v1/activity/feed` with the initial six event types

The remaining gap is that the feed is still missing several core Slack-like activity categories that already exist elsewhere in the product model:

- artifact edits do not appear in the shared timeline
- workflow/tool executions are still isolated from the workspace activity model
- DM activity and thread replies are not represented as first-class feed events
- mentions and reactions are durable elsewhere, but not surfaced consistently in the unified feed

Without these event types, the activity rail remains incomplete and several surfaces still depend on fragmented or local fallback state.

## Scope

Phase 64C is a combined release with two tightly-bound layers.

### Layer 1: Feed Completion

Extend `GET /api/v1/activity/feed` with six new event types:

- `artifact_updated`
- `tool_run`
- `reply`
- `dm_message`
- `mention`
- `reaction`

### Layer 2: Slack Persistence

Tighten the persistence and mapping needed for those feed items so they survive refresh, process restart, and cross-tab usage:

- stable actor identity
- stable channel / DM / entity linking
- stable timestamps based on real persisted data
- minimal metadata for downstream UI rendering

## Explicitly Out Of Scope

This phase does not open new top-level product areas beyond what is needed for the feed:

- no new file archive UX contract
- no new user-group or profile APIs
- no new workspace home redesign APIs
- no generic event bus redesign
- no new mention-fanout storage model unless existing persistence proves unusable

Those remain for later Phase 64 follow-up work.

## Architecture

### 1. Keep The Existing Unified Feed Contract

Do not change the response shape already consumed by Windsurf.

`GET /api/v1/activity/feed` continues returning:

- `items`
- `next_cursor`
- `total`

And each item continues using the existing `UnifiedActivityFeedItem` shape:

- `id`
- `event_type`
- `workspace_id`
- `actor_id`
- `actor_name`
- `channel_id`
- `channel_name`
- `dm_id`
- `entity_id`
- `entity_title`
- `entity_kind`
- `title`
- `body`
- `link`
- `occurred_at`
- `meta`

This phase adds event sources and mapping rules, not a contract change.

### 2. Reuse Existing Persistence First

Use existing durable models and only add persistence where the current model is too weak to map a stable feed item.

Expected source families:

- artifacts / artifact versions
- workflow runs or tool runs
- channel messages
- DM messages
- inbox / mentions / message metadata
- message reactions

### 3. Single Feed Item Per Source Record

Each persisted source record should map to one primary feed event in this phase.

Examples:

- a thread message becomes `reply`, not also `message`
- a DM message becomes `dm_message`, not also `message`
- a reaction record becomes `reaction`, not also a generic message mutation

This avoids double-counting and keeps filter semantics clear.

## Event Type Design

## A. `artifact_updated`

### Source

Artifact persistence and artifact version restore/update history.

### Product Meaning

An artifact or canvas document changed in a durable way and should appear in the workspace timeline.

### Feed Mapping

- `event_type`: `artifact_updated`
- `workspace_id`: artifact workspace
- `channel_id`: artifact channel when present
- `actor_id` / `actor_name`: editor or restoring user when known
- `title`: concise artifact action title
- `body`: optional excerpt or artifact kind summary
- `link`: stable artifact/canvas route
- `occurred_at`: artifact `updated_at` or version restore timestamp
- `meta`:
  - `artifact_id`
  - `artifact_type`
  - `version_id` when applicable

### Constraints

- use real persisted timestamps, not seeded or hardcoded placeholders
- if linked channel metadata is missing, still return the feed item with the artifact link

## B. `tool_run`

### Source

Existing workflow run / tool run persistence.

### Product Meaning

A workspace automation or tool execution happened and should be observable from the shared activity rail.

### Feed Mapping

- `event_type`: `tool_run`
- `workspace_id`: run workspace
- `actor_id` / `actor_name`: initiating user when known
- `channel_id`: optional related channel
- `title`: tool/workflow name + outcome summary
- `body`: optional status or scope summary
- `link`: workflows page or run detail route if already available
- `occurred_at`: run `updated_at` or terminal timestamp, otherwise `created_at`
- `meta`:
  - `run_id`
  - `tool_name`
  - `status`

### Constraints

- unknown status is allowed and should map to `status=unknown`
- tool runs should remain visible even if a linked channel or actor cannot be resolved

## C. `reply`

### Source

Channel messages with non-empty `thread_id`.

### Product Meaning

A thread reply happened inside a channel conversation.

### Feed Mapping

- `event_type`: `reply`
- `workspace_id`: message workspace
- `channel_id` / `channel_name`: parent channel
- `actor_id` / `actor_name`: sender
- `title`: sender + thread reply wording
- `body`: message content snippet
- `link`: channel route with thread anchor/deep link
- `occurred_at`: message `created_at`
- `meta`:
  - `message_id`
  - `thread_id`

### Constraints

- thread replies must not also appear as generic `message` rows
- channel-scoped filtering must include replies for that channel

## D. `dm_message`

### Source

Persisted DM messages.

### Product Meaning

A direct message happened and should show in the shared activity stream where DM-aware filtering is enabled.

### Feed Mapping

- `event_type`: `dm_message`
- `workspace_id`: conversation workspace
- `dm_id`: DM conversation id
- `actor_id` / `actor_name`: sender
- `title`: sender + DM wording
- `body`: message content snippet
- `link`: DM deep link
- `occurred_at`: message `created_at`
- `meta`:
  - `message_id`

### Constraints

- DM messages must not appear as generic channel `message` rows
- DM filter must return only `dm_message` items for the requested conversation unless another explicit `event_type` is requested

## E. `mention`

### Source

Existing durable mention-related data, using the lightest reliable path available in the codebase:

- inbox / mentions items
- message metadata
- group mention metadata
- entity mention metadata

### Product Meaning

A user-relevant mention happened and should be visible as a first-class activity event.

### Feed Mapping

- `event_type`: `mention`
- `workspace_id`: mention workspace
- `channel_id` or `dm_id`: where the mention happened
- `actor_id` / `actor_name`: author of the mentioning message
- `title`: mention summary
- `body`: message content snippet
- `link`: message deep link
- `occurred_at`: mention creation or source message time
- `meta`:
  - `message_id`
  - `mention_kind` (`user | group | entity`)

### Constraints

- phase 64C does not require a new heavy mention event table
- only deterministic, already-persisted mentions should be mapped
- if a mention cannot be tied to a stable source message, skip it rather than inventing a feed row

## F. `reaction`

### Source

Persisted message reactions.

### Product Meaning

Someone reacted to a persisted message and that action should be visible in the shared activity stream.

### Feed Mapping

- `event_type`: `reaction`
- `workspace_id`: message workspace
- `channel_id` or `dm_id`: owning conversation
- `actor_id` / `actor_name`: user who reacted when available
- `title`: reaction summary
- `body`: optional target message snippet
- `link`: target message deep link
- `occurred_at`: reaction timestamp or updated timestamp
- `meta`:
  - `emoji`
  - `message_id`

### Constraints

- if the target message no longer exists, skip the reaction item
- reactions should be filtered by their owning scope, not by ad-hoc store state

## API Behavior

`GET /api/v1/activity/feed` keeps the current parameters:

- `workspace_id` required
- `channel_id` optional
- `dm_id` optional
- `actor_id` optional
- `event_type` optional
- `limit` optional
- `cursor` optional

Behavior rules:

1. `workspace_id` remains required
2. `event_type` may target any existing or newly-added feed event type
3. sorting remains `occurred_at desc`
4. `next_cursor` remains based on the last returned `occurred_at`
5. a single bad source row must not fail the entire feed

## Metadata Rules

Keep `meta` intentionally light:

- `artifact_updated`
  - `artifact_id`
  - `artifact_type`
  - `version_id`
- `tool_run`
  - `run_id`
  - `tool_name`
  - `status`
- `reply`
  - `message_id`
  - `thread_id`
- `dm_message`
  - `message_id`
- `mention`
  - `message_id`
  - `mention_kind`
- `reaction`
  - `message_id`
  - `emoji`

## Error Handling

The feed should degrade per item, not per request.

- if actor lookup fails, return the item without `actor_name`
- if channel lookup fails, return the item without `channel_name`
- if artifact or workflow linkage is partial, return the item with the best stable link available
- if a reaction cannot resolve its target message, drop that item
- if a mention cannot resolve a stable source message, drop that item

## Testing Strategy

Phase 64C should land with contract tests in three slices.

### Slice A

- `artifact_updated`
- `tool_run`
- event type filtering for those rows

### Slice B

- `reply`
- `dm_message`
- correct links and scope mapping

### Slice C

- `mention`
- `reaction`
- dedupe and skip-invalid-row behavior

Regression coverage must confirm:

- existing six feed event types still work
- sorting and cursor behavior remain stable
- the handler never returns blank `event_type`, `title`, or `occurred_at`

## Release Strategy

Ship as one unified Phase 64C release, but implement in three internal cuts:

1. `64C-A`: `artifact_updated` + `tool_run`
2. `64C-B`: `reply` + `dm_message`
3. `64C-C`: `mention` + `reaction`

This gives Windsurf stable intermediate contracts while keeping the public release story coherent.

## Next-Round Record

The user explicitly chose option `1` for this round and wants the other two larger directions remembered for later rounds.

That means:

- this round completes `Feed Completion + Slack Persistence`
- next-round continuation `2` means keep strengthening the unified feed without opening unrelated product areas
- next-round continuation `3` means broaden into larger Slack-like capability work: files, groups, home, status, profiles, workflows/tools, and adjacent persistence cleanup
