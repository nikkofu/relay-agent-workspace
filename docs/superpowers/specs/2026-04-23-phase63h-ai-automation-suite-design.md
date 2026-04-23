# Phase 63H AI Automation Suite Design

## Goal

Deliver the next AI-native backend wave for Relay by adding three connected capabilities in one coordinated phase:

1. entity brief auto-regeneration
2. schedule booking from structured compose slots
3. compose activity digest aggregation

The goal is not to build a heavyweight workflow platform. The goal is to make Relay's AI behavior observable, recoverable, and automatable in the same way Slack surfaces are observable and durable.

## Why This Phase Exists

Phase 63E through 63G made three things true:

- entity Ask AI is now streamable and persistent
- channel summaries and compose suggestions are now websocket-visible workspace events
- compose activity is now durable and refreshable

The remaining gap is that some of the most valuable AI behavior is still either manual or read-only:

- stale entity briefs still require an explicit user click
- schedule intent can suggest slots but cannot actually create a booking artifact/result
- compose activity can be listed, but not summarized into higher-level operational signals

This phase closes those gaps with a light automation layer, not a full general-purpose orchestration framework.

## Architecture

### 1. Lightweight AI Automation Layer

Add a small automation seam that sits between handlers and existing domain services.

Responsibilities:

- enqueue and dedupe background work
- track job state
- expose minimal retry/status APIs
- emit workspace-visible realtime events

This layer should remain narrow and specific to AI-native collaboration tasks. It should not replace the existing workflow/tool system.

### 2. Keep Existing Read Models

Do not create parallel summary or digest systems where one already exists.

- Entity briefs continue to persist into `AISummary` with `scope_type=knowledge_entity`
- Channel summaries continue to persist into `AISummary` with `scope_type=channel`
- Compose activity remains the event source for compose analytics

The new phase adds orchestration and aggregation on top of those existing models.

### 3. Persist Actions, Not Just Results

Each major automation path should have a durable state model:

- brief regeneration job state
- schedule booking state
- compose digest response derived from persisted compose activity

This keeps the product restart-safe, multi-tab safe, and suitable for later agent collaboration features.

## Capability A: Entity Brief Auto-Regeneration

### Product Behavior

When `knowledge.entity.brief.changed` is emitted because an entity changed, received a new ref, or got a new timeline event:

- the entity should be marked stale as today
- the backend should enqueue a regeneration job
- duplicate triggers in a short window should collapse into one pending job
- the worker should eventually regenerate the brief
- success should continue to emit `knowledge.entity.brief.generated`
- failure should be visible and retryable

First version should support both:

- event-driven regeneration from `knowledge.entity.brief.changed`
- periodic stale-entity sweep for entities that were missed or remained stale too long

### Persistence Model

Add `AIAutomationJob`:

- `id`
- `job_type` (`entity_brief_regen`)
- `scope_type` (`knowledge_entity`)
- `scope_id`
- `workspace_id`
- `status` (`pending | running | succeeded | failed | cancelled`)
- `trigger_reason`
- `dedupe_key`
- `attempt_count`
- `last_error`
- `scheduled_at`
- `started_at`
- `finished_at`
- `created_at`
- `updated_at`

Add a unique index on:

- `job_type`
- `scope_type`
- `scope_id`
- `status`

with the implementation rule that only one `pending` or one `running` brief-regeneration job may exist per entity at a time. If the current database/index strategy cannot express that partial uniqueness directly in SQLite through GORM, the service layer must enforce it inside a transaction:

1. query for existing `pending/running` job for the entity
2. if found, return that job instead of creating another
3. otherwise insert a new `pending` job

Also add a non-unique lookup index on:

- `scope_type`
- `scope_id`
- `created_at`

### APIs

- `GET /api/v1/knowledge/entities/:id/brief/automation`
- `POST /api/v1/knowledge/entities/:id/brief/automation/run`
- `POST /api/v1/knowledge/entities/:id/brief/automation/retry`

Handler semantics:

- `POST /run`
  - if a `pending` or `running` job already exists, return `200` with the existing job and `created=false`
  - otherwise create a new `pending` job and return `202` with `created=true`
- `POST /retry`
  - only allowed when the latest job is `failed` or `succeeded`
  - if a `pending` or `running` job already exists, return that job without creating another
  - retry creates a fresh job row instead of mutating history in place

### Realtime Events

- `knowledge.entity.brief.regen.queued`
- `knowledge.entity.brief.regen.started`
- `knowledge.entity.brief.generated` (existing success path remains authoritative)
- `knowledge.entity.brief.regen.failed`

### Design Constraints

- never delete or blank an existing brief on failure
- dedupe by entity and time bucket so bursts do not flood the system
- dedupe key should be deterministic: `entity_brief_regen:<entity_id>:<stale_epoch_bucket>`
- first version uses a 2-minute stale epoch bucket for event-triggered dedupe
- periodic sweep should only pick stale entities without an active pending/running job
- implementation should reuse current prompt-builder and summary-persistence path instead of duplicating brief generation logic

## Capability B: Schedule Booking

### Product Behavior

Users can choose one of the `compose.proposed_slots[]` options and turn it into a durable booking result.

First version should do both:

- internal Relay booking record + ICS output as the default success path
- adapter shape for future/optional external providers such as Google Calendar or Outlook

This means the system is useful immediately even before external provider integrations are complete.

### Persistence Model

Add `AIScheduleBooking`:

- `id`
- `workspace_id`
- `channel_id`
- `dm_id`
- `requested_by`
- `intent_source_compose_id`
- `title`
- `description`
- `starts_at`
- `ends_at`
- `timezone`
- `attendee_ids_json`
- `provider` (`internal | google | outlook | open`)
- `status` (`draft | booked | failed | cancelled`)
- `external_ref`
- `ics_content`
- `last_error`
- `created_at`
- `updated_at`

Rationale:

- first version stores ICS inline to avoid introducing file-storage coupling or local path portability issues
- external calendar deep links may be returned in the response, but ICS remains the durable fallback artifact

### APIs

- `POST /api/v1/ai/schedule/book`
- `GET /api/v1/ai/schedule/bookings`
- `GET /api/v1/ai/schedule/bookings/:id`
- `POST /api/v1/ai/schedule/bookings/:id/cancel`

`POST /api/v1/ai/schedule/book` request contract:

- exactly one of `channel_id` or `dm_id` is required
- `compose_id` is required
- `slot` is required and must include:
  - `starts_at`
  - `ends_at`
  - `timezone`
  - `attendee_ids`
- optional:
  - `title`
  - `description`
  - `provider`

Provider behavior in first version:

- `provider=internal`
  - always supported
  - persists booking
  - generates `ics_content`
  - returns optional `ics_download_url` if the API later exposes a download route
- `provider=google|outlook|open`
  - accepted as adapter-ready values
  - if no adapter is configured, booking is still persisted with `status=booked` and `external_ref=""`
  - response marks `delivery.mode="ics_fallback"`
  - no external sync attempt is required in first version

### Realtime Events

- `schedule.event.booked`
- `schedule.event.updated`
- `schedule.event.cancelled`

### Design Constraints

- validate scope exactly as compose does: channel or DM, not both
- booking handler must reject both-empty and both-present scope combinations with `400`
- keep a booking row even if an external provider call fails
- return usable internal fallback output such as ICS even when provider sync fails
- preserve `compose_id` linkage so activity, schedule, and message drafting can be correlated later

## Capability C: Compose Activity Digest

### Product Behavior

Relay should expose aggregate co-drafting signals, not just a raw list of compose activity rows.

The digest should support:

- workspace view
- channel view
- DM-compatible scope fields, even if DM UI follows later

The primary intended consumers are:

- `#agent-collab`
- workspace activity surfaces
- channel info / insight panes

### Model Change

Extend `AIComposeActivity` with:

- `user_id`

Migration rule:

- `user_id` is nullable for historical rows
- add an index on `user_id, created_at`
- all new rows must populate `user_id`
- digest responses must count historical null-user rows in totals, but exclude them from `unique_users` and place them under `"unknown"` only when a `group_by=user` breakdown is requested

Without this field, the activity stream shows what happened but not who initiated it, which limits its value as a collaboration signal.

### API

- `GET /api/v1/ai/compose/activity/digest`

Supported query dimensions:

- `workspace_id`
- `channel_id`
- `dm_id`
- `window` (`1h | 24h | 7d | custom`)
- `intent`
- `group_by` (`intent | channel | user | provider | model`)

Scope precedence:

1. `dm_id` or `channel_id` may be provided, but not both
2. if `channel_id` or `dm_id` is provided, `workspace_id` is optional and acts as an additional safety filter
3. if only `workspace_id` is provided, digest is workspace-wide

`window=custom` requires:

- `start_at`
- `end_at`

and must reject:

- missing `start_at` or `end_at`
- `end_at <= start_at`
- windows longer than 30 days in first version

### Response Shape

Return a compact aggregation payload:

- `summary`
  - `total_requests`
  - `total_suggestions`
  - `unique_users`
  - `top_intent`
- `breakdown`
- `top_channels`
- `top_users`
- `timeline`

When `group_by=user` is requested, response rows should expose:

- `user_id`
- `request_count`
- `suggestion_count`

and historical rows with null `user_id` should be grouped under:

- `user_id=""`
- `label="unknown"`

### Design Constraints

- first version should query directly from persisted compose activity
- no snapshot/materialized digest table is needed yet
- enforce small/default windows and capped limits to avoid runaway scans

## Error Handling

### Brief Auto-Regeneration

- LLM/provider failures mark job `failed`
- old brief remains available
- retry is explicit
- repeated upstream failures must not create duplicate pending jobs
- `/run` and `/retry` must be concurrency-safe under duplicate HTTP submissions

### Schedule Booking

- invalid scope or invalid slot returns `400`
- provider failure marks booking `failed`
- internal booking record remains queryable
- ICS/internal result should remain available when possible
- `/cancel` on an already-cancelled booking should be idempotent and return the existing row

### Compose Digest

- aggregation failure should not affect compose write-paths
- query validation should reject nonsensical filters early
- historical null-user rows must not break `group_by=user`

## Testing Strategy

### Brief Auto-Regeneration

- enqueue on brief change
- dedupe repeated enqueue attempts
- worker success updates brief and emits realtime events
- worker failure persists error state
- stale sweep only picks eligible entities
- concurrent enqueue attempts return the same pending/running job instead of inserting duplicates
- `/run` and `/retry` semantics are covered explicitly
- stale sweep does not requeue entities with active pending/running jobs

### Schedule Booking

- successful booking persists row and returns ICS metadata
- invalid slot/scope is rejected
- failed external provider keeps row with `failed` status
- realtime booking event emitted on success
- invalid scope combinations are rejected
- cancel is idempotent
- internal provider persists usable `ics_content`

### Compose Digest

- `user_id` persists on compose activity rows
- historical rows with null `user_id` aggregate safely
- digest aggregates by workspace/channel correctly
- breakdowns by intent/user/provider/model are correct
- window and limit guards are enforced
- `window=custom` validation is covered

## Delivery Plan

Recommended implementation order:

1. add `user_id` to `AIComposeActivity` and ship `GET /api/v1/ai/compose/activity/digest`
2. add entity brief auto-regeneration job model, worker, APIs, and realtime events
3. add schedule booking APIs, persistence, ICS generation, and realtime events

This order balances risk and product value:

- digest is the lowest-risk extension to current work
- brief auto-regeneration is the highest-value always-on knowledge improvement
- schedule booking is the most externally extensible path and benefits from landing after the automation seam exists
