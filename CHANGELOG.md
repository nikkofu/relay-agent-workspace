# Changelog

All notable changes to Relay Agent Workspace are documented in this file.

## [0.6.32] - 2026-04-23

This release completes Phase 64B backend support for Windsurf's unified activity rail.

### Added

- `GET /api/v1/activity/feed`
- Contract now matches `UnifiedActivityFeedItem` with:
  - `event_type`
  - `workspace_id`
  - `actor_id` / `actor_name`
  - `channel_id` / `channel_name`
  - `dm_id`
  - `entity_id` / `entity_title` / `entity_kind`
  - `title`, `body`, `link`, `occurred_at`, `meta`
- Current feed sources:
  - `message`
  - `file_uploaded`
  - `schedule_booking`
  - `compose_activity`
  - `knowledge_ask`
  - `automation_job`
- Cursor pagination via `next_cursor`, plus aggregated `total`.

### Windsurf Handoff

- The `UnifiedActivityRail` All tab can now consume `GET /api/v1/activity/feed?workspace_id=...`.
- Activity rows now have the entity/channel/actor fields Windsurf requested in `docs/AGENT-COLLAB.md`.
- Next recommended backend slice is Phase 64C: widen the unified feed with `artifact_updated`, `tool_run`, DM activity, and richer message/reaction variants.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase64BUnifiedActivityFeedContract -count=1`
- `go test ./internal/handlers -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`
- `git diff --check`

## [0.6.30] - 2026-04-23

This release closes Phase 63 as an AI automation arc and prepares Phase 64 for broader Slack-core convergence.

### Fixed

- `POST /api/v1/knowledge/entities` now accepts UI-created entities that omit `workspace_id` when a primary workspace exists.
- `tags[]` on knowledge entity creation is preserved into `metadata_json`, fixing file/wiki creation flows such as `{ "title": "Principles of Game Design", "kind": "file", "tags": [...] }`.

### Improved

- `GET /api/v1/knowledge/ask/recent` now returns denormalized `entity_title` and `entity_kind` on each item, so Windsurf can render shared ask feed rows without falling back to truncated entity IDs.
- Added Phase 64 planning document for Slack Core Convergence and AI-native workspace observability.

### Verification Used For This Release

- `go test ./internal/handlers -run 'Test(KnowledgeEntityCRUDEndpoints|Phase63IEntityAskRecentFeedAndRealtime)' -count=1`
- Full release verification listed in `docs/releases/v0.6.30.md`.

## [0.6.28] - 2026-04-23

This release implements Codex Phase 63I, closing the next always-on AI loops after Windsurf's `v0.6.27` UI pass.

### Added

- **Auto-summarize scheduler worker**:
  - Added background channel auto-summary execution on API startup and every minute.
  - Respects persisted `ChannelAutoSummarySetting` rows, including `is_enabled`, `min_new_messages`, `message_limit`, `provider`, and `model`.
  - Writes refreshed channel `AISummary` cache and updates `last_run_at` / `last_message_at`.
  - Broadcasts websocket `channel.summary.updated` with `reason=auto_run`.
- **Knowledge ask rolling feed**:
  - `GET /api/v1/knowledge/ask/recent`
  - Supports `workspace_id` plus optional `entity_id`, `user_id`, and `limit`.
  - Returns newest-first persisted `KnowledgeEntityAskAnswer` rows for shared workspace visibility.
  - `POST /api/v1/knowledge/entities/:id/ask` and `POST /api/v1/knowledge/entities/:id/ask/stream` now emit websocket `knowledge.entity.ask.answered`.
- **Workspace automation audit view API**:
  - `GET /api/v1/ai/automation/jobs`
  - Supports `workspace_id`, `status`, `job_type`, `scope_type`, `scope_id`, and `limit`.
  - Returns `{ items, total }` for workspace-level monitoring of durable `AIAutomationJob` rows.

### Windsurf Handoff

- Add an automation audit panel using `GET /api/v1/ai/automation/jobs?workspace_id=...&limit=20`, with status filters for `pending/running/failed/succeeded`.
- Surface shared entity Ask activity in Following Hub, `#agent-collab`, or workspace Activity by hydrating `GET /api/v1/knowledge/ask/recent?workspace_id=...` and listening for `knowledge.entity.ask.answered`.
- Reuse the existing channel auto-summary UI, but expect background `channel.summary.updated` events with `reason=auto_run` even when the user does not click Run now.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase63I -count=1`
- `go test ./internal/handlers -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`
- `git diff --check`

## [0.6.26] - 2026-04-23

This release implements Codex Phase 63H, turning the AI compose and knowledge layer from refreshable signals into a more complete automation suite.

### Added

- **Compose activity digest API**:
  - `GET /api/v1/ai/compose/activity/digest`
  - Supports `workspace_id`, `channel_id`, `dm_id`, `window`, `start_at`, `end_at`, `intent`, `group_by`, and `limit`.
  - New `AIComposeActivity.user_id` persistence for fresh rows.
  - Historical null-user rows stay counted in totals, map to `unknown` only for `group_by=user`, and are excluded from `summary.unique_users`.
- **Entity brief automation APIs**:
  - `GET /api/v1/knowledge/entities/:id/brief/automation`
  - `POST /api/v1/knowledge/entities/:id/brief/automation/run`
  - `POST /api/v1/knowledge/entities/:id/brief/automation/retry`
  - Added durable `AIAutomationJob` rows for `entity_brief_regen`.
  - `knowledge.entity.brief.changed` now queues background regeneration work.
  - Added websocket events:
    - `knowledge.entity.brief.regen.queued`
    - `knowledge.entity.brief.regen.started`
    - `knowledge.entity.brief.regen.failed`
  - Added API-process scheduler hooks for stale sweep + pending job processing.
- **Schedule booking APIs**:
  - `POST /api/v1/ai/schedule/book`
  - `GET /api/v1/ai/schedule/bookings`
  - `GET /api/v1/ai/schedule/bookings/:id`
  - `POST /api/v1/ai/schedule/bookings/:id/cancel`
  - Added durable `AIScheduleBooking` rows with `requested_by`, inline `ics_content`, provider/status fields, and idempotent cancel semantics.
  - Added websocket events:
    - `schedule.event.booked`
    - `schedule.event.cancelled`

### Windsurf Handoff

- Add a digest strip or detail panel using `GET /api/v1/ai/compose/activity/digest?workspace_id=...&group_by=user` and `...&channel_id=...&group_by=intent`.
- Surface `user_id` on compose activity rows by hydrating existing user-store display names; treat missing users as `Unknown`.
- On entity detail pages, hydrate `GET /knowledge/entities/:id/brief/automation` so the stale-ring card can show `pending/running/failed` automation state in addition to the existing brief cache.
- Wire schedule chips to `POST /api/v1/ai/schedule/book`, then hydrate list/detail from `/ai/schedule/bookings` and react to `schedule.event.booked` / `schedule.event.cancelled`.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase63H -count=1`
- `go test ./internal/handlers -count=1`
- `go test ./internal/automation -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`
- `git diff --check`

## [0.6.24] - 2026-04-23

This release implements Codex Phase 63G, making AI compose co-drafting activity durable after Windsurf's `v0.6.23` always-on summary UI pass.

### Added

- **Compose activity persistence**:
  - Successful `POST /api/v1/ai/compose` and finalized `POST /api/v1/ai/compose/stream` responses now persist an `AIComposeActivity` row.
  - Rows capture `compose_id`, `workspace_id`, `channel_id` or `dm_id`, `thread_id`, `intent`, `suggestion_count`, `provider`, `model`, and `created_at`.
- **Compose activity API**:
  - `GET /api/v1/ai/compose/activity`
  - Supports `channel_id`, `dm_id`, `workspace_id`, `intent`, and `limit` query filters.
  - Returns newest-first `{ "items": [...] }` for shared co-drafting and agent-collab activity panes.
- **Realtime payload enrichment**:
  - `knowledge.compose.suggestion.generated` now includes both `compose` and persisted `activity` payloads.

### Windsurf Handoff

- Hydrate the current in-memory `composeSuggestionActivity` store from `GET /ai/compose/activity?channel_id=...&limit=50`.
- Keep websocket `knowledge.compose.suggestion.generated` as the live append path, now preferring `payload.activity` when present.
- Consider adding a small co-drafting activity pane in `#agent-collab`, channel info, or workspace Activity.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase63G -count=1`
- `go test ./internal/handlers -run 'TestPhase63' -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.22] - 2026-04-23

This release implements Codex Phase 63F, moving Relay's knowledge layer toward always-on AI assistance after Windsurf's `v0.6.21` entity Ask AI UI pass.

### Added

- **Channel auto-summarize APIs**:
  - `GET /api/v1/channels/:id/knowledge/auto-summarize`
  - `PUT /api/v1/channels/:id/knowledge/auto-summarize`
  - `POST /api/v1/channels/:id/knowledge/auto-summarize`
  - Persists per-channel settings and writes refreshed summaries into the existing `AISummary` channel cache.
- **Realtime channel summary updates**:
  - `POST /channels/:id/knowledge/auto-summarize` emits websocket `channel.summary.updated`.
- **Collaborative compose visibility**:
  - `POST /api/v1/ai/compose` and `POST /api/v1/ai/compose/stream` now emit websocket `knowledge.compose.suggestion.generated` after suggestions finalize.
- **Structured schedule intent slots**:
  - `schedule` compose responses now include additive `proposed_slots[]` with ISO start/end times, duration, timezone, attendee ids, and reason.

### Windsurf Handoff

- Add channel header/sidebar controls for `GET|PUT|POST /channels/:id/knowledge/auto-summarize`.
- Listen for `channel.summary.updated` to refresh the channel summary card without a manual reload.
- Listen for `knowledge.compose.suggestion.generated` where shared co-drafting visibility is useful.
- Render `compose.proposed_slots[]` as calendar chips for schedule intent suggestions.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase63F -count=1`
- `go test ./internal/handlers -run 'TestPhase63' -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.20] - 2026-04-23

This release implements Codex Phase 63E, adding streamable and persistent Entity Ask AI APIs after Windsurf's `v0.6.19` composer UI pass.

### Added

- **Entity Ask AI stream API**:
  - `POST /api/v1/knowledge/entities/:id/ask/stream`
  - Emits SSE events: `start`, `answer.delta`, `answer.done`, `done`, and `error`.
  - Reuses the existing grounded entity prompt and citation builder.
- **Entity Ask AI history API**:
  - `GET /api/v1/knowledge/entities/:id/ask/history`
  - Returns current-user Q&A rows for the entity, newest first.
- **Entity Ask AI persistence**:
  - Sync `POST /api/v1/knowledge/entities/:id/ask` now persists each answer.
  - Streaming ask persists the final answer after a successful stream.

### Windsurf Handoff

- Hydrate the entity detail Ask AI module with `GET /knowledge/entities/:id/ask/history` on page load.
- Switch long-running entity questions to `POST /knowledge/entities/:id/ask/stream` and progressively render `answer.delta`.
- On `answer.done`, append the returned answer/history item to the Ask AI history list.
- Keep sync `POST /knowledge/entities/:id/ask` as fallback if streaming is unavailable.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase63E -count=1`
- `go test ./internal/handlers -run 'TestPhase63' -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.18] - 2026-04-23

This release implements Codex Phase 63D, extending the AI composer into DM scopes and adding intent plus feedback-summary contracts after Windsurf's `v0.6.17` streaming UI pass.

### Added

- **DM-aware AI compose APIs**:
  - `POST /api/v1/ai/compose`
  - `POST /api/v1/ai/compose/stream`
  - Both endpoints now accept exactly one of `channel_id` or `dm_id`.
  - Channel compose keeps optional `thread_id`; DM compose uses recent private-message context.
- **Composer intent variants**:
  - Supported `intent` values are now `reply`, `summarize`, `followup`, and `schedule`.
  - The response shape remains compatible with existing suggestion cards: `suggestions[]`, `citations[]`, `context_entities[]`, `provider`, and `model`.
- **Compose feedback summary API**:
  - `GET /api/v1/ai/compose/:id/feedback/summary`
  - Returns aggregate `up`, `down`, and `edited` counts plus recent feedback rows.
- **DM-scoped compose feedback**:
  - `POST /api/v1/ai/compose/:id/feedback` now accepts `dm_id` as an alternative to `channel_id`.

### Windsurf Handoff

- Enable the AI Suggest button in DM composers by calling `POST /ai/compose/stream` with `{ dm_id, draft, intent, limit }`.
- Add a small intent selector for channel/thread/DM composers using `reply`, `summarize`, `followup`, and `schedule`.
- Send feedback with the same scope used for generation: `{ channel_id, thread_id }` or `{ dm_id }`.
- Optionally build a learning-signal panel using `GET /ai/compose/:id/feedback/summary`.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase63D -count=1`
- `go test ./internal/handlers -run 'TestPhase63(B|C|D)' -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.16] - 2026-04-23

This release implements Codex Phase 63C, extending the grounded composer from one-shot suggestions into a stream-aware reply assistant with explicit suggestion feedback capture after Windsurf's `v0.6.15` UI pass.

### Added

- **Streaming AI compose API**:
  - `POST /api/v1/ai/compose/stream`
  - Supports channel and thread reply-suggestion streaming for the same grounded compose scope as `POST /api/v1/ai/compose`.
  - Emits SSE events:
    - `start`
    - `suggestion.delta`
    - `suggestion.done`
    - `done`
    - `error`
- **AI compose feedback API**:
  - `POST /api/v1/ai/compose/:id/feedback`
  - Persists one feedback row per `(compose_id, user_id)` with upsert semantics.
  - Accepts `channel_id`, optional `thread_id`, optional `intent`, `feedback` (`up|down|edited`), optional `suggestion_text`, optional `provider`, and optional `model`.

### Windsurf Handoff

- Upgrade the composer suggestion popover to consume `POST /ai/compose/stream` for progressive rendering.
- Add thumbs-up / thumbs-down / edited actions per suggestion using `POST /ai/compose/:id/feedback`.
- Keep `POST /ai/compose` as fallback for environments where SSE is unavailable.

### Verification Used For This Release

- `go test ./internal/handlers -run 'TestPhase63(B|C)' -count=1`
- `go test ./internal/handlers -run 'TestPhase63C(ComposeStreamContract|ComposeFeedbackContract)' -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.14] - 2026-04-23

This release implements Codex Phase 63B, pushing AI-native assistance into the Slack-style message flow after Windsurf's `v0.6.13` entity ask/share UI pass.

### Added

- **Grounded AI compose API**:
  - `POST /api/v1/ai/compose`
  - Supports channel and thread reply-suggestion generation.
  - Accepts `channel_id`, optional `thread_id`, optional `draft`, optional `intent` (`reply` only in this phase), and optional `limit`.
  - Returns `suggestions[]`, `citations[]`, `context_entities[]`, `provider`, and `model`.

### Windsurf Handoff

- Add composer-side AI suggestion UI for channel and thread scopes using `POST /ai/compose`.
- Show returned `citations[]` and `context_entities[]` beside or beneath each suggestion.
- Insert a suggestion into the draft only; do not auto-send.

### Verification Used For This Release

- `go test ./internal/handlers -run 'TestPhase63(BAIComposeContract|EntityAskWeeklyShareAndBriefInvalidation)' -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.12] - 2026-04-22

This release implements Codex Phase 63A, extending the AI-native knowledge loop after Windsurf's `v0.6.11` cache-hydration UI pass.

### Added

- **Entity-scoped grounded Q&A API**:
  - `POST /api/v1/knowledge/entities/:id/ask`
  - Accepts `question`, optional `provider`, and optional `model`.
  - Grounds the answer against entity refs, timeline events, and linked entities, then returns `answer`, `citations[]`, and entity metadata.
- **Weekly brief snapshot share API**:
  - `POST /api/v1/knowledge/weekly-brief/:id/share`
  - Returns a share payload for a persisted weekly brief snapshot using the existing `AISummary` cache as the source of truth.
  - Weekly brief responses now include a stable snapshot `id`.
- **Entity brief invalidation realtime event**:
  - websocket `knowledge.entity.brief.changed`
  - Emitted when a cached entity brief becomes stale after new refs, events, or entity edits.

### Windsurf Handoff

- Add an entity-detail Ask AI surface that calls `POST /knowledge/entities/:id/ask` and renders returned `citations[]`.
- Add share CTA for persisted weekly digests using `brief.id` from `GET|POST /knowledge/weekly-brief` and `POST /knowledge/weekly-brief/:id/share`.
- Listen for `knowledge.entity.brief.changed` to pulse a Refresh / Regenerate CTA when a cached brief is stale.

### Verification Used For This Release

- `go test ./internal/handlers -run 'TestPhase6(1|2|3)' -count=1`
- `go test ./internal/knowledge -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.10] - 2026-04-22

This release implements Codex Phase 62, closing the cache/read and atomic-read gaps left after Windsurf's `v0.6.9` UI integration pass.

### Added

- **Cached entity brief API**:
  - `GET /api/v1/knowledge/entities/:id/brief`
  - Returns cached `AISummary` content for an entity without invoking the LLM gateway.
  - Returns `{ "brief": null }` when no cached brief exists.
- **Entity brief realtime event**:
  - websocket `knowledge.entity.brief.generated`
  - Emitted after `POST /api/v1/knowledge/entities/:id/brief` successfully persists a new brief.
- **Cached weekly brief API**:
  - `GET /api/v1/knowledge/weekly-brief?workspace_id=...`
  - Returns cached per-user weekly knowledge brief without invoking the LLM gateway.
  - Returns `{ "brief": null }` when no cached weekly brief exists.
- **Atomic notification bulk-read API**:
  - `POST /api/v1/notifications/bulk-read`
  - Accepts `item_ids[]`, de-duplicates IDs, writes all read markers in one transaction, and emits websocket `notifications.bulk_read`.

### Windsurf Handoff

- Hydrate entity brief cards on page load via `GET /knowledge/entities/:id/brief`; use POST only for Generate/Regenerate.
- Listen for `knowledge.entity.brief.generated` to update entity detail across tabs.
- Hydrate the Following Hub weekly digest on load via `GET /knowledge/weekly-brief?workspace_id=...`.
- Replace per-item "mark all read" loops with `POST /notifications/bulk-read`.

### Verification Used For This Release

- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.8] - 2026-04-22

This release implements Codex Phase 61: AI-native knowledge brief APIs, historical activity backfill, followed-stats realtime updates, and bulk presence hydration after Windsurf's `v0.6.7` UI handoff.

### Added

- **AI knowledge entity brief API**:
  - `POST /api/v1/knowledge/entities/:id/brief`
  - Uses the configured LLM gateway and persists cached output in `AISummary` scope `knowledge_entity`.
  - Request supports `provider`, `model`, and `force`.
- **AI weekly followed-knowledge brief API**:
  - `POST /api/v1/knowledge/weekly-brief`
  - Combines followed stats, followed entities, and workspace trending entities into an LLM prompt.
  - Request supports `workspace_id`, `provider`, `model`, and `force`.
- **Knowledge activity backfill APIs**:
  - `GET /api/v1/knowledge/entities/:id/activity/backfill-status`
  - `POST /api/v1/knowledge/entities/:id/activity/backfill`
  - Scans historical channel messages and files for entity-title matches, creates missing `KnowledgeEntityRef` rows, and emits realtime ref/trending updates.
- **Bulk presence API**:
  - `GET /api/v1/presence/bulk`
  - Supports optional `channel_id` and returns both hydrated users and aggregate counts.
- **Realtime followed stats event**:
  - websocket `knowledge.followed.stats.changed`
  - Emitted after follow, unfollow, per-follow notification change, and bulk follow notification updates.

### Windsurf Handoff

- Add entity-detail **Generate brief** / **Regenerate** actions using `POST /knowledge/entities/:id/brief`.
- Add a Following Hub / Home weekly digest action using `POST /knowledge/weekly-brief`.
- Use `GET /knowledge/entities/:id/activity/backfill-status` to show whether sparklines are historically complete, and call `POST /knowledge/entities/:id/activity/backfill` from an admin/dev action.
- Listen for `knowledge.followed.stats.changed` so the Following Hub stats strip updates without explicit refetch triggers.
- Prefer `GET /presence/bulk?channel_id=...` on reconnect or channel switch for large member lists.

### Verification Used For This Release

- `go test ./internal/handlers -run TestPhase61KnowledgeBriefBackfillStatsRealtimeAndPresenceBulk -count=1`
- `go test ./internal/knowledge -count=1`
- `go test ./...`
- `GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.6] - 2026-04-22

This release pairs Windsurf's `v0.6.5` UI pass with Codex Phase 60 backend follow-up. Relay's knowledge layer now has aggregate follow stats, shareable entity deeplinks, and realtime trending broadcasts so the knowledge surfaces can move from pull-only to push-aware.

### Added

- **Followed knowledge stats API**:
  - `GET /api/v1/users/me/knowledge/followed/stats`
  - Returns:
    - `total_count`
    - `spiking_count`
    - `muted_count`
    - `by_kind[]`
- **Knowledge entity share API**:
  - `POST /api/v1/knowledge/entities/:id/share`
  - Returns:
    - `entity_id`
    - `workspace_id`
    - `title`
    - `url`
    - `short_url`
    - `relative_path`
- **Realtime trending event**:
  - websocket `knowledge.trending.changed`
  - Payload includes:
    - `workspace_id`
    - `days`
    - `items[]`

### Changed

- `knowledge.entity.ref.created` flows now also emit `knowledge.trending.changed`, including:
  - direct `POST /api/v1/knowledge/entities/:id/refs`
  - auto-linked message entity refs
  - auto-linked file entity refs
- Included Windsurf's `v0.6.5` UI pass in this release line:
  - Following Hub bulk mute/restore
  - Trending cards on Knowledge and Home
  - entity activity sparkline
  - workspace knowledge alert tuning tab

### Windsurf Handoff

- Use `GET /api/v1/users/me/knowledge/followed/stats` for a Following Hub header strip or weekly AI-native follow digest summary.
- Use `POST /api/v1/knowledge/entities/:id/share` for share actions on Trending cards, entity headers, and sparkline surfaces.
- Listen for websocket `knowledge.trending.changed` to live-refresh Trending cards without polling.
- Remaining recommended backend target after this release:
  - `presence.bulk`
  - activity history backfill for pre-Phase-57 entities

### Verification Used For This Release

- `go test ./internal/knowledge -run TestFollowedStatsAndSharedEntityLink -count=1`
- `go test ./internal/handlers -run TestPhase60KnowledgeShareStatsAndTrendingRealtime -count=1`
- `go test ./...`
- `bash -lc 'GOCACHE=$(pwd)/.cache/go-build go build ./...'`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.4] - 2026-04-22

This release pairs Windsurf's `v0.6.3` Following Hub UI with Codex Phase 59 backend contracts. It turns the knowledge-follow loop into an operational surface with bulk controls, workspace-level spike tuning, entity activity history, and a trending feed for future Home/Knowledge panels.

### Added

- **Bulk knowledge follow settings API**:
  - `PATCH /api/v1/users/me/knowledge/followed/bulk`
  - Accepts:
    - `entity_ids[]`
    - `notification_level`
- **Workspace knowledge settings APIs**:
  - `GET /api/v1/workspace/settings`
  - `PATCH /api/v1/workspace/settings`
  - Fields:
    - `workspace_id`
    - `spike_threshold`
    - `spike_cooldown_minutes`
- **Knowledge entity activity API**:
  - `GET /api/v1/knowledge/entities/:id/activity`
  - Returns daily `buckets[]` for sparkline/mini-chart UI.
- **Knowledge trending API**:
  - `GET /api/v1/knowledge/trending`
  - Returns ranked entities with:
    - `recent_ref_count`
    - `previous_ref_count`
    - `velocity_delta`
    - `related_channel_ids`
    - `last_ref_at`

### Changed

- `knowledge.entity.activity.spiked` now reads workspace-level `spike_threshold` and `spike_cooldown_minutes` instead of relying only on hardcoded defaults.
- `workspaces` now persist knowledge alert settings:
  - `knowledge_spike_threshold`
  - `knowledge_spike_cooldown_mins`
- Seeded workspaces now default to:
  - `spike_threshold = 3`
  - `spike_cooldown_minutes = 360`
- Included Windsurf's `v0.6.3` UI pass in this release line:
  - `/workspace/knowledge/following`
  - locale-aware inbox date formatting
  - Following Hub inline notification pickers and Mute All flow

### Windsurf Handoff

- Replace the current N-request Mute All implementation with `PATCH /api/v1/users/me/knowledge/followed/bulk`.
- Add workspace-level spike tuning controls using `GET|PATCH /api/v1/workspace/settings`.
- Render entity sparklines from `GET /api/v1/knowledge/entities/:id/activity`.
- Add Trending modules on Knowledge and Home using `GET /api/v1/knowledge/trending`.

### Verification Used For This Release

- `go test ./internal/knowledge -run 'Test(BulkFollowUpdatesWorkspaceSettingsActivityAndTrending|UpdateFollowNotificationLevelAndDetectSpikeAlerts)' -count=1`
- `go test ./internal/handlers -run 'Test(Phase59KnowledgeOpsEndpoints|KnowledgeEntityFollowEndpoints)' -count=1`
- `go test ./...`
- `bash -lc 'GOCACHE=$(pwd)/.cache/go-build go build ./...'`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.6.1] - 2026-04-22

This release combines Windsurf's `v0.6.0` UI completion pass with Codex Phase 57 backend follow-up. It closes the next AI-native knowledge loop: followed entities can now carry notification preferences and emit realtime spike alerts.

### Added

- **Knowledge follow settings API**:
  - `PATCH /api/v1/users/me/knowledge/followed/:id`
  - Accepts:
    - `notification_level`
  - Supported values:
    - `all`
    - `digest_only`
    - `silent`
- **Realtime entity spike event**:
  - websocket `knowledge.entity.activity.spiked`
  - Payload includes:
    - `entity`
    - `user_ids`
    - `channel_id`
    - `recent_ref_count`
    - `previous_ref_count`
    - `delta`
    - `related_channel_ids`
    - `occurred_at`

### Changed

- `knowledge_entity_follows` now persist:
  - `notification_level`
  - `last_alerted_at`
- New follows default to `notification_level = "all"`.
- Spike alerts are emitted only for `all`-level followers and are rate-limited per follow row using `last_alerted_at`.
- Included Windsurf's `v0.6.0` UI pass in this release train:
  - Settings page hydrates from `GET /api/v1/me/settings`
  - Knowledge Inbox detail consumes `GET /api/v1/knowledge/inbox/:id`
  - Digest schedule dialog uses `POST /api/v1/channels/:id/knowledge/digest/preview-schedule`
  - DMs landing page, workspace header actions, set-status dialog, and right-panel scrolling fixes shipped

### Windsurf Handoff

- Listen for websocket `knowledge.entity.activity.spiked` and only surface it when `payload.user_ids` includes the current user.
- Use `PATCH /api/v1/users/me/knowledge/followed/:id` to let users choose `all | digest_only | silent`.
- Recommended next UI slice:
  - follow settings picker on entity pages / hover cards
  - toast + pulse treatment for spiking followed entities
  - locale/timezone picker on Profile tab to consume the already-shipped settings fields

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'TestKnowledgeEntityFollowEndpoints' -count=1`
- `cd apps/api && go test ./internal/knowledge -run 'Test(UpdateFollowNotificationLevelAndDetectSpikeAlerts)' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.99] - 2026-04-22

This release implements Phase 56: knowledge digest drill-down plus cross-device user settings sync. It closes the main backend gaps left after Windsurf's v0.5.98 UI work.

### Added

- **Knowledge inbox detail API**:
  - `GET /api/v1/knowledge/inbox/:id`
  - Returns the selected digest item plus `entity_contexts[]` with top-entity sample messages captured from the channel timeline.
- **Digest schedule dry-run API**:
  - `POST /api/v1/channels/:id/knowledge/digest/preview-schedule`
  - Accepts the same scheduling fields as `PUT /channels/:id/knowledge/digest/schedule` plus optional `count`.
  - Returns normalized `schedule`, `upcoming_runs[]`, and a current `digest` preview for confidence before saving.
- **Settings hydration API**:
  - `GET /api/v1/me/settings`
  - Returns `provider`, `model`, `mode`, `theme`, `message_density`, `locale`, and `timezone`.

### Changed

- `PATCH /api/v1/me/settings` now supports partial updates for:
  - `provider`
  - `model`
  - `mode`
  - `theme`
  - `message_density`
  - `locale`
  - `timezone`
- Added persisted user preference fields for:
  - `theme_preference`
  - `message_density`
  - `locale`
- Settings validation now guards invalid `theme`, `message_density`, and `timezone` values instead of silently persisting bad data.

### Windsurf Handoff

- Use `GET /api/v1/me/settings` to hydrate the Settings page on load instead of relying on `localStorage` defaults.
- Persist Appearance tab changes through `PATCH /api/v1/me/settings` with `theme`, `message_density`, `locale`, and `timezone`.
- Use `GET /api/v1/knowledge/inbox/:id` for a real digest drill-down surface with entity sample messages.
- Use `POST /api/v1/channels/:id/knowledge/digest/preview-schedule` inside the schedule dialog to show the next runs and current digest preview before save.
- Remaining recommended backend target after this release:
  - websocket `knowledge.entity.activity.spiked`

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'Test(GetMeSettingsReturnsHydratedPreferences|PatchMeSettingsPersistsPreferences|PreviewChannelKnowledgeDigestScheduleReturnsUpcomingRuns|GetKnowledgeInboxItemReturnsDigestContext)' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.97] - 2026-04-22

This release implements the next knowledge-collaboration backend slice for Windsurf: per-user entity follow plus composer-side reverse entity lookup.

### Added

- **Knowledge entity follow APIs**:
  - `GET /api/v1/users/me/knowledge/followed`
  - `POST /api/v1/knowledge/entities/:id/follow`
  - `DELETE /api/v1/knowledge/entities/:id/follow`
- **Composer reverse lookup API**:
  - `POST /api/v1/knowledge/entities/match-text`
  - Request body:
    - `workspace_id`
    - `text`
    - `limit`
  - Match items return:
    - `entity_id`
    - `entity_title`
    - `entity_kind`
    - `source_kind`
    - `matched_text`
    - `start`
    - `end`

### Changed

- Added persistent `knowledge_entity_follows` storage via GORM migration.
- Entity follow creation is idempotent per `(entity_id, user_id)`.
- Entity text matching is deterministic, workspace-scoped, ignores archived entities, and prefers the longest non-overlapping entity titles so `Launch Program` does not double-match `Launch`.
- Cleaned a frontend collaboration warning by removing an unused `Plus` import from `apps/web/components/layout/primary-nav.tsx`.

### Windsurf Handoff

- Add a Knowledge follow tab or filter using `GET /api/v1/users/me/knowledge/followed`.
- Use `POST /api/v1/knowledge/entities/:id/follow` and `DELETE /api/v1/knowledge/entities/:id/follow` for follow toggles on wiki pages, hover cards, or digest surfaces.
- Use `POST /api/v1/knowledge/entities/match-text` from the composer draft text to show passive “entity detected” hints and one-click conversion into explicit `@entity` mentions.
- Remaining recommended backend targets after this release:
  - `GET /api/v1/knowledge/inbox/:id` for richer digest detail context
  - `POST /api/v1/channels/:id/knowledge/digest/preview-schedule`
  - websocket `knowledge.entity.activity.spiked`

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'Test(KnowledgeEntityFollowEndpoints|MatchKnowledgeEntitiesInTextEndpoint)' -count=1`
- `cd apps/api && go test ./internal/knowledge -run 'Test(FollowEntityAndListFollowedEntities|MatchEntitiesInTextReturnsLongestNonOverlappingSpans)' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.95] - 2026-04-22

This release fixes the v0.5.94 lint regression in `apps/web/components/message/message-composer.tsx`.

### Fixed

- Removed the stale `// eslint-disable-next-line react-hooks/exhaustive-deps` directive from `message-composer.tsx`.
- Root cause: the project ESLint flat config does not register the `react-hooks` plugin, so ESLint 9 treats that disable directive as an unknown rule error.
- The draft-restore effect behavior from v0.5.94 remains unchanged: it still depends on `[scope, editor]` and uses `draftsRef` to avoid re-populating sent draft content.

### Verification Used For This Release

- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.5.94] - 2026-04-22

This release contains Windsurf's UI bug-fix pass for workspace usability and Agent-Collab statistics.

### Fixed

- Home dashboard scroll behavior inside the workspace resizable panel.
- Recent Conversations raw HTML previews now render as plain text.
- User hover card Message action opens a docked DM overlay instead of forcing route navigation.
- Composer draft clearing no longer re-populates the last sent content after send.
- AI Assistant avatar now uses `/ai-wand-avatar.svg`.
- Agent-Collab statistics include Windsurf in assignee breakdown and add contribution heatmap/trend-line visualizations.

## [0.5.93] - 2026-04-22

This release implements Phase 53: Channel Persistence Hardening. It fixes a refresh-loss bug where newly created channels could appear locally, then disappear after reload because the frontend posted a legacy mock workspace ID (`ws_1`) instead of the persisted workspace ID (`ws-1`).

### Fixed

- **Channel creation persistence**:
  - `apps/web/stores/channel-store.ts` now creates channels with `workspace-store.currentWorkspace.id` instead of deriving the workspace from an unmapped channel payload or falling back to `ws_1`.
  - Channel API payloads are now mapped from backend snake_case into frontend camelCase (`workspace_id` -> `workspaceId`, `member_count` -> `memberCount`, `is_starred` -> `isStarred`, etc.).
  - `POST /api/v1/channels` now rejects unknown `workspace_id` values instead of silently creating orphan channels.

### Changed

- **Legacy DB repair**:
  - API startup now repairs old channel rows with `workspace_id = "ws_1"` by moving them to the canonical Relay workspace `ws-1` when that workspace exists.
  - Empty duplicate legacy channels with the same name are cleaned up during repair to avoid duplicate sidebar entries after recovery.
  - This should recover channels such as `#game` that were created through the previous frontend fallback and then disappeared from `GET /api/v1/channels?workspace_id=ws-1`.

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'TestCreateChannel(RejectsUnknownWorkspace|CreatesChannelAndOwnerMembership)' -count=1`
- `cd apps/api && go test ./internal/db -run TestRepairLegacyWorkspaceIDsMovesMockWorkspaceChannels -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.92] - 2026-04-22

This release implements Phase 52: Digest Automation and Knowledge Inbox APIs. Relay can now auto-publish scheduled channel digests, surface them in a cross-channel knowledge inbox, and expose knowledge digest aggregates through the Home API.

### Added

- **Digest Schedule APIs**:
  - `GET /api/v1/channels/:id/knowledge/digest/schedule`
  - `PUT /api/v1/channels/:id/knowledge/digest/schedule`
  - `DELETE /api/v1/channels/:id/knowledge/digest/schedule`
  - Schedule fields include:
    - `window`
    - `timezone`
    - `day_of_week`
    - `day_of_month`
    - `hour`
    - `minute`
    - `limit`
    - `pin`
    - `is_enabled`
    - `last_published_at`
    - `next_run_at`
- **Knowledge Inbox API**:
  - `GET /api/v1/knowledge/inbox`
  - Query params:
    - `scope=all|starred`
    - `limit`
  - Inbox items return:
    - `id`
    - `channel`
    - `message`
    - `digest`
    - `is_read`
    - `occurred_at`
- **Background Digest Scheduler**:
  - The API server now runs an in-process digest scheduler that checks persisted schedules every minute and publishes due digest messages automatically.
- **Realtime Digest Event**:
  - New websocket event: `knowledge.digest.published`

### Changed

- **Shared Digest Publish Path**:
  - Manual and scheduled digest publishing now use the same backend publish flow, keeping `message.metadata.knowledge_digest` consistent across both paths.
- **Home Aggregation**:
  - `GET /api/v1/home` now also returns:
    - `knowledge_inbox_count`
    - `recent_knowledge_digests`
- **Notification Compatibility**:
  - Knowledge inbox entries use stable notification item IDs (`knowledge-digest-<message_id>`) so the existing `POST /api/v1/notifications/read` flow can mark digest items read without a new notification subsystem.

### Windsurf Handoff

- Use `GET /api/v1/channels/:id/knowledge/digest/schedule` plus `PUT`/`DELETE` to add channel-level scheduling controls in the digest banner or channel settings.
- Use `GET /api/v1/knowledge/inbox?scope=all` for a top-level cross-channel knowledge inbox route and `scope=starred` for a tighter “followed channels” mode.
- Consume `home.knowledge_inbox_count` and `home.recent_knowledge_digests` for Home dashboard cards or badges.
- Listen for `knowledge.digest.published` if the UI wants to live-refresh cross-channel digest surfaces without waiting for a manual fetch.

### Verification Used For This Release

- `cd apps/api && go test ./internal/knowledge -run 'Test(UpsertDigestScheduleAndComputeNextRunAt|ProcessDigestSchedulesPublishesDueDigestOnce|ListKnowledgeInboxReturnsDigestMessagesAndReadState)' -count=1`
- `cd apps/api && go test ./internal/handlers -run 'Test(DigestScheduleEndpointsSupportUpsertGetAndDelete|GetKnowledgeInboxReturnsDigestMessages|GetHomeIncludesKnowledgeDigestSummary)' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.91] - 2026-04-22

This release implements Phase 51: Knowledge Discovery APIs. Relay now exposes entity-centric message discovery, live hover-card activity summaries, and channel knowledge digests that can be previewed or published as pinned messages.

### Added

- **Entity Message Discovery API**:
  - `GET /api/v1/search/messages/by-entity`
  - Required query param:
    - `entity_id`
  - Optional query params:
    - `channel_id`
    - `limit`
  - Response payload returns entity-scoped message results with:
    - `entity_id`
    - `entity_title`
    - `match_sources`
    - refreshed `metadata`
    - search `snippet`
- **Knowledge Entity Hover Summary API**:
  - `GET /api/v1/knowledge/entities/:id/hover`
  - Optional query params:
    - `channel_id`
    - `days`
  - Hover payload returns:
    - `ref_count`
    - `channel_ref_count`
    - `message_ref_count`
    - `file_ref_count`
    - `recent_ref_count`
    - `last_activity_at`
    - `related_channels[]`
- **Channel Knowledge Digest APIs**:
  - `GET /api/v1/channels/:id/knowledge/digest`
  - `POST /api/v1/channels/:id/knowledge/digest/publish`
  - Supported digest windows:
    - `daily`
    - `weekly`
    - `monthly`
  - Digest payload returns:
    - `headline`
    - `summary`
    - `top_movements[]`
    - `recent_ref_count`
    - `total_refs`
- **Digest Message Metadata**:
  - Published digest messages now preserve structured `message.metadata.knowledge_digest` for downstream UI rendering.

### Changed

- **Message Metadata Preservation**:
  - Message metadata refresh now preserves a structured `knowledge_digest` block alongside reactions, attachments, and entity mentions.
- **Knowledge Discovery Surface**:
  - Relay's knowledge layer is no longer limited to side panels and wiki pages. The backend now provides one contract for:
    - searching messages by entity
    - enriching hover cards with live activity
    - publishing Slack-style digest messages without inventing separate storage

### Windsurf Handoff

- Use `GET /api/v1/search/messages/by-entity?entity_id=...&channel_id=...` for entity click-through drilldowns from hover cards, mention chips, or knowledge panels.
- Use `GET /api/v1/knowledge/entities/:id/hover?channel_id=...&days=7` to enrich the existing `EntityMentionChip` hover content with live ref counts, last activity, and related channels.
- Use `GET /api/v1/channels/:id/knowledge/digest?window=weekly&limit=5` for a non-destructive banner preview, and `POST /api/v1/channels/:id/knowledge/digest/publish` with `{ "window": "weekly", "limit": 5, "pin": true }` when the UI wants to publish the digest into the channel as a pinned message.
- Published digest messages intentionally do not auto-link themselves back into knowledge refs. This avoids recursive “digest creates more digest input” noise.

### Verification Used For This Release

- `cd apps/api && go test ./internal/knowledge -run 'Test(GetEntityHoverSummaryAggregatesLiveRefStats|BuildChannelKnowledgeDigestRanksRecentMovements)' -count=1`
- `cd apps/api && go test ./internal/handlers -run 'Test(SearchMessagesByEntityReturnsKnowledgeBackedAndExplicitMatches|GetKnowledgeEntityHoverReturnsLiveActivitySummary|GetChannelKnowledgeDigestReturnsTopMovements|PublishChannelKnowledgeDigestCreatesPinnedMessage)' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.90] - 2026-04-22

This release implements Phase 50: Message Entity Mentions and Knowledge Velocity APIs. Relay messages now return structured knowledge-entity mention metadata, and channel knowledge summaries now expose velocity/anomaly fields for header badges and trend alerts.

### Added

- **Structured Message Entity Mentions**:
  - Channel message metadata now includes `entity_mentions` when content explicitly mentions a knowledge entity as `@Entity Title`.
  - Mention payload fields include:
    - `entity_id`
    - `entity_title`
    - `entity_kind`
    - `source_kind`
    - `mention_text`
- **Knowledge Summary Velocity Fields**:
  - `GET /api/v1/channels/:id/knowledge/summary` now also returns:
    - `velocity.recent_window_days`
    - `velocity.previous_ref_count`
    - `velocity.recent_ref_count`
    - `velocity.delta`
    - `velocity.is_spiking`

### Changed

- **Message Metadata Hydration**:
  - `POST /api/v1/messages`, `GET /api/v1/messages`, and `GET /api/v1/messages/:id/thread` now hydrate explicit knowledge entity mentions into `message.metadata`.
  - Explicit `@Entity Title` mentions are resolved longest-title-first so `@Launch Program` does not double-match a shorter entity like `@Launch`.
- **Channel Knowledge Alerts**:
  - Knowledge summaries now expose a backend-computed velocity signal rather than making the UI guess from raw counts alone.

### Windsurf Handoff

- Use `message.metadata.entity_mentions` to render `@Entity Title` tokens in the feed/thread as linked knowledge mentions with hover cards.
- Use `summary.velocity.is_spiking` and `summary.velocity.delta` from `GET /api/v1/channels/:id/knowledge/summary` for the channel-header anomaly badge.
- The bulk entity-link confirmation flow after file extraction is not part of `0.5.90`; if you want to build it next, I will add a dedicated detection/review contract instead of overloading the current auto-link path.

### Verification Used For This Release

- `cd apps/api && go test ./internal/knowledge -run 'TestGetChannelKnowledgeSummaryAggregatesTopEntitiesAndTrend|TestFindMentionedEntitiesResolvesExplicitEntityMentions' -count=1`
- `cd apps/api && go test ./internal/handlers -run 'TestCreateMessageHydratesExplicitKnowledgeEntityMentions|TestGetChannelKnowledgeSummaryReturnsTopEntitiesAndTrend' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.89] - 2026-04-22

This release implements Phase 49: Channel Knowledge Summary and Entity Mention APIs. Relay now exposes a channel-level knowledge summary endpoint for active-channel sidebars/cards and a scoped entity suggestion endpoint for `@entity:` composer autocomplete.

### Added

- **Channel Knowledge Summary API**:
  - `GET /api/v1/channels/:id/knowledge/summary`
  - Query params:
    - `limit`
    - `days`
  - Response shape: `{ "summary": { "channel_id": string, "window_days": number, "total_refs": number, "recent_ref_count": number, "top_entities": [...] } }`
  - Top entity payload includes:
    - `entity_id`
    - `entity_title`
    - `entity_kind`
    - `ref_count`
    - `message_ref_count`
    - `file_ref_count`
    - `last_ref_at`
    - `trend`
- **Knowledge Entity Suggest API**:
  - `GET /api/v1/knowledge/entities/suggest`
  - Query params:
    - `q`
    - `channel_id`
    - `workspace_id`
    - `limit`
  - Response shape: `{ "query": string, "suggestions": [...] }`
  - Suggestion payload includes:
    - `id`
    - `title`
    - `kind`
    - `summary`
    - `source_kind`
    - `ref_count`
    - `channel_ref_count`

### Changed

- **Channel Knowledge Aggregation**:
  - Channel knowledge now has both a recent-ref feed (`/channels/:id/knowledge`) and an aggregated summary view (`/channels/:id/knowledge/summary`) built from canonical `KnowledgeEntityRef` records across channel messages and files.
- **Scoped Entity Ranking**:
  - Entity suggestions can now scope to the active channel's workspace and rank by channel-local reference frequency, which is the intended backend contract for `@entity:` autocomplete.

### Windsurf Handoff

- Use `GET /api/v1/channels/:id/knowledge/summary?days=7&limit=5` to render a compact "most referenced entities + sparkline" card in the channel knowledge panel or header area.
- Use `GET /api/v1/knowledge/entities/suggest?q=...&channel_id=...&limit=8` for `@entity:` autocomplete in `MessageComposer`.
- The non-intrusive `knowledge.entity.ref.created` banner/toast can reuse the existing websocket event. When it fires for the active channel, refresh `/channels/:id/knowledge` and `/channels/:id/knowledge/summary` and surface the newest linked entity.

### Verification Used For This Release

- `cd apps/api && go test ./internal/knowledge -run 'TestGetChannelKnowledgeSummaryAggregatesTopEntitiesAndTrend|TestSuggestEntitiesScopesToChannelWorkspaceAndRanksByChannelRefs' -count=1`
- `cd apps/api && go test ./internal/handlers -run 'TestGetChannelKnowledgeSummaryReturnsTopEntitiesAndTrend|TestSuggestKnowledgeEntitiesReturnsScopedAutocompleteResults' -count=1`
- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.88] - 2026-04-22

This release implements Phase 48: Channel Knowledge Context. Relay now exposes channel-scoped knowledge refs for active-channel banners/sidebars and hydrates citation search results from the canonical `KnowledgeEntityRef` table created by message/file auto-linking.

### Added

- **Channel Knowledge Context API**:
  - `GET /api/v1/channels/:id/knowledge`
  - Response shape: `{ "context": { "channel_id": string, "refs": [...] } }`
  - Ref payload fields include `entity_id`, `entity_title`, `entity_kind`, `ref_kind`, `ref_id`, `role`, `source_title`, `source_snippet`, and `created_at`.

### Changed

- **Citation Entity Hydration**:
  - `GET /api/v1/citations/lookup` now hydrates `entity_id` and `entity_title` from `KnowledgeEntityRef` when direct message/file refs exist.
  - This means auto-linked messages and files become entity-aware in citation/search results without needing separate evidence-link rows.
- **Knowledge Context Service**:
  - Added channel aggregation for message and file refs, sorted by newest association first with `limit` support.

### Windsurf Handoff

- Use `GET /api/v1/channels/:id/knowledge` for the Phase 48 channel banner or right-side knowledge context panel.
- On `knowledge.entity.ref.created`, if the event belongs to the active channel, refresh this endpoint and show the newly linked entity/ref.
- Citation search cards can now rely on hydrated `entity_id/entity_title` from auto-linked `KnowledgeEntityRef` data.

### Verification Used For This Release

- `cd apps/api && go test ./internal/knowledge -run TestLookupHydratesEntityFromKnowledgeEntityRef -count=1`
- `cd apps/api && go test ./internal/handlers -run TestGetChannelKnowledgeContextReturnsRecentRefs -count=1`
- `cd apps/api && go test ./internal/knowledge`
- `cd apps/api && go test ./internal/handlers`
- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm --filter relay-agent-workspace exec tsc --noEmit`

## [0.5.87] - 2026-04-22

This release implements Phase 47: Knowledge Live Events and Auto-Linking. Relay's wiki layer now updates in real time, accepts external/live business events, exposes richer graph metadata, and automatically links newly created messages/files to mentioned knowledge entities.

### Added

- **Knowledge Realtime Events**:
  - `knowledge.entity.created`
  - `knowledge.entity.updated`
  - `knowledge.entity.ref.created`
  - `knowledge.event.created`
  - `knowledge.link.created`
- **Knowledge Event Ingest API**:
  - `POST /api/v1/knowledge/events/ingest`
  - Request fields: `entity_id`, `event_type`, `title`, `body`, `actor_user_id`, `source_kind`, `source_ref`.
- **Deterministic Entity Auto-Linking**:
  - New channel messages that mention a knowledge entity title create `KnowledgeEntityRef` records with `ref_kind=message` and `role=discussion`.
  - Newly uploaded files whose filename, extracted text, or summary mentions an entity title create refs with `ref_kind=file` and `role=evidence`.
- **Richer Graph Payloads**:
  - Graph nodes now include typed reference metadata: `source_kind`, `ref_kind`, `ref_id`, and `role`.
  - Graph edges now include `weight`, `direction`, and `role`.

### Changed

- Knowledge create/update/ref/event/link handlers now broadcast WebSocket events so entity pages can refresh without manual reload.
- File upload now emits `file.extraction.updated` first, then any derived `knowledge.entity.ref.created` events.
- Message creation keeps the existing `message.created` event order before derived knowledge ref events.

### Windsurf Handoff

- Listen for:
  - `knowledge.entity.created`
  - `knowledge.entity.updated`
  - `knowledge.entity.ref.created`
  - `knowledge.event.created`
  - `knowledge.link.created`
- Refresh entity list/detail/timeline/refs/graph views when these events arrive.
- Add a lightweight "Live Event" composer or debug action using `POST /api/v1/knowledge/events/ingest`.
- Update graph UI to use `edge.weight`, `edge.direction`, `edge.role`, and typed ref-node metadata.

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'TestKnowledgeEntity(GraphEndpoint|EndpointsBroadcastRealtimeEvents|EventIngestCreatesTimelineEventAndBroadcasts)|Test(CreateMessageAutoLinksMentionedKnowledgeEntity|UploadFileAutoLinksMentionedKnowledgeEntity)'`
- `cd apps/api && go test ./internal/knowledge`
- `cd apps/api && go test ./internal/handlers`
- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.86] - 2026-04-21

This release implements Phase 46: Knowledge Entities and Wiki Substrate. Relay now has first-class knowledge entities that can connect files, messages, artifacts, workflows, and future business objects into wiki-style pages and relationship previews.

### Added

- **Knowledge Entity APIs**:
  - `GET /api/v1/knowledge/entities`
  - `POST /api/v1/knowledge/entities`
  - `GET /api/v1/knowledge/entities/:id`
  - `PATCH /api/v1/knowledge/entities/:id`
- **Knowledge Reference APIs**:
  - `GET /api/v1/knowledge/entities/:id/refs`
  - `POST /api/v1/knowledge/entities/:id/refs`
- **Knowledge Timeline APIs**:
  - `GET /api/v1/knowledge/entities/:id/timeline`
  - `POST /api/v1/knowledge/entities/:id/events`
- **Knowledge Link And Graph Preview APIs**:
  - `GET /api/v1/knowledge/entities/:id/links`
  - `POST /api/v1/knowledge/links`
  - `GET /api/v1/knowledge/entities/:id/graph`
- **Knowledge Wiki Persistence Models**:
  - `KnowledgeEntity`
  - `KnowledgeEntityRef`
  - `KnowledgeEntityLink`
  - `KnowledgeEvent`

### Changed

- **Citation Entity Hydration**:
  - Citation lookup now hydrates `entity_title` when a citation is connected to a `KnowledgeEntity`.
- **Knowledge Service Expansion**:
  - `apps/api/internal/knowledge/` now handles both citation lookup and the entity/wiki substrate.

### Windsurf Handoff

- Build the first wiki/entity detail page using:
  - entity detail
  - refs
  - timeline
  - graph preview
- Add entity badges and links wherever `entity_id` / `entity_title` appear in `CitationCard`.
- Suggested next UI routes:
  - `/workspace/knowledge`
  - `/workspace/knowledge/[id]`

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.85] - 2026-04-21

This release implements Phase 45: AI Citation Lookup. Relay now has a unified evidence lookup layer that can search across file chunks, channel messages, thread messages, and artifact sections, while reserving entity-aware fields for the upcoming wiki and graph phases.

### Added

- **Unified Citation Lookup API**:
  - `GET /api/v1/citations/lookup?q=...`
  - Returns mixed evidence from:
    - `file_chunk`
    - `message`
    - `thread`
    - `artifact_section`
- **Knowledge Citation Service**:
  - Added `apps/api/internal/knowledge/` for cross-source evidence lookup, normalization, and deterministic ranking.
- **Evidence Persistence Models**:
  - Added `KnowledgeEvidenceLink`
  - Added `KnowledgeEvidenceEntityRef`
- **Entity-Aware Citation Fields**:
  - Citation payloads now reserve:
    - `entity_id`
    - `entity_title`
    - `source_kind`
    - `source_ref`
    - `ref_kind`
    - `locator`
    - `score`

### Changed

- **File Citation Contract**:
  - `GET /api/v1/files/:id/citations` now aligns with the unified citation shape instead of returning a file-specific chunk-only schema.
- **AI Execute Request Compatibility**:
  - `POST /api/v1/ai/execute` now accepts a future-facing `citations` field in the request body without rejecting the payload shape.

### Windsurf Handoff

- Add a shared citation card renderer that consumes `GET /api/v1/citations/lookup`.
- Treat `evidence_kind` as the primary visual switch:
  - `file_chunk`
  - `message`
  - `thread`
  - `artifact_section`
- Use:
  - `title`
  - `snippet`
  - `locator`
  - `source_kind`
  - `ref_kind`
  - `entity_id`
- For now, `entity_title` may be empty. Do not block UI on it; Phase 46 will hydrate richer entity details.
- Next frontend target after integration:
  - citation picker / inspector
  - entity-aware evidence surfaces in files, messages, and artifacts

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.84] - 2026-04-21

This release implements Phase 44: File Extraction, Search, and Citation Substrate. Files now move beyond metadata-only objects and become indexed knowledge sources for search, retrieval, and future AI citation flows.

### Added

- **File Extraction Lifecycle APIs**:
  - `GET /api/v1/files/:id/extraction`
  - `POST /api/v1/files/:id/extraction/rebuild`
  - `GET /api/v1/files/:id/chunks`
  - `GET /api/v1/files/:id/citations`
- **File Content Retrieval API**:
  - `GET /api/v1/files/:id/extracted-content`
  - Returns normalized extracted text and summary without colliding with the existing raw download endpoint at `GET /api/v1/files/:id/content`.
- **File Content Search API**:
  - `GET /api/v1/search/files?q=...`
  - Returns file matches with snippets, locator metadata, and match reasons.
- **File Indexing Package**:
  - Added `apps/api/internal/fileindex/` for extraction routing, chunking, and OCR abstraction.
- **Office And PDF Extraction Support**:
  - Real parsing paths for `txt`, `md`, `pdf`, `docx`, `xlsx`, and `pptx`.
- **Mock OCR Path**:
  - Image files now flow through an OCR provider interface with an initial mock implementation.
- **Realtime File Indexing Event**:
  - Added websocket `file.extraction.updated`.

### Changed

- **File Payloads** now include:
  - `extraction_status`
  - `content_summary`
  - `last_indexed_at`
  - `needs_ocr`
  - `ocr_provider`
  - `ocr_is_mock`
  - `is_searchable`
  - `is_citable`
- **Legacy Office Formats**:
  - `doc`, `xls`, and `ppt` now fail explicitly with extraction status instead of pretending they are searchable.

### Windsurf Handoff

- Add file indexing state badges to Files and inline message file cards:
  - `processing`
  - `ready`
  - `failed`
  - `ocr needed`
- Use:
  - `GET /api/v1/files/:id/extraction`
  - `GET /api/v1/files/:id/extracted-content`
  - `GET /api/v1/files/:id/chunks`
  - `GET /api/v1/files/:id/citations`
  - `GET /api/v1/search/files?q=...`
- Listen for websocket `file.extraction.updated` to refresh file/search state live.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/fileindex`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.83] - 2026-04-21

This release implements Phase 43: Message-Level File Attachment APIs. Shared files can now render as rich cards inside channel messages and thread views without requiring the frontend to reconstruct file metadata from multiple endpoints.

### Added

- **Rich Message File Attachments**: `metadata.attachments` for `kind="file"` now includes preview metadata, uploader, counters, knowledge fields, archive/retention state, and nested `file` / `preview` payloads suitable for inline cards.
- **Message Files Endpoint**: Added `GET /api/v1/messages/:id/files` to return message-scoped hydrated file attachments for lazy loading or detail drawers.

### Changed

- **UUID Normalization**: Newly created channel messages, share-generated messages, DM conversations, DM messages, workspace invites, and agents now use prefixed UUIDs instead of timestamp-style string IDs.
- **README Sync**: Updated the published capability list so GitHub-facing project documentation matches the shipped backend surface.

### Windsurf Handoff

- Render inline file cards directly from `message.metadata.attachments[*].file` and `message.metadata.attachments[*].preview`.
- Use `GET /api/v1/messages/:id/files` as a fallback or lazy-load path for message thread drawers and richer attachment inspectors.
- Expect newly created IDs from message/DM/invite flows to use prefixed UUIDs: `msg-*`, `dm-*`, `dm-msg-*`, `invite-*`, and `agent-*`.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestCreateMessageHydratesArtifactAndFileAttachments|TestDMEndpointsListCreateAndSendMessages|TestWorkspaceInvitesEndpointsCreateAndListInvites'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.5.82] - 2026-04-21

This release implements Phase 42: File Collaboration Integration. Files are now first-class collaborative knowledge objects with commenting, sharing, starring, and knowledge metadata.

### Added

- **Starred Files Filter**: Header toggle to browse only starred files. `fetchStarredFiles()` calls `GET /api/v1/files/starred`.
- **Star Toggle**: Star icon per file row (inline) and in preview dialog header. Optimistic UI update with toast. `toggleFileStar()` calls `POST /api/v1/files/:id/star`.
- **Comment Count Badge**: Displays `comment_count` inline in file row when > 0.
- **Expanded Preview Dialog**: Rebuilt into 4-tab layout (Details / Comments / Shares / Knowledge).
  - **Details**: Metadata grid (uploader, date, preview kind, comment/share counts) + tags chips.
  - **Comments**: Threaded comment list + composer (Cmd+Enter to send). Calls `GET/POST /api/v1/files/:id/comments`.
  - **Shares**: Channel share history + Share-to-Channel dialog (channel picker + optional message). Calls `GET/POST /api/v1/files/:id/share`.
  - **Knowledge**: View/edit `source_kind`, `knowledge_state`, `summary`, `tags`. Wiki + Ready badges. Calls `PATCH /api/v1/files/:id/knowledge`.
- **Types**: Added `FileComment`, `FileShare` to `types/index.ts`; added `comment_count`, `share_count`, `starred`, `tags`, `knowledge_state`, `source_kind`, `summary` to `FileAsset`.
- **Store**: Added 7 new actions to `file-store.ts`: `fetchFileComments`, `createFileComment`, `toggleFileStar`, `fetchStarredFiles`, `shareFile`, `fetchFileShares`, `updateFileKnowledge`.

### Collab request for Codex (Phase 43)

- Message-level File Attachment viewer: render a shared file as a rich card inline in channel message thread.
- Enrich attachment payload or add `GET /api/v1/messages/:id/files`.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.81] - 2026-04-21

This release implements Phase 42: File Collaboration And Knowledge Metadata APIs, making files first-class collaborative knowledge objects.

### Added

- **File Comments API**: Added `GET /api/v1/files/:id/comments` and `POST /api/v1/files/:id/comments`.
- **File Share API**: Added `GET /api/v1/files/:id/shares` and `POST /api/v1/files/:id/share` for sharing a file into a channel or thread with an attached message.
- **File Stars API**: Added `POST /api/v1/files/:id/star` and `GET /api/v1/files/starred`.
- **Knowledge Metadata API**: Added `PATCH /api/v1/files/:id/knowledge` with `knowledge_state`, `source_kind`, `summary`, and `tags`.
- **Hydrated File Fields**: File detail/list payloads now expose `comment_count`, `share_count`, `starred`, and structured `tags`.

### Knowledge Positioning

- Files can now be marked as `wiki` or other source kinds and carry summary/tags for future LLM and wiki retrieval flows.
- File share operations create a real message attachment record so file knowledge can surface in conversations, not just in the archive.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestFile(CollaborationCommentsStarsAndKnowledgeMetadata|ShareCreatesAttachmentMessageAndAuditEvent)$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.80] - 2026-04-21

This release implements Phase 41: Agent-Collab Payload Simplification. Windsurf consumes Codex's hardened backend contracts directly, eliminating manual parsing workarounds.

### Changed

- **`extractTools()` helper**: Member tool arrays now prefer `primary_tools_array` (native `[]string` from v0.5.79 API) when present. `parsePrimaryTools` string-split is retained as fallback for older API binary or offline static data.
- **`comm_log.to` is now always present** in the API payload (Codex: `""` for broadcasts, `"Name"` for direct). `groupCommLog` in `collab-store.ts` already handled this shape correctly — no changes needed.
- **Static fallback data updated**: `ACTIVE_SUPERPOWERS`, `TASKS` (added t86–t88), and `COMM_SECTIONS` (Phase 41 + Phase 40 sections) reflect current state.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.79] - 2026-04-21

This release implements Phase 41: Agent-Collab Contract Hardening, closing Windsurf's live hub payload gaps.

### Changed

- **Comm Log Targets**: `comm_log` entries now always include a `to` field. Direct messages parsed from `A → B` Markdown bullets return `to: "B"`, while broadcasts return `to: ""`.
- **Member Tool Arrays**: `GET /api/v1/agent-collab/members`, `GET /api/v1/agent-collab/snapshot`, and `agent_collab.sync` now include `primary_tools_array` beside the existing `primary_tools` string for each member.
- **Backward Compatibility**: Existing `primary_tools` and flat `comm_log` payloads remain unchanged, so Windsurf's current client fallback logic continues to work.

### Fixed

- **Agent-Collab Status Encoding**: Corrected the malformed Phase 40 status marker in `docs/AGENT-COLLAB.md`.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/agentcollab -run 'TestParseMarkdownReturnsDirectCommTargetsAndToolArrays$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestAgentCollabMembersAndCommLogEndpoints$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.78] - 2026-04-21

This release implements Phase 40: Agent-Collab Dynamic Hub Integration. Windsurf wires the #agent-collab hub to Codex's live backend APIs, replacing static data while keeping offline fallback.

### Added

- **Live Data Integration**: `collab-store.ts` extended with `members: LiveMember[]`, `commLog: LiveCommSection[]`, `isLive: boolean`, `fetchMembers()`, and `postCommLog()`.
- **Live/Static Badge**: Channel header now shows a pulsing green **Live** badge when connected to the backend, or a grey **Static** badge when using fallback data.
- **Live Members**: `GET /api/v1/agent-collab/members` feeds member profiles into the Overview tab. Visual attributes (avatar, colors) are merged from static `MEMBER_MAP` for known members.
- **Live Comm Log**: Snapshot and `agent_collab.sync` `comm_log` payload feeds the Comm Log tab with real-time entries.
- **Live Superpowers**: Active superpower state on member cards sourced from snapshot `active_superpowers`.

### Fixed

- **`parsePrimaryTools`**: API returns `primary_tools` as a comma-separated string; added client-side splitter to convert it to `string[]`.
- **`groupCommLog`**: API returns `comm_log` as a flat array of messages; added grouping by `title` + `date` to render as structured sections in the Comm Log tab.

### Notes for Codex (Phase 41 request)

- `primary_tools` is handled client-side; optionally normalize to `[]string` server-side.
- `comm_log` entries currently lack a `to` field — all show as Broadcast. Please add `to` to the payload so From→To DM-style messages render correctly.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.77] - 2026-04-21

This release implements Phase 40: Dynamic Agent-Collab APIs, turning Windsurf's new #agent-collab hub into a backend-backed collaboration surface.

### Added

- **Agent-Collab Members API**: Added `GET /api/v1/agent-collab/members`, parsed from the `Member Profiles` table in `docs/AGENT-COLLAB.md`.
- **Agent-Collab Comm Log API**: Added `POST /api/v1/agent-collab/comm-log` to persist new communication log entries back into `docs/AGENT-COLLAB.md`.
- **Expanded Snapshot Payload**: `GET /api/v1/agent-collab/snapshot` and `agent_collab.sync` now include `members` and `comm_log` in addition to `active_superpowers` and `task_board`.
- **Realtime Hub Refresh**: Creating a comm-log entry broadcasts `agent_collab.sync` on `ch-collab`, allowing Windsurf's hub to update without a manual refresh.

### API Contract

- `GET /api/v1/agent-collab/members` returns `{ "members": MemberProfile[] }`.
- `POST /api/v1/agent-collab/comm-log` accepts `{ "from": string, "to"?: string, "title": string, "content": string }`.
- `POST /api/v1/agent-collab/comm-log` returns `{ "entry": CommLogEntry }`.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/agentcollab`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestAgentCollabMembersAndCommLogEndpoints$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.76] - 2026-04-21

This release implements Phase 39: Agent-Collab Hub Page, a comprehensive team collaboration dashboard built by Windsurf. Windsurf joins the team as Web/UI Agent.

### Added

- **Agent-Collab Hub Page**: New dedicated page at `#agent-collab` channel (`web/components/agent-collab/`) replacing the generic channel view with a rich team dashboard.
- **Kanban Board**: Full task board with 85 tasks searchable and filterable by assignee and category (API/Frontend/Infra/UX), grouped by date within the Done column.
- **Communication Log**: Chronological message log with `From` / `From → To` distinction, automatic `@mention` highlighting for all team members, HTTP method badges, and endpoint detection.
- **Statistics Tab**: Daily velocity bar chart, per-category task breakdown, and per-contributor progress bars.
- **Overview Tab**: Live stats bar, team member profile cards (tasks done/ready, active superpower, tools), phase timeline, and assignee breakdown.
- **New Team Member — Windsurf**: Web/UI Agent specializing in Component Architecture, TypeScript, UX Flows, and Agent Collaboration UI. Takes over part of Gemini's frontend scope.

### Fixed

- **Scroll**: Fixed missing `flex-1 min-h-0` on the root div of `AgentCollabPage`, which caused content to overflow without a scrollbar.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.75] - 2026-04-21

This release implements Phase 38: Artifact Duplicate/Fork Integration, enabling users to copy knowledge objects across the workspace and fork specific history versions.

### Added

- **Artifact Duplication**: Integrated `POST /api/v1/artifacts/:id/duplicate` across multiple surfaces. Users can now copy canvases from the toolbar, header, and home dashboard.
- **Canvas Forking**: Added a "Fork as new" action when viewing specific versions in the History panel, allowing users to bootstrap new canvases from any historical state.
- **Contextual Actions**: Added DropdownMenus to artifact cards in the channel header and message attachments for quick duplication and discovery.

### Fixed

- **UI Robustness**: Improved button group layout in the Canvas panel header to accommodate multi-action version previews.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.74] - 2026-04-21

This release implements Phase 38: Artifact Duplicate/Fork APIs, closing a canvas workflow gap for copying an existing artifact into the same or another channel.

### Added

- **Artifact Duplicate API**: Added `POST /api/v1/artifacts/:id/duplicate` for canvas duplicate/fork flows.
- **Target Channel Override**: The duplicate request accepts optional `channel_id` so a canvas can be copied into another channel.
- **Title Override**: The duplicate request accepts optional `title`; otherwise the backend creates a copy title.
- **Version Snapshot**: Duplicated artifacts start at `version: 1` and receive an initial persisted version snapshot.
- **Realtime Sync**: Duplicate creation broadcasts `artifact.updated` so open canvas panels can refresh.

### API Contract

- Request: `POST /api/v1/artifacts/:id/duplicate`
- Body: `{ "channel_id"?: string, "title"?: string }`
- Response: `{ "artifact": Artifact }`
- Behavior: preserves source type/content/template/provider/model, resets status to `draft`, sets `source` to `duplicate`, and uses a new prefixed UUID artifact ID.

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'TestArtifactCRUDAndAI_generate$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.73] - 2026-04-21

This release implements Phase 37: Home Contract and Draft Lifecycle integration, hardening the home data flow and enabling explicit draft cleanup.

### Added

- **Draft Lifecycle Cleanup**: Added `deleteDraft` to `DraftStore` consuming the new `DELETE /api/v1/drafts/:scope` endpoint.
- **Explicit Cleanup**: The message composer now explicitly removes persisted drafts from the backend upon successful message delivery or when the input is cleared.
- **Hardened Home UI**: Transitioned `HomeDashboard` to consume top-level hardened aliases (`stats`, `recent_activity`, `recent_artifacts`) directly from the `/api/v1/home` payload.

### Fixed

- **Clean Lint State**: Pruned unused `cn` and `formatDistanceToNow` imports in the dashboard component.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.71] - 2026-04-21


This release hardens the home payload for the current dashboard UI and adds explicit draft deletion.

### Added

- **Home Dashboard Compatibility Fields**: `GET /api/v1/home` now returns:
  - `stats`
  - `recent_activity`
  - top-level `recent_artifacts`
- **Explicit Draft Cleanup**: Added `DELETE /api/v1/drafts/:scope` so channel, DM, and thread draft state can be cleared instead of only overwritten.

### Changed

- home payload now exposes dashboard-oriented aliases without removing the existing `activity`, `profile`, `recent_lists`, `recent_tool_runs`, or `recent_files` contracts
- `stats.pending_actions` is aligned with unread work, while `stats.active_threads` summarizes current thread activity across the user’s channels
- `recent_activity` now gives the Home UI channel-aware recent conversation cards with `channel_id`, `channel_name`, and `last_message`

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'Test(DeleteDraftRemovesScopedDraft|GetHomeReturnsWorkspaceSummary)$'`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.71] - 2026-04-21

This release fixes a critical runtime error and improves navigation reliability.

### Fixed

- **Home Dashboard Crash**: Fixed a `ReferenceError: Clock is not defined` in `HomeDashboard` by restoring the missing Lucide icon import.
- **Navigation Sync**: Enhanced the `WorkspacePage` to automatically clear the active channel state when the `c` URL parameter is missing. This ensures the Home Dashboard correctly renders when navigating back from a specific channel.

### Verification Used For This Release

- `pnpm lint` (Verified PASS)

## [0.5.70] - 2026-04-21

This release refines navigation persistence and AI context precision.

### Fixed

- **Home Navigation Persistence**: Reimplemented the Home sidebar button to explicitly clear the active channel state. This ensures users can always return to the Home Dashboard even after browsing specific channels.
- **AI Context Precision**: Updated the "Summarize this channel" suggestion to dynamically inject the active channel's name and ID into the prompt (e.g., `Summarize the #general channel (ID: ch_1)`). This prevents the AI from incorrectly referencing external platforms like YouTube and locks it to the workspace context.
- **UI Cleanup**: Pruned unused icons and redundant store imports in the Dashboard and Navigation components.

### Verification Used For This Release

- `pnpm lint` (Verified PASS)

## [0.5.69] - 2026-04-21

This release addresses accessibility warnings and improves external asset reliability.

### Fixed

- **Dialog Accessibility**: Added missing `DialogDescription` components to `UserProfile`, `FilesPage`, `PeopleDirectoryPage`, `WorkflowsPage`, and `ChannelInfo`. This resolves the browser warning "Missing Description for DialogContent".
- **External Asset Reliability**: Configured `unoptimized: true` for GitHub and Avatar domains in `next.config.mjs` to resolve the "resolved to private ip" error, ensuring avatars load correctly in all environments.

### Verification Used For This Release

- `pnpm lint` (Verified PASS)

## [0.5.68] - 2026-04-21

This release improves the responsiveness of the channel information sidebar.

### Fixed

- **Scrollable Channel Tabs**: Updated the `TabsList` in `ChannelInfo` to support horizontal scrolling when multiple tabs (About, Members, Pins, Files, Settings) overflow the container width, ensuring all actions remain accessible.

### Verification Used For This Release

- `pnpm lint` (Verified PASS)

## [0.5.67] - 2026-04-21

This release implements Phase 35: Structured Work Aggregation and Phase 36: ID Normalization Verification.

### Added

- **Home Dashboard Aggregation**: The workspace home now features dedicated sections for `Recent Lists`, `Recent Automations`, and `Latest Files`, providing a unified view of structured work.
- **Enhanced Activity Feeds**: Structured event types (`list_completed`, `tool_run`, `file_uploaded`) are now rendered with custom iconography, status badges, and "View Context" navigation links.
- **Smart Navigation**: Activities now link directly to the Workflows and Files surfaces, closing the loop between chat signals and operational objects.

### Fixed

- **ID Normalization**: Verified the frontend consistently treats all primary keys as opaque strings, supporting the transition to prefixed UUIDs across all core models.
- **Pin Scoping**: Confirmed that pinned messages are correctly scoped to the active channel in the `ChannelInfo` panel.
- **Clean Lint State**: Resolved all remaining unused import and variable warnings across the web app.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.64] - 2026-04-21


This release normalizes generated string primary keys to UUID-style IDs and fixes channel-scoped pin retrieval.

### Added

- **Unified ID Generator**: New runtime ID generation now uses prefixed UUID-style values for newly created:
  - channels
  - user groups
  - workflow runs
  - lists
  - tool runs
  - files
  - artifacts
  - AI conversations
  - AI conversation messages

### Fixed

- **Channel Pins Filtering**: `GET /api/v1/pins?channel_id=...` now correctly filters by channel instead of returning global pinned messages.
- **Channel URL Ergonomics**: Newly created channels no longer use timestamp-shaped IDs such as `ch-20260421062955.353728`.

### Changed

- newly created string primary keys now follow a consistent `prefix-uuid` contract instead of timestamp concatenation
- storage paths for uploaded files continue to work with the new file IDs because filenames are derived from the normalized prefixed UUID
- collaboration notes now warn frontend consumers not to make assumptions about timestamp-shaped channel IDs

### Verification Used For This Release

- `cd apps/api && go test ./internal/handlers -run 'Test(ExecuteAIStreamsSSE|CreateChannelCreatesChannelAndOwnerMembership|GetPinsReturnsPinnedMessagesWithChannelAndUser|WorkspaceListLifecycleEndpoints|ToolExecutionEndpoints|ArtifactLifecycleAndVersionHistory|FileUploadAndListing|CreateUserGroupLifecycleAndDeletion|WorkflowRunsCanBeCreatedAndListed)$'`
- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.65] - 2026-04-21

This release implements Phase 35: Structured Work Aggregation, connecting lists, tool runs, and files into the Slack-like home and activity surfaces.

### Added

- **Home Structured Work Aggregation**: `GET /api/v1/home` now includes:
  - `recent_lists`
  - `recent_tool_runs`
  - `recent_files`
- **Structured Activity Signals**: `GET /api/v1/activity` and `GET /api/v1/inbox` now include:
  - `list_completed`
  - `tool_run`
  - `file_uploaded`
- **Channel-Aware Home Hydration**: Home-level tool run and file summaries now respect current-user channel membership in addition to direct ownership.

### Changed

- checklist completion now surfaces in the shared activity model used by Activity and Inbox
- tool execution completions now appear as first-class workspace events instead of only isolated detail records
- channel file uploads now flow into notification-style feeds for members of that channel
- collaboration docs now hand Gemini the next frontend integration pass for home/activity structured work widgets

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.64] - 2026-04-21

This release implements Phase 34: Structured Contract Alignment, cleaning up frontend-side fallbacks and hardening the integration with hardened backend aliases for lists, tools, and virtual artifacts.

### Added

- **Virtual Artifact Bootstrap**: Removed local `new-doc` stub hydration in favor of the hardened `GET /api/v1/artifacts/new-doc` backend endpoint, ensuring consistent ownership and template metadata.
- **Contract Hardening**: Stores now consume camelCase aliases (`userId`, `channelId`, `finishedAt`, `durationMs`) directly from hardened backend payloads.
- **Harden Home Mapping**: Updated `WorkspaceStore` to support the new `home` response structure while maintaining backward compatibility.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.62] - 2026-04-21


This release hardens the Phase 33 contracts so Gemini's integrated list, tool-run, and canvas flows can consume the backend without frontend-side guesswork.

### Added

- **List Compatibility Aliases**: `lists` and list-item payloads now expose UI-friendly aliases including `user_id` and `list_id`, while list creation can derive `workspace_id` from `channel_id`.
- **Tool Run Compatibility Aliases**: `tool runs` now expose `user_id`, `channel_id`, `finished_at`, and `duration_ms`, and the list endpoint now supports `channel_id` filtering.
- **Canvas Bootstrap Compatibility**: virtual `new-doc` and template-derived artifact responses now include `user_id` aliases for the frontend artifact store.

### Changed

- `POST /api/v1/lists` now accepts channel-first creation flows from the UI without requiring explicit `workspace_id`.
- `POST /api/v1/tools/:id/execute` now lifts top-level `channel_id` into stored execution input for downstream filtering and UI hydration.
- `AGENT-COLLAB.md` now reflects that Phase 33 frontend integration is complete and records a Phase 34 contract-alignment handoff.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.62] - 2026-04-21

This release implements Phase 33: Structured Work Objects Integration, bridging the gap between chat, structured lists, and automation tools.

### Added

- **Workspace Lists**: New `ListStore` and integration for first-class list objects with item-level completion and assignment.
- **Tool Run History**: New `ToolStore` and detail cards for viewing execution logs and results of background tool runs.
- **Template-First Canvases**: Integrated a new "New Canvas" overlay in `CanvasPanel` with template selection and blank document bootstrap.
- **Template Bootstrap**: Consumption of `GET /api/v1/artifacts/templates` and `POST /api/v1/artifacts/from-template` for faster knowledge creation.

### Fixed

- **Centralized Types**: Relocated `FileAsset` and `WorkflowRun` to global types for better consistency across stores.
- **UI Robustness**: Added missing icon and store imports in `CanvasPanel`.
- **Naming Conflicts**: Renamed internal `Plus` component to `PlusBadge` to avoid Lucide icon collisions.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.60] - 2026-04-21


This release starts the next structured-work wave for Relay, adding lightweight lists, tool execution history, and template-first canvas creation.

### Added

- **Workspace Lists API**: Added `GET/POST/PATCH/DELETE /api/v1/lists` plus item mutation endpoints for shared checklists and operational tracking.
- **Tool Execution APIs**: Added `GET /api/v1/tools/runs`, `GET /api/v1/tools/runs/:id`, and `POST /api/v1/tools/:id/execute` for visible tool history and manual execution.
- **Artifact Template APIs**: Added `GET /api/v1/artifacts/templates`, `POST /api/v1/artifacts/from-template`, and virtual `GET /api/v1/artifacts/new-doc` support.
- **Structured Work Seeds**: Added sample list data and sample tool execution logs to the runtime seed set.

### Changed

- `artifacts` and `artifact_versions` now persist `template_id` metadata for template-derived canvases.
- README, phase targets, and collaboration docs now treat lists, tool runs, and templates as first-class backend capabilities.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.60] - 2026-04-21

This release implements Phase 32: Operational Shell Controls, bridging operational placeholders with real backend actions for workflows, files, and channels.

### Added

- **Workflow Transparency**: Added raw execution log viewing via `GET /api/v1/workflows/runs/:id/logs`.
- **Workflow History Management**: Enabled log deletion for finished workflow runs.
- **Richer File Previews**: Integrated a new preview dialog consuming `GET /api/v1/files/:id/preview` with metadata and visual previews.
- **Channel Preferences**: Connected the Notifications control in `ChannelInfo` to `GET/PATCH /api/v1/channels/:id/preferences`.
- **Channel Membership**: Wired the "Leave channel" action to `POST /api/v1/channels/:id/leave` with a confirmation flow.

### Fixed

- **UI Robustness**: Added missing `Loader2`, `Badge`, and `FileText` components to various surfaces.
- **Next.js Compliance**: Migrated file preview `<img>` tags to `next/image`.
- **Clean State**: Ensured 100% clean lint and production build stability.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.58] - 2026-04-21


This release implements Phase 32: Operational Shell Controls, extending Slack-style admin and automation operations for workflows, files, and channels.

### Added

- **Workflow Run Logs API**: Added `GET /api/v1/workflows/runs/:id/logs`, returning ordered execution logs with `level`, `message`, parsed `metadata`, and `created_at`.
- **Workflow Run Delete API**: Added `DELETE /api/v1/workflows/runs/:id`, deleting the run plus associated steps/logs and broadcasting `workflow.run.deleted` over WebSocket.
- **Workflow Log Persistence**: Added `workflow_run_logs` storage and seeded sample run logs for the workflow run detail surface.
- **File Preview Metadata API**: Added `GET /api/v1/files/:id/preview`, returning `file_id`, `name`, `content_type`, `preview_kind`, `preview_url`, `download_url`, `is_previewable`, `size`, `channel_id`, uploader metadata, and expiry metadata.
- **Channel Preferences API**: Added `GET /api/v1/channels/:id/preferences` and `PATCH /api/v1/channels/:id/preferences` with `notification_level` values of `all`, `mentions`, or `none`, plus `is_muted`.
- **Leave Channel API**: Added `POST /api/v1/channels/:id/leave`, removing the current user from channel membership and recomputing `member_count`.

### Changed

- Auto-migration now includes `WorkflowRunLog` and `ChannelPreference`.
- Workflow create/cancel/retry flows now append operator-readable workflow logs.
- README current release line now reflects the latest published backend/API baseline.

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.58] - 2026-04-21

This release implements Phase 31: Contract Hardening, enriching profile surfaces and stabilizing backend data contracts for workflows and files.

### Added

- **Richer Personal Profiles**: Added support for `pronouns`, `location`, `phone`, and `bio` fields. These are now editable and visible on profile cards.
- **UI-Friendly Field Aliases**: Optimized `FileStore` and `WorkflowStore` to consume hardened backend fields (`type`, `size`, `workflowName`, `durationMs`, etc.) directly.
- **Enhanced Audit Logs**: Support for both `events` and `audit_history` aliases in the file audit surface, ensuring compatibility with upcoming backend changes.

### Fixed

- **Centralized Types**: Moved `FileAsset` and other core models to the centralized `types/index.ts`.
- **UI Polish**: Fixed incorrect icon usage in the user profile details grid.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.56] - 2026-04-21


This release hardens the operational shell contracts behind profiles, workflows, and files so the current UI can consume richer data without store-side guesswork.

### Added

- extended user profile fields:
  - `pronouns`
  - `location`
  - `phone`
  - `bio`
- `workflow_run_steps`
- richer workflow run detail hydration with:
  - `workflow_name`
  - `triggered_by`
  - `finished_at`
  - `duration_ms`
  - `error`
  - `steps`

### Expanded

- `PATCH /api/v1/users/:id` now supports:
  - `pronouns`
  - `location`
  - `phone`
  - `bio`
- `GET /api/v1/users/:id` now returns the same extended identity fields
- `GET /api/v1/workflows/runs`
- `GET /api/v1/workflows/runs/:id`
  now include flat compatibility fields alongside nested workflow data
- `GET /api/v1/files`
- `GET /api/v1/files/:id`
  now include UI-friendly aliases:
  - `type`
  - `size`
  - `userId`
  - `channelId`
  - `createdAt`
- `GET /api/v1/files/:id/audit` now returns both:
  - `events`
  - `audit_history`

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`

## [0.5.56] - 2026-04-21

This release implements Phase 30: Operational Maturity, adding governance, automation control, and richer presence features.

### Added

- **Richer Custom Status**: Added support for status emojis and auto-expiration durations (Select from 30m, 1h, 4h, etc.).
- **User Group Membership**: Built a dedicated member editor for user groups with add/remove support and role visualization.
- **Group Mentions**: Integrated user groups into the mention lookup popover in the message composer.
- **File Governance**: Added file audit logs and configurable retention policies to the Files surface.
- **Workflow Run Lifecycle**: Added detailed execution step views, cancellation, and retry logic for automation workflows.
- **Realtime Automation**: Wired WebSocket events to trigger live UI updates for workflow status changes.

### Fixed

- **Store Hydration**: Fixed missing field mappings for workflow runs and user status expiry.
- **Clean Build**: Verified production build stability in ~9s.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.52] - 2026-04-21


This release finishes the CI packaging repair for the Phase 30 operational maturity wave.

### Fixed

- declared `@typescript-eslint/eslint-plugin` directly in `apps/web/package.json`
- declared `@typescript-eslint/parser` directly in `apps/web/package.json`
- aligned workspace package versions to the new release line

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm build` still hangs after `Creating an optimized production build ...` in this environment and remains a frontend investigation item

## [0.5.54] - 2026-04-21

This release republishes the Phase 30 operational maturity work with a packaging fix for GitHub Actions.

### Fixed

- declared `@next/eslint-plugin-next` directly in `apps/web/package.json` so CI and Release runners resolve the flat ESLint config correctly
- aligned workspace package versions to the new release line

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm build` still hangs after `Creating an optimized production build ...` in this environment and remains a frontend investigation item

## [0.5.53] - 2026-04-21

This release expands Relay's operational maturity with richer status controls, user-group membership APIs, file lifecycle governance, and workflow run control surfaces.

### Added

- `GET /api/v1/user-groups/mentions`
- `GET /api/v1/user-groups/:id/members`
- `POST /api/v1/user-groups/:id/members`
- `DELETE /api/v1/user-groups/:id/members/:userId`
- `GET /api/v1/workflows/runs/:id`
- `POST /api/v1/workflows/runs/:id/cancel`
- `POST /api/v1/workflows/runs/:id/retry`
- `PATCH /api/v1/files/:id/retention`
- `GET /api/v1/files/:id/audit`

### Expanded

- `PATCH /api/v1/users/:id/status` now supports:
  - `status_emoji`
  - `expires_in_minutes`
- `GET /api/v1/files/:id` now returns:
  - `preview_kind`
  - `preview_url`
  - `retention_days`
  - `expires_at`
- `workflow.run.updated` is now emitted for:
  - run creation
  - cancellation
  - retry

### Data Model

- richer `users` status fields:
  - `status_emoji`
  - `status_expires_at`
- richer `workflow_runs` field:
  - `retry_of_run_id`
- richer `file_assets` fields:
  - `description`
  - `retention_days`
  - `expires_at`
  - `updated_at`
- new `file_asset_events`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm build` still hangs after `Creating an optimized production build ...` in this environment and remains a frontend investigation item

## [0.5.52] - 2026-04-21

This release implements Phase 29: Admin and Realtime Integration, enhancing user profiles, group management, and asset lifecycles.

### Added

- **Profile Editing**: Users can now edit their title, department, and timezone directly from their profile card.
- **Group Management**: Full CRUD support for user groups, including handle-based naming and member counting.
- **File Deletion**: Enabled permanent deletion for files and assets across the workspace.
- **Richer File Filters**: Added uploader and content-type filtering to the Files page.
- **Realtime Automations**: Integrated websocket listeners for live workflow run updates.

### Fixed

- **Signature Consistency**: Updated all `fetchFiles` callers to use the new object-based parameter pattern.
- **UI Robustness**: Resolved syntax errors in the People page and ensured clean build states across the Next.js 16 toolchain.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.50] - 2026-04-21


This release expands Relay's admin and operational shell with profile editing, user group CRUD, richer file management, and workflow run realtime.

### Added

- `PATCH /api/v1/users/:id`
- `POST /api/v1/user-groups`
- `PATCH /api/v1/user-groups/:id`
- `DELETE /api/v1/user-groups/:id`
- `DELETE /api/v1/files/:id`

### Expanded

- `GET /api/v1/files` now supports:
  - `uploader_id`
  - `content_type`
  - `is_archived`
- `GET /api/v1/files/archive` now supports:
  - `uploader_id`
  - `content_type`
- `POST /api/v1/workflows/:id/runs` now emits realtime `workflow.run.updated`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm build` still hangs after `Creating an optimized production build ...` in this environment and needs a fresh frontend investigation

## [0.5.50] - 2026-04-21

This release implements Phase 28: Operational Shell Integration, providing core administrative and organizational surfaces.

### Added

- **People Directory**: New people directory page with multi-parameter filtering (search, department, status, group).
- **Files & Archive**: Dedicated files management page with archive/restore lifecycle support.
- **Notification Preferences**: Persistent notification settings page with granular controls for inbox, mentions, and mutes.
- **Workflow Orchestration**: Automation surface for triggering workflows and viewing run history.
- **Navigation Update**: Expanded primary navigation with direct access to People, Files, and Workflows.

### Fixed

- **Build Stability**: Resolved hanging production builds in Next.js 16 by fixing missing dependencies (`@radix-ui/react-select`, `@radix-ui/react-tabs`) and invalid TipTap configurations.
- **UI Consistency**: Migrated settings switches to standard Radix UI primitives.
- **Clean Lint**: Restored 100% clean lint state with no warnings.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.48] - 2026-04-20


This release expands the Slack-parity shell into more operational workspace surfaces: directory filtering, notification preferences, file archive lifecycle, and workflow run history.

### Added

- `GET /api/v1/workflows/runs`
- `POST /api/v1/workflows/:id/runs`
- `GET /api/v1/notifications/preferences`
- `PATCH /api/v1/notifications/preferences`
- `GET /api/v1/files/archive`
- `PATCH /api/v1/files/:id/archive`

### Expanded

- `GET /api/v1/users` now supports:
  - `q`
  - `department`
  - `status`
  - `timezone`
  - `user_group_id`

### Data Model

- `workflow_runs`
- `notification_preferences`
- `notification_mute_rules`
- richer `file_assets` fields:
  - `is_archived`
  - `archived_at`

### Integration Fixes

- aligned `DirectoryStore` with backend `groups/group` payload keys
- aligned `WorkspaceStore` and `HomeDashboard` with the `GET /api/v1/home` response shape

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm --filter relay-agent-workspace lint`
- `pnpm build` still enters Next.js 16 optimized production build and does not exit in this environment; tracked for Gemini follow-up
## [0.5.48] - 2026-04-20

This release fixes a technical conflict in the message editor configuration.

### Fixed

- **TipTap Duplicate Extension**: Resolved the "Duplicate extension names found: ['link']" warning by explicitly disabling the default link extension in `StarterKit` in favor of our custom-styled version.

### Verification Used For This Release

- `pnpm lint` (Verified PASS)

## [0.5.47] - 2026-04-20

This release focuses on stability, engineering standards, and codebase cleanup after the Next.js 16 upgrade.

### Fixed

- **Store Syntax Errors**: Resolved critical parsing errors in `UserStore` and `WorkspaceStore` caused by placeholder leaks.
- **Lint Sanitization**: Achieved zero errors and zero warnings across the entire web application.
- **Redundancy Cleanup**: Removed duplicate `HomeDashboard` implementation in `WorkspacePage` and consolidated into a reusable component.
- **Type Safety**: Fixed a missing field in `SearchStore` initialization that caused build failures.
- **Image Optimization**: Migrated remaining native `<img>` tags to `next/image` for better performance and LCP.
- **Dead Code Removal**: Pruned over 10 unused variables, imports, and parameters across stores and hooks.

### Verification Used For This Release

- `pnpm run build` (Verified PASS)
- `pnpm lint` (Verified PASS)

## [0.5.45] - 2026-04-20


This release expands the Slack-parity foundation beyond channels and messages into home, people, groups, status, tools, and workflow surfaces.

### Added

- `GET /api/v1/home`
- `GET /api/v1/users/:id`
- `PATCH /api/v1/users/:id/status`
- `GET /api/v1/user-groups`
- `GET /api/v1/user-groups/:id`
- `GET /api/v1/workflows`
- `GET /api/v1/tools`

### Data Model

- `user_groups`
- `user_group_members`
- `workflow_definitions`
- `tool_definitions`
- richer `users` fields for:
  - `title`
  - `department`
  - `timezone`
  - `working_hours`

### Behavior Notes

- `GET /api/v1/home` returns a Slack-style landing payload with:
  - `user`
  - `profile`
  - `activity`
  - `starred_channels`
  - `recent_dms`
  - `drafts`
  - `tools`
  - `workflows`
- `GET /api/v1/users/:id` now returns a hydrated personal profile with local time, working hours, focus areas, top channels, and recent artifacts
- `PATCH /api/v1/users/:id/status` persists `status` and `status_text`, refreshes presence expiry, and emits websocket `presence.updated`
- user groups now expose both list and detail contracts, including hydrated member users
- workflows and tools are now explicit backend registries instead of UI-only assumptions

### Documentation

- added [docs/phases/phase-10-home-profiles-groups-workflows.md](./docs/phases/phase-10-home-profiles-groups-workflows.md)
- added [docs/releases/v0.5.46.md](./docs/releases/v0.5.46.md)
- updated `docs/AGENT-COLLAB.md` with the next Gemini handoff for home, people, user groups, and workflow surfaces

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.5.44] - 2026-04-20

This release implements Phase 26: Intelligent Search and Artifact Backlinks, deepening the connectivity and discoverability of knowledge within the workspace.

### Added

- **Artifact Backlinks**: New "Referencing Messages" sidebar in `CanvasPanel` to see where an artifact is mentioned in discussions.
- **Intelligent Ranked Search**: Integrated AI-native ranked search results at the top of the search palette.
- **Realtime Read State Sync**: Leveraged websocket `notifications.read` to synchronize notification read state across all open client windows.
- **Search Score Badges**: Ranked search results now display their relevance score and match reason.

### Fixed

- **Navigation**: Intelligent search results correctly navigate to channels, users, or artifacts.
- **Store Reconciliation**: Added `markAsReadLocally` to `ActivityStore` for consistent multi-device read states.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.42] - 2026-04-20


This release adds the first ranked search layer and cross-surface backlink/read-sync primitives.

### Added

- `GET /api/v1/artifacts/:id/references`
- `GET /api/v1/search/intelligent?q=...`
- realtime `notifications.read` event broadcast from `POST /api/v1/notifications/read`

### Search Contract

- intelligent search returns ranked typed items with:
  - `type`
  - `id`
  - `label`
  - `reason`
  - `score`

### Documentation

- added [docs/phases/phase-10-intelligent-search-backlinks-and-notification-sync.md](./docs/phases/phase-10-intelligent-search-backlinks-and-notification-sync.md)
- added [docs/releases/v0.5.43.md](./docs/releases/v0.5.43.md)
- updated `docs/AGENT-COLLAB.md` with Gemini handoff for backlinks, ranked search, and notification realtime sync

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.42] - 2026-04-20

This release implements Phase 25: Knowledge References and Knowledge Search, enabling deep linking and discovery of artifacts and files across the workspace.

### Added

- **Message Attachments**: Messages can now include artifact references and file attachments, rendered inline with rich previews.
- **Composer Integration**: The message composer now correctly tracks uploaded files and attaches them to the outgoing message.
- **Knowledge Search**: Expanded the global search to include dedicated sections for artifacts and files, complete with typeahead suggestions.
- **Smart Suggestions**: Global search suggestions now include matching artifacts and files for faster discovery.

### Fixed

- **Type Safety**: Updated `Message` and `Attachment` types to support rich hydrated artifact and file metadata.
- **UI Consistency**: Standardized the rendering of files and artifacts across the channel introduction, message items, and search results.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.40] - 2026-04-20


This release expands Relay's knowledge graph across messages, artifacts, files, and search.

### Added

- `POST /api/v1/messages` now accepts optional:
  - `artifact_ids`
  - `file_ids`
- `GET /api/v1/messages` and `GET /api/v1/messages/:id/thread` now hydrate linked artifacts and files into `metadata.attachments`
- `GET /api/v1/search` now returns:
  - `results.artifacts`
  - `results.files`
- `GET /api/v1/search/suggestions` now includes typed:
  - `artifact` suggestions
  - `file` suggestions

### Data Model

- `message_artifact_references`
- `message_file_attachments`

### Documentation

- added [docs/phases/phase-10-knowledge-references-and-search-expansion.md](./docs/phases/phase-10-knowledge-references-and-search-expansion.md)
- added [docs/releases/v0.5.41.md](./docs/releases/v0.5.41.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for message references and search expansion

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.40] - 2026-04-20

This release completes Phase 24: Artifact Restore and Enhanced History, providing a robust workflow for document version control and visualization.

### Added

- **Official Version Restore**: Wired the "Restore this version" button to the new `POST /api/v1/artifacts/:id/restore/:version` endpoint.
- **Structured Diff Spans**: Integrated the new `diff.spans` payload for more accurate and richer diff rendering in the comparison view.
- **Line Number Hints**: Added hoverable line number markers (`L10 → L12`) in the diff view when provided by the backend.
- **Robust History Hydration**: Automatically refreshes the version history and active artifact state after a successful rollback.

### Fixed

- **Diff Fallback**: Improved the diff rendering logic to gracefully fallback to manual `unified_diff` parsing if structured spans are missing.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.38] - 2026-04-20

This release implements Phase 24: Artifact Restore, allowing users to revert collaborative canvases to any previous version in their history.

### Added

- **Version Restore**: Integrated `POST /api/v1/artifacts/:id/restore/:version` to enable official version rollbacks.
- **Structured Diff Mapping**: Updated `artifact-store.ts` to handle the new `diff.spans` field for richer future diff rendering.
- **Restore CTA**: Wired the "Restore this version" button in the `CanvasPanel` history sidebar to the backend restore flow.

### Fixed

- **History Hydration**: After a successful restore, the version history list is automatically refreshed to reflect the new state.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.36] - 2026-04-20


This release upgrades the canvas history backend from view-only history into a real rollback-ready workflow with restore support and structured diff spans.

### Added

- `POST /api/v1/artifacts/:id/restore/:version`
- structured `diff.spans` in `GET /api/v1/artifacts/:id/diff/:from/:to`

### Artifact Contract

- restore responses include:
  - `artifact`
  - `restored_from_version`
- diff responses now include:
  - `spans[].kind`
  - `spans[].content`
  - optional `spans[].from_line`
  - optional `spans[].to_line`

### Documentation

- added [docs/phases/phase-10-artifact-restore-and-structured-diff.md](./docs/phases/phase-10-artifact-restore-and-structured-diff.md)
- added [docs/releases/v0.5.37.md](./docs/releases/v0.5.37.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for canvas restore and richer compare rendering

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.36] - 2026-04-20

This release implements Phase 23: Search Suggestions and Rich Results, significantly improving the search experience with real-time typeahead and enhanced result metadata.

### Added

- **Search Suggestions**: Integrated `GET /api/v1/search/suggestions` to provide instant typeahead suggestions in the search dialog.
- **Message Snippets**: Search results for messages now render highlighted snippets, making it easier to identify relevant context.
- **Match Reasons**: Channels and users in search results now display the reason for the match (e.g., "Matched in description").
- **Unified Search Store**: Expanded `search-store.ts` to manage suggestions and handle richer result payloads.
- **Improved Navigation**: Clicking on search results now correctly navigates to channels or opens docked chats for users.

### Fixed

- **Dialog UX**: Optimized the search input and results list for a smoother command palette experience.
- **Lint Stability**: Resolved all unused variable and unescaped quote errors in the search component.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.34] - 2026-04-20


This release upgrades workspace search from a raw result dump to a more UI-friendly discovery API with suggestions and richer result shaping.

### Added

- `GET /api/v1/search/suggestions?q=...`

### Search Contract

- `GET /api/v1/search` now returns richer result items including:
  - `messages[].snippet`
  - `channels[].match_reason`
  - `users[].match_reason`
- `GET /api/v1/search/suggestions` returns typed suggestion items with:
  - `type`
  - `id`
  - `label`
  - `hint`

### Documentation

- added [docs/releases/v0.5.35.md](./docs/releases/v0.5.35.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for search suggestions

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.34] - 2026-04-20

This release stabilizes the AI-Canvas collaboration flow by aligning the frontend with recent backend command and sentinel improvements.

### Added

- **Command Forwarding**: Explicit slash commands (like `/canvas`) are now forwarded to the AI execute endpoint, skipping unnecessary classification logic.
- **Robust Diff Mapping**: Enhanced `artifact-store.ts` to handle the standardized `{ diff: ... }` response envelope for artifact comparisons.

### Fixed

- **Canvas Creation**: Corrected the save flow for new documents (`new-doc`) to ensure they are created as real artifacts upon first save.
- **Data Integrity**: Improved `mapArtifact` with better null-safety and field mapping for `created_by_user` and `updated_by_user`.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.32] - 2026-04-20


This release closes the loop between the latest Gemini UI work and the backend by removing the `new-doc` artifact 404 path and adding explicit AI command forwarding.

### Added

- `POST /api/v1/ai/execute` now accepts optional `command`

### Fixed

- opening a new canvas no longer triggers `GET /api/v1/artifacts/new-doc`
- saving a new canvas now creates a real artifact instead of trying to patch a sentinel id
- artifact diff store mapping now correctly reads the backend `{ diff: ... }` envelope

### Documentation

- added [docs/releases/v0.5.33.md](./docs/releases/v0.5.33.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for command forwarding and canvas sentinel handling

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.32] - 2026-04-20

This release improves the AI Assistant's responsiveness and stability, specifically around slash commands and panel layout.

### Added

- **Slash Command Filtering**: Integrated dynamic filtering in the `AISlashCommand` menu. Commands now filter in real-time as you type after the slash.
- **Smart /canvas Interception**: The UI now intercepts `/canvas` commands locally to provide instant visual feedback (opening the canvas) while the AI processes the prompt.

### Fixed

- **AI Panel Scrolling**: Resolved a recurring issue where the AI Assistant dialogue wouldn't show a scrollbar by stabilizing the flex-height inheritance.
- **HTML Command Leaks**: Fixed a bug where slash commands entered in the rich editor would leak as raw HTML tags in the message stream.
- **Artifact Store Safety**: Added robust null-checks in `mapArtifact` to prevent client-side crashes on malformed API responses.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.30] - 2026-04-20

This release implements Phase 22: Artifact Diff UI, allowing users to compare different versions of collaborative canvases side-by-side.

### Added

- **Artifact Comparison UI**: New `ArtifactDiffView` component to render unified diffs with line-level highlighting.
- **Compare Mode**: Enhanced `CanvasPanel` with a "Compare Versions" toggle in the history sidebar.
- **Multi-Version Selection**: Users can now select any two historical versions to generate a visual diff.
- **Diff Summary**: Real-time statistics showing the number of lines added and removed between versions.
- **Lazy Diff Loading**: Optimized network usage by fetching diff data only when a specific comparison is requested.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.28] - 2026-04-20


This release adds artifact diff APIs so Relay can compare any two stored canvas versions and power a real visual comparison view.

### Added

- `GET /api/v1/artifacts/:id/diff/:from/:to`

### Diff Contract

- diff responses include:
  - `artifact_id`
  - `from_version`
  - `to_version`
  - `from_content`
  - `to_content`
  - `unified_diff`
  - `summary.added_lines`
  - `summary.removed_lines`

### Documentation

- added [docs/phases/phase-10-artifact-diff.md](./docs/phases/phase-10-artifact-diff.md)
- added [docs/releases/v0.5.29.md](./docs/releases/v0.5.29.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for artifact comparison views

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.28] - 2026-04-20

This release implements Phase 21: Artifact Version History, enabling users to browse and restore previous versions of collaborative canvases.

### Added

- **Version History UI**: Integrated a history sidebar in the `CanvasPanel` to browse all previous versions of an artifact.
- **Version Preview**: Users can now view the content, title, and metadata of any historical version without affecting the current live state.
- **One-Click Restore**: Added a "Restore this version" capability to quickly rollback the active canvas to a prior state.
- **History Metadata**: Each version now displays the timestamp and the user who made the change.
- **Expanded Artifact Store**: Updated `artifact-store.ts` to manage version lists and lazy-loaded version details.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.26] - 2026-04-19


This release adds artifact version history so Relay canvases are now auditable and ready for a real history panel in the UI.

### Added

- `GET /api/v1/artifacts/:id/versions`
- `GET /api/v1/artifacts/:id/versions/:version`
- persisted `artifact_versions` snapshots on artifact create, update, and AI canvas generation

### Artifact Contract

- artifact payloads now include:
  - `version`
- version payloads include:
  - `artifact_id`
  - `version`
  - `title`
  - `type`
  - `status`
  - `content`
  - `updated_by`
  - `updated_by_user`
  - `created_at`

### Documentation

- added [docs/phases/phase-10-artifact-version-history.md](./docs/phases/phase-10-artifact-version-history.md)
- added [docs/releases/v0.5.27.md](./docs/releases/v0.5.27.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for artifact version history

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.26] - 2026-04-19

This release implements Phase 20: Presence Refinements, making user status more dynamic and context-aware.

### Added

- **Presence Heartbeat**: Implemented a 30-second client-side heartbeat to maintain active session presence.
- **Enriched Last Seen**: Added "Last seen" timestamps using `formatDistanceToNow` for offline users.
- **Scoped Presence**: Added automatic presence fetching for channel members when switching channels.
- **Status Metadata**: Integrated `status_text` support in avatars and profiles.

### Fixed

- UI Consistency: Unified presence mapping across `presence-store.ts` and `user-store.ts`.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.24] - 2026-04-19


This release upgrades Relay presence with heartbeat refresh, derived last-seen metadata, and scoped presence lists.

### Added

- `POST /api/v1/presence/heartbeat`
- scoped `GET /api/v1/presence?channel_id=...`

### Presence Contract

- user payloads now include:
  - `status_text`
  - `last_seen_at`
- presence freshness is tracked with heartbeat expiry timestamps
- websocket `presence.updated` payloads use the enriched user structure

### Documentation

- added [docs/phases/phase-10-presence-refinements.md](./docs/phases/phase-10-presence-refinements.md)
- added [docs/releases/v0.5.25.md](./docs/releases/v0.5.25.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for presence refinements

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.24] - 2026-04-19

This release implements Phase 19: File Assets and Artifact Identity, enabling real file uploads and richer collaboration context.

### Added

- **Real File Uploads**: Integrated multipart file upload API in the message composer.
- **File Attachments**: Uploaded files are automatically linked in the editor for instant sharing.
- **Channel Files Tab**: New "Files" tab in the `ChannelInfo` panel to browse and download all assets shared in a channel.
- **Unified File Store**: Added `file-store.ts` for managing file uploads and listing.
- **Artifact Identity**: The `CanvasPanel` now displays the user who last edited the artifact.
- **Hydrated Metadata**: The `Artifact` model now includes full `createdByUser` and `updatedByUser` objects.

### Fixed

- **Store Types**: Updated `Artifact` interface to support rich user metadata.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.21] - 2026-04-19


This release adds file asset APIs and upgrades artifact responses so Relay can support real attachments and richer canvas identity context.

### Added

- `POST /api/v1/files/upload`
- `GET /api/v1/files`
- `GET /api/v1/files/:id`
- `GET /api/v1/files/:id/content`

### Artifact Contract

- artifact responses now include:
  - `created_by_user`
  - `updated_by_user`
- websocket `artifact.updated` payloads use the same hydrated artifact structure

### Documentation

- added [docs/phases/phase-10-file-assets.md](./docs/phases/phase-10-file-assets.md)
- added [docs/releases/v0.5.22.md](./docs/releases/v0.5.22.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for files and artifact identity hydration
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped file baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.21] - 2026-04-19

This release implements Phase 18: Artifact Lifecycle integration, bringing real dynamic canvases and AI generation to the workspace.

### Added

- **Dynamic CanvasPanel**: Replaced static demo with a full Artifact-backed surface for code and documents.
- **AI Canvas Generation**: Integrated `POST /api/v1/ai/canvas/generate` into the assistant flow, allowing AI to create new artifacts on demand.
- **Real-time Sync**: Leveraged websocket `artifact.updated` events for seamless live collaboration on canvases.
- **Unified Artifact Store**: Added `artifact-store.ts` for managing artifact CRUD and loading states.
- **Channel Header Artifacts**: Dynamic "Recent Artifacts" bar in the channel view for quick access to collaboration assets.

### Fixed

- **Engineering Excellence**: Enforced strict linting in `pnpm dev` and `pnpm build` scripts.
- **Lint Cleanup**: Resolved all remaining unused variable and import errors across the frontend.
- **Event Handling**: Stabilized websocket event mapping for unified transient state management.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint` (Verified PASS)

## [0.5.18] - 2026-04-18


This release adds backend-backed artifact lifecycle APIs so Relay's canvas surface can create, persist, update, and AI-generate collaborative outputs.

### Added

- `GET /api/v1/artifacts`
- `POST /api/v1/artifacts`
- `GET /api/v1/artifacts/:id`
- `PATCH /api/v1/artifacts/:id`
- `POST /api/v1/ai/canvas/generate`

### Realtime And Hardening

- artifact mutations now broadcast `artifact.updated`
- activity feed reaction items now use stable unique IDs, avoiding duplicate-key collisions when the same emoji appears multiple times on one message

### Documentation

- added [docs/phases/phase-10-artifact-lifecycle.md](./docs/phases/phase-10-artifact-lifecycle.md)
- added [docs/releases/v0.5.20.md](./docs/releases/v0.5.20.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for canvas and artifacts
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped artifact baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.18] - 2026-04-18

This release focuses on UX polish, branding consistency, and critical UI bug fixes including hydration errors and list rendering stability.

### Added

- **Sidebar Interactivity**: Activated "Add channels" and "Add teammates" buttons with dedicated creation/invitation dialogs.
- **Unified Branding**: Completed the full rebranding from "Acme Corp" to "Relay" across frontend, backend seed data, and documentation.
- **Starred & Pinned Page**: Created a dedicated view at `/workspace/starred` to aggregate all important collaboration assets.

### Fixed

- **Hydration Errors**: Resolved `ResizablePanel` and `DialogTitle` hydration mismatches by implementing stable mounting guards.
- **List Rendering**: Eliminated duplicate key errors in Activity and Later pages through enhanced ID mapping and deduplication logic.
- **Scroll Stability**: Fixed missing scrollbars in DM docked windows by optimizing nested flex containers and removing redundant scroll areas.
- **Date Safety**: Resolved "Invalid time value" crashes by adding defensive date formatting and reliable store-level fallbacks.
- **Syntax & Compilation**: Fixed a parsing error and duplicate `useState` import in the channel creation dialog.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.17] - 2026-04-18


This release implements Phase 17: AI Summaries integration, providing instant insights for both message threads and channels.

### Added

- **AI Thread Summaries**: Integrated persistent thread summary generation in the Thread Panel.
- **AI Channel Summaries**: Added a dedicated AI summary section in the Channel Info panel for high-level context.
- **Dynamic Summarization**: Real-time summary generation and regeneration supported via the backend AI flow.
- **Store Integration**: Expanded `MessageStore` and `ChannelStore` to handle summary state and automatic hydration.

### Fixed

- **UI Polish**: Replaced static thread summary mocks with real API-driven content and loading states.
- **Hydration**: Summaries are now automatically fetched when opening threads or switching channels.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.16] - 2026-04-18


This release implements Phase 16: AI Conversation Persistence, allowing users to browse and resume previous AI assistant chats.

### Added

- **AI History View**: Integrated a history list in the AI Assistant panel to browse previous conversations.
- **Session Continuation**: AI chats now persist across sessions and can be resumed by passing `conversation_id`.
- **New Chat**: Added a "New Chat" button to quickly reset the assistant context.
- **AI Store**: Introduced `ai-store.ts` for centralized management of AI messages and conversation history.

### Fixed

- **Store Sync**: Unified message handling between `useAIChat` and the persistent conversation list.
- **UI Interaction**: History toggle automatically fetches the latest list when opened.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.15] - 2026-04-18


This release adds persistent AI summaries for threads and channels so Relay can surface reusable context snapshots instead of relying on frontend-only mock summary cards.

### Added

- `GET /api/v1/messages/:id/summary`
- `POST /api/v1/messages/:id/summary`
- `GET /api/v1/channels/:id/summary`
- `POST /api/v1/channels/:id/summary`

### Persistence

- added `ai_summaries`
- each summary is keyed by `scope_type + scope_id`
- stored summary fields include:
  - `provider`
  - `model`
  - `content`
  - `reasoning`
  - `message_count`
  - `last_message_at`

### Documentation

- added [docs/phases/phase-10-ai-summaries.md](./docs/phases/phase-10-ai-summaries.md)
- added [docs/releases/v0.5.15.md](./docs/releases/v0.5.15.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for AI summaries
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped AI summary baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.13] - 2026-04-18


This release adds persistent AI conversation history so Relay's assistant panel can recover prior prompts, responses, and reasoning instead of acting as a purely ephemeral stream.

### Added

- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`

### Execute Behavior

- `POST /api/v1/ai/execute` now persists:
  - the user prompt
  - the assistant response
  - reasoning text
  - provider/model metadata
- a `conversation_id` can be passed to continue an existing AI conversation
- if `conversation_id` is omitted, a new one is created and returned in the stream

### Documentation

- added [docs/phases/phase-10-ai-conversation-persistence.md](./docs/phases/phase-10-ai-conversation-persistence.md)
- added [docs/releases/v0.5.13.md](./docs/releases/v0.5.13.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for AI conversation persistence
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped AI history baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.12] - 2026-04-18

This release implements persistent notification read states across the activity feed, inbox, and mentions views.

### Added

- **Unread Indicators**: Integrated blue dot indicators and distinctive styling for unread notification items.
- **Mark as Read**: Clicking an activity item now automatically marks it as read in the backend.
- **Bulk Read Action**: Added a "Mark all as read" button to clear notifications in the current view.
- **Unread Counts**: Live unread count badges on Activity, Inbox, and Mentions tabs.
- **Store Support**: Updated `activity-store.ts` with optimistic read state updates and persistent API synchronization.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.11] - 2026-04-18


This release adds persistent notification read state so Relay inbox and mentions surfaces can track which collaboration signals have already been acknowledged.

### Added

- `POST /api/v1/notifications/read`

### Feed Behavior

- `GET /api/v1/inbox` now includes `is_read`
- `GET /api/v1/mentions` now includes `is_read`

### Documentation

- added [docs/phases/phase-10-notification-read-state.md](./docs/phases/phase-10-notification-read-state.md)
- added [docs/releases/v0.5.11.md](./docs/releases/v0.5.11.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for notification read state
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped notification baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.10] - 2026-04-18

This release implements Phase 14: Stars and Pins integration, providing persistent surfaces for important channels and messages.

### Added

- **Channel Starring**: Users can now star/unstar channels directly from the message area header.
- **Pinned Messages Tab**: Integrated a new "Pins" tab in the `ChannelInfo` panel to display all pinned messages for the current channel.
- **Unpin Action**: Conveniently unpin messages directly from the `ChannelInfo` panel.
- **Store Expansion**: Added `toggleStar` to `ChannelStore` and `fetchPins` to `MessageStore` with optimistic UI updates.

### Fixed

- **Header UI**: Synchronized the header star button state with the backend `isStarred` attribute.
- **Data Integrity**: Automatically refreshes the pinned messages list when a message's pin state is toggled.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.9] - 2026-04-18


This release adds the first starred and pinned discovery APIs so Relay can surface saved channels and pinned references as dedicated collaboration destinations.

### Added

- `GET /api/v1/starred`
- `POST /api/v1/channels/:id/star`
- `GET /api/v1/pins`

### Behavior

- starred channels are returned from persisted `channels.is_starred`
- channel star toggle returns the updated channel plus `is_starred`
- pins return hydrated message items with:
  - `message`
  - `channel`
  - `user`

### Documentation

- added [docs/phases/phase-10-stars-pins.md](./docs/phases/phase-10-stars-pins.md)
- added [docs/releases/v0.5.9.md](./docs/releases/v0.5.9.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for starred and pinned surfaces
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped discovery baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.8] - 2026-04-18

This release implements Phase 13: Presence and Typing indicators, bringing real-time social signals to the workspace.

### Added

- **Live Presence**: Real-time user status (Online, Away, Busy, Offline) across the entire UI.
- **Typing Indicators**: Visual feedback when teammates are typing in Channels, DMs, or Threads.
- **Presence Store**: New `presence-store.ts` for managing transient real-time states.
- **Websocket Expansion**: Integrated `presence.updated` and `typing.updated` events into the unified websocket hook.

### Fixed

- UI consistency: Status dots now appear correctly on all user avatars and navigation elements.
- Typing debounce: Efficiently manages typing broadcast state to minimize network overhead.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.7] - 2026-04-18


This release adds the first realtime presence layer for Relay by shipping presence APIs and typing broadcasts for channels, DMs, and threads.

### Added

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`

### Realtime

- `POST /api/v1/presence` broadcasts `presence.updated`
- `POST /api/v1/typing` broadcasts `typing.updated`

### Documentation

- added [docs/phases/phase-10-presence-typing.md](./docs/phases/phase-10-presence-typing.md)
- added [docs/releases/v0.5.7.md](./docs/releases/v0.5.7.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for presence and typing
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped realtime baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.6] - 2026-04-18

This release implements the Phase 12 Drafts integration, bringing autosave persistence to all composer surfaces.

### Added

- **Drafts Persistence**: Integrated `GET /api/v1/drafts` and `PUT /api/v1/drafts/:scope` for reliable message autosaving.
- **Rich Thread Replies**: Replaced the basic thread textarea with the rich `MessageComposer`, enabling bold, italic, and slash commands in threads.
- **Unified Draft Store**: Added `draft-store.ts` to manage draft state across Channels, DMs, and Threads.
- **Autosave & Restore**: Content is automatically saved as you type and restored when you return to a conversation.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.5] - 2026-04-18


This release starts the next Slack-parity backend wave by adding draft persistence APIs for channel, DM, and thread composers.

### Added

- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`

### Behavior

- drafts are returned for the current user only
- drafts are ordered by `updated_at desc`
- one draft is stored per `user_id + scope`
- recommended scope keys:
  - `channel:<channelId>`
  - `dm:<dmId>`
  - `thread:<messageId>`

### Persistence

- added `drafts`
- seeded one example draft for local development visibility

### Documentation

- added [docs/phases/phase-10-drafts.md](./docs/phases/phase-10-drafts.md)
- added [docs/releases/v0.5.5.md](./docs/releases/v0.5.5.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for drafts
- updated `README.md` and `docs/phase8-api-expansion.md` to reflect the shipped drafts baseline

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.4] - 2026-04-18

This release implements the Phase 11 notification UI wave, introducing a unified Activity, Inbox, and Mentions experience.

### Added

- **Inbox & Mentions UI**: Integrated new notification tabs in the Activity page.
- **Unified Activity Store**: Expanded `activity-store.ts` to support fetching Inbox and Mentions from backend endpoints.
- **Smart Navigation**: Activity items now support clicking to jump directly to the relevant channel or DM conversation.
- **Interactive Tabs**: Seamless switching between "All Activity", "Inbox", and "Mentions" views.

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.3] - 2026-04-18


This release unifies Gemini's `v0.5.2` channel-management frontend work with the next Codex backend wave by adding inbox and mentions APIs.

### Added

- `GET /api/v1/inbox`
- `GET /api/v1/mentions`

### Behavior

- inbox returns aggregated collaboration signals for the current user:
  - direct mentions
  - thread replies
  - reactions on your messages
  - DM activity
- mentions returns the direct-mention subset only

### Documentation

- added [docs/phases/phase-10-inbox-mentions.md](./docs/phases/phase-10-inbox-mentions.md)
- added [docs/releases/v0.5.3.md](./docs/releases/v0.5.3.md)
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for inbox and mentions
- updated `docs/phase8-api-expansion.md` shipped baseline to include the new endpoints

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.2] - 2026-04-18

This release completes the Phase 10 frontend integration for channel management, introducing member lists, metadata editing, and a unified channel info panel.

### Added

- `ChannelInfo` panel for managing channel metadata and membership
- support for editing channel `topic` and `purpose`
- member management UI (list, add, and remove members)
- extended `Channel` and `ChannelMember` types for robust metadata handling
- integrated `ChannelInfo` into the main message area header

### Fixed

- channel metadata sync: ensures `topic` and `purpose` are correctly updated across the UI
- member list hydration: automatically fetches and updates members when switching channels

### Verification Used For This Release

- `pnpm build`
- `cd apps/web && pnpm lint`

## [0.5.1] - 2026-04-18

This release starts Phase 10 with the first Slack-parity backend wave and adds explicit stage documentation for ongoing Codex + Gemini collaboration.

### Added

- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`

### Data Model

- added `channel_members`
- added `workspace_invites`
- extended `channels` with:
  - `topic`
  - `purpose`
  - `is_archived`

### Documentation

- added [docs/phases/phase-10-slack-parity-foundation.md](./docs/phases/phase-10-slack-parity-foundation.md)
- added [docs/releases/v0.5.1.md](./docs/releases/v0.5.1.md)
- updated `docs/AGENT-COLLAB.md` with Gemini handoff for members, invites, and metadata editing
- updated `docs/phase8-api-expansion.md` shipped baseline to include the new endpoints

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.5.0] - 2026-04-18

This release synchronizes Gemini's docked DM overhaul with the backend contract, confirms the DM payload shape in handler tests, and documents the next Slack-parity API wave for Relay.

### Added

- DM conversation responses now formally include `user_ids`:
  - `GET /api/v1/dms`
  - `POST /api/v1/dms`
- handler coverage now verifies:
  - DM list payloads include `user_ids`
  - DM create/open payloads include `user_ids`

### Frontend Integration Sync

- `/workspace/dms` now redirects back into the unified workspace shell
- docked DM chat windows are mounted from the workspace layout
- DM store now maps `user_ids` and defensive `last_message_at` fallbacks
- message-store duplicate guards reduce repeated DM and thread inserts
- IME-safe Enter handling was added to the rich message composer

### Planning And Documentation

- updated `docs/phase8-api-expansion.md` with the next Slack-parity API layer:
  - channel members
  - invites
  - channel topic and purpose
  - stars and pins
  - inbox and mentions
  - drafts
- updated `docs/AGENT-COLLAB.md` with the `v0.5.0` handoff and next-phase objectives
- updated `README.md` to reflect the docked DM experience and the new planning track

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.4.1] - 2026-04-18

This release closes the DM realtime gap and expands the post-Phase-8 backend with Activity, Later, and Search APIs so Gemini can finish the remaining workspace pages against live data.

### Added

- `GET /api/v1/activity`
  - returns recent collaboration signals for the current user
  - includes mentions, thread replies, reactions on your messages, and DM activity
- `GET /api/v1/later`
  - returns the current user's saved messages with channel and sender context
- `GET /api/v1/search?q=...`
  - returns grouped results across:
    - `channels`
    - `users`
    - `messages`
    - `dms`

### Realtime And DM Improvements

- `POST /api/v1/dms/:id/messages` now broadcasts `message.created`
- DM websocket payloads now include:
  - `id`
  - `dm_id`
  - `user_id`
  - `content`
  - `created_at`
- `POST /api/v1/dms` now accepts either:
  - `user_id`
  - `user_ids`
  which keeps the backend compatible with the current frontend store shape

### Documentation Sync

- updated `README.md` current status to the `v0.4.1` release line
- updated `docs/phase8-api-expansion.md` so DM, Activity, Later, and Search are no longer listed as missing
- updated `docs/AGENT-COLLAB.md` with the Gemini handoff for page integration

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.9] - 2026-04-18

This release starts the first post-Phase-8 backend expansion by adding the minimum DM API surface needed for Gemini to replace the static DMs page.

### Added

- `GET /api/v1/dms`
  - returns recent DM conversations for the current user
  - includes the counterpart user plus last message preview
- `POST /api/v1/dms`
  - creates or reopens a 1:1 DM conversation
  - payload: `{ "user_id": string }`
- `GET /api/v1/dms/:id/messages`
  - returns ordered DM history
- `POST /api/v1/dms/:id/messages`
  - sends a DM message
  - payload:
    - `content`
    - `user_id`

### Data Model

- added:
  - `dm_conversations`
  - `dm_members`
  - `dm_messages`
- seeded one initial DM thread between `user-1` and `user-2`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.7] - 2026-04-18

This release adds a real backend collaboration insight path for user profiles and fixes `#agent-collab` so it renders meaningful content on first load.

### Added

- `GET /api/v1/agent-collab/snapshot`
  - returns:
    - `active_superpowers`
    - `task_board`
- dynamic `ai_insight` generation on existing user payloads:
  - `GET /api/v1/me`
  - `GET /api/v1/users`

### Fixed

- `#agent-collab` no longer depends on catching a startup websocket event
- frontend collab store now correctly maps backend lowercase JSON keys
- the dashboard now renders both:
  - agent cards
  - task board rows

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- `pnpm build`

## [0.3.4] - 2026-04-18

This release closes the audit loop on the original LLM/thread/user API delivery plan and synchronizes repository-facing docs with the shipped backend state.

### Audit Result

- `docs/superpowers/plans/2026-04-17-llm-thread-user-api.md` has been audited against the live codebase
- conclusion: the original plan scope is complete
- the plan file now reflects shipped status instead of stale unchecked tasks

### Documentation Sync

- updated `README.md` to point at the latest release line
- updated `docs/AGENT-COLLAB.md` with the audit result and Gemini handoff note

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.3.3] - 2026-04-18

This release aligns the backend with Gemini's latest realtime AI integration pass, adds reasoning-aware SSE support, and refreshes the repository docs for GitHub-facing consumption.

### Added

- `POST /api/v1/ai/execute` may now emit `event: reasoning` in addition to:
  - `start`
  - `chunk`
  - `done`
  - `error`

### Realtime Coverage

- reaction mutations now broadcast `reaction.updated`
- pin toggles now broadcast `message.updated`
- deletions now broadcast `message.deleted`

### Documentation Refresh

- updated `README.md` with current product positioning and shipped backend surface
- replaced broken local-only links with GitHub-safe relative links
- rewrote `docs/phase8-api-expansion.md` to reflect the current monorepo and broader backend target
- updated `docs/AGENT-COLLAB.md` with Gemini handoff notes for reasoning and websocket coverage

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.3.1] - 2026-04-18

This release completes the first persistent message interaction APIs for Gemini's channel message actions UI.

### Added

- `POST /api/v1/ai/feedback`
  - Payload: `{ "message_id": string, "is_good": boolean }`
  - Response: `{ "feedback": { ... } }`
- `POST /api/v1/messages/:id/reactions`
  - Payload: `{ "emoji": string }`
  - Toggle behavior for the current user
  - Response:
    - `message`
    - `added`
- `DELETE /api/v1/messages/:id`
  - Deletes the target message
  - Deletes child replies when the target is a thread parent
  - Response: `{ "deleted": true, "message_id": string }`
- `POST /api/v1/messages/:id/pin`
  - Toggle message pin state
  - Response:
    - `message`
    - `is_pinned`
- `POST /api/v1/messages/:id/later`
  - Toggle save-for-later state for the current user
  - Response:
    - `message_id`
    - `saved`
- `POST /api/v1/messages/:id/unread`
  - Marks the message as an unread checkpoint for the current user
  - Response:
    - `message_id`
    - `unread`

### Data Model And Behavior

- `Message` now includes:
  - `is_pinned`
- New persistence tables:
  - `message_reactions`
  - `saved_messages`
  - `unread_markers`
  - `ai_feedback`
- `GET /api/v1/messages` and `GET /api/v1/messages/:id/thread` now rebuild `metadata.reactions` from persisted reaction rows
- deleting a thread reply now recomputes the parent message:
  - `reply_count`
  - `last_reply_at`

### Verification Used For This Release

- `cd apps/api && go test ./...`
- `cd apps/api && go build ./...`
- `pnpm build`

## [0.2.8] - 2026-04-17

This release removes the last hardcoded AI settings path for the frontend and adds persistence for user AI preferences plus stronger thread reply metadata.

### Added

- `GET /api/v1/ai/config`
  - Returns enabled providers and configured models
  - Response shape:
    - `default_provider`
    - `providers[]`
- `PATCH /api/v1/me/settings`
  - Persists AI preference fields on the current user:
    - `provider`
    - `model`
    - `mode`

### Thread Integrity Improvements

- reply creation now updates:
  - `reply_count`
  - `last_reply_at`
- `Message` model now stores:
  - `last_reply_at`
- `User` model now stores:
  - `ai_provider`
  - `ai_model`
  - `ai_mode`

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.2.5] - 2026-04-17

This release hardens local LLM configuration loading and records the first real upstream validation pass against the configured providers.

### Fixed

- layered LLM config merge now preserves provider defaults while applying local overrides
- local YAML parsing now tolerates tab characters in `llm.local.yaml`

### Verified

- local `gemini` provider configuration successfully streamed real SSE output through `POST /api/v1/ai/execute`
- local `openrouter` configuration loaded correctly, but the selected free model returned upstream `429` rate limiting during validation

### Notes

- with the current local config, Gemini is the verified working provider for frontend integration
- OpenRouter can still be used after switching to a non-rate-limited model or retrying later

## [0.2.4] - 2026-04-17

This release adds the first real LLM gateway architecture plus the backend APIs Gemini requested for user resolution, message threading, and AI streaming.

### Added

- Provider-based LLM gateway in `apps/api/internal/llm`
- Config loading in `apps/api/internal/config`
- Config files:
  - `apps/api/config/llm.base.yaml`
  - `apps/api/config/llm.example.yaml`
  - `apps/api/config/llm.local.yaml`
  - `apps/api/config/llm.secrets.local.yaml`
- Supported provider kinds:
  - `openai`
  - `openai-compatible`
  - `openrouter`
  - `gemini`

### API Surface Added In This Release

- `GET /api/v1/users`
  - Returns all users
  - Supports optional query param `id`
  - Response: `{ "users": [...] }`

- `GET /api/v1/messages/:id/thread`
  - Returns parent message plus replies
  - Response:
    - `parent`
    - `replies`

- `POST /api/v1/ai/execute`
  - Accepts:
    - `prompt`
    - `channel_id`
    - optional `provider`
    - optional `model`
  - Returns `text/event-stream`
  - Stream events currently emitted:
    - `start`
    - `chunk`
    - `done`
    - `error`

### Model And Handler Changes

- `Message` now includes:
  - `thread_id`
  - `reply_count`
- `GET /api/v1/messages` now returns top-level channel messages only
- `POST /api/v1/messages` accepts optional `thread_id`
- Reply creation increments parent `reply_count`

### LLM Notes

- OpenAI and OpenAI-compatible providers use configurable `api_style`
  - `responses`
  - `chat_completions`
- OpenRouter defaults to `responses`
- Gemini uses the official native Gemini streaming protocol over `streamGenerateContent?alt=sse`
- Env overrides are supported through:
  - `LLM_DEFAULT_PROVIDER`
  - `LLM_PROVIDER_<NAME>_API_KEY`
  - `LLM_PROVIDER_<NAME>_BASE_URL`
  - `LLM_PROVIDER_<NAME>_MODEL`
  - `LLM_PROVIDER_<NAME>_API_STYLE`
  - `LLM_PROVIDER_<NAME>_ENABLED`

### Verification Used For This Release

- `pnpm build`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`

## [0.2.2] - 2026-04-16

This release adds the backend sync path for the `#agent-collab` workspace view and aligns the repository version with the latest cross-agent handoff.

### Added

- `docs/AGENT-COLLAB.md` file watcher in `apps/api`
- Markdown table parser for:
  - `Task Board`
  - `Active Superpowers`
- Realtime broadcast for `agent_collab.sync`

### Realtime Event Added

- `agent_collab.sync`
  - Broadcast through `GET /api/v1/realtime`
  - Fixed target channel id: `ch-collab`
  - Payload shape:
    - `active_superpowers`
    - `task_board`

Example payload:

```json
{
  "type": "agent_collab.sync",
  "channel_id": "ch-collab",
  "payload": {
    "active_superpowers": [],
    "task_board": []
  }
}
```

### Backend Notes

- Watch path defaults to `../../docs/AGENT-COLLAB.md` from `apps/api`
- Sync is pushed on service startup and on subsequent file `write/create` events
- Parsing currently targets the two collaboration tables only

### Verification Used For This Release

- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- Automated coverage added for:
  - Markdown table parsing
  - Direct sync broadcast
  - Watcher-triggered broadcast after file write

## [0.2.0] - 2026-04-16

This release turns Relay Agent Workspace into a monorepo with a working backend foundation. It is the first version suitable for frontend-to-backend integration across `apps/web` and `apps/api`.

### Added

- Monorepo structure with `apps/web` and `apps/api`
- Go API service built with Gin
- SQLite persistence using GORM auto-migrations
- Seed data for:
  - `Organization`
  - `Team`
  - `User`
  - `Agent`
  - `Workspace`
  - `Channel`
  - `Message`
- In-memory realtime WebSocket hub
- REST handlers for collaboration and organization workflows

### API Surface In This Release

Base URL:
- Web: `http://localhost:3000`
- API: `http://localhost:8080`

Health:
- `GET /ping`
  - Returns `{"message":"pong"}`

Current user:
- `GET /api/v1/me`
  - Returns the seeded current user
  - Response shape:
    - `user.id`
    - `user.org_id`
    - `user.name`
    - `user.email`
    - `user.avatar`
    - `user.status`

Organizations:
- `GET /api/v1/orgs`
  - Lists organizations visible to the current seeded user
  - Response: `{ "organizations": [...] }`

Teams:
- `GET /api/v1/orgs/:id/teams`
  - Lists teams under an organization
  - Example: `/api/v1/orgs/org_1/teams`
  - Response: `{ "teams": [...] }`

Agents:
- `POST /api/v1/orgs/:id/agents`
  - Creates an agent under an organization
  - Required JSON body:
    - `name`
    - `type`
    - `owner_id`
  - Response: `{ "agent": { ... } }`

Workspaces:
- `GET /api/v1/workspaces`
  - Lists workspaces from SQLite
  - Response: `{ "workspaces": [...] }`

Channels:
- `GET /api/v1/channels`
  - Supports query param:
    - `workspace_id`
  - Example: `/api/v1/channels?workspace_id=ws_1`
  - Response: `{ "channels": [...] }`

Messages:
- `GET /api/v1/messages`
  - Required query param:
    - `channel_id`
  - Example: `/api/v1/messages?channel_id=ch_1`
  - Returns channel messages ordered by `created_at asc`
  - Response: `{ "messages": [...] }`

- `POST /api/v1/messages`
  - Creates a new message and persists it to SQLite
  - Required JSON body:
    - `channel_id`
    - `content`
    - `user_id`
  - Response: `{ "message": { ... } }`
  - Side effect:
    - broadcasts a realtime `message.created` event to websocket clients

Realtime:
- `GET /api/v1/realtime`
  - WebSocket upgrade endpoint
  - Current shipped events:
    - `realtime.connected`
    - `message.created`
  - `message.created` event envelope includes:
    - `id`
    - `type`
    - `workspace_id`
    - `channel_id`
    - `entity_id`
    - `ts`
    - `payload`

### Notes For Frontend Integration

- Persistence uses `apps/api/db/relay.db`
- Database auto-migration runs on API startup
- Seed data is idempotent and backfills newly added records
- Realtime is currently in-memory and single-instance only
- Message pagination, presence, typing, AI execute, and SSE are not part of `0.2.0`

### Verification Used For This Release

- `pnpm build`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...`
- `cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...`
- Manual local verification of:
  - `GET /ping`
  - `GET /api/v1/orgs`
  - `GET /api/v1/orgs/org_1/teams`
  - `POST /api/v1/orgs/org_1/agents`
  - `GET /api/v1/workspaces`
  - `GET /api/v1/channels?workspace_id=ws_1`
  - `GET /api/v1/messages?channel_id=ch_1`
  - `POST /api/v1/messages`
  - `GET /api/v1/realtime` websocket upgrade and event receipt

### Internal Commits Included

- `e8e3bba` refactor: migrate to monorepo structure with Go backend boilerplate (Phase 8.1)
- `f628d06` feat: implement phase 8 core collaboration api
- `32202a8` chore: move api sqlite database into db directory
- `c752f9d` feat: add organization team and agent api endpoints
- `a1e8699` feat: add realtime websocket hub
