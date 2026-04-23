# Phase 65A User Mention Semantics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `@user` mentions a durable backend primitive by persisting them on message creation and using the same source for unified feed, mentions, and the mention branch of inbox.

**Architecture:** Keep message creation as the source of truth, enrich messages with `user_mentions[]` render data, and add a lightweight `MessageMention` query index for mention-centric APIs. Reuse existing realtime and feed contracts rather than introducing a second notification system.

**Tech Stack:** Go, Gin, GORM, SQLite, existing message/DM handlers, existing websocket hub, existing unified activity feed handler.

**Spec:** `docs/superpowers/specs/2026-04-24-phase65a-user-mention-semantics-design.md`

---

### Task 1: Phase 65A Contract Tests

**Files:**
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `TestPhase65AChannelMessagePersistsUserMentions`.
- [ ] Add `TestPhase65ADMMessagePersistsUserMentions`.
- [ ] Add `TestPhase65AGetMentionsUsesMessageMention`.
- [ ] Add `TestPhase65AUnifiedFeedReturnsUserMentionRows`.
- [ ] Add `TestPhase65AInboxMentionBranchUsesMessageMention`.
- [ ] Add `TestPhase65AMentionCreatedBroadcast`.
- [ ] Run `go test ./internal/handlers -run TestPhase65A -count=1`.
- [ ] Confirm RED before implementation.

### Task 2: Mention Persistence Model And Parsing Helpers

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `MessageMention` model with indexes needed for user, workspace, channel, and DM mention queries.
- [ ] Extend message metadata shape with `user_mentions[]`.
- [ ] Add helper logic to resolve explicit `@user` mentions against workspace users.
- [ ] Enforce longest-name-first and per-message dedupe.
- [ ] Keep parsing deterministic; do not reintroduce fuzzy display-name scanning.

### Task 3: Channel Message Mention Persistence

**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] After channel message creation, resolve workspace scope and parse user mentions.
- [ ] Persist `message.metadata.user_mentions[]` on the created message.
- [ ] Insert `MessageMention` rows for valid targets.
- [ ] Allow self-mentions in persistence but exclude them from later query surfaces.
- [ ] Ensure message creation still succeeds if mention enrichment fails.
- [ ] Run `go test ./internal/handlers -run TestPhase65AChannelMessagePersistsUserMentions -count=1`.

### Task 4: DM Message Mention Persistence

**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Resolve the DM workspace scope from the conversation participants or related user/org context.
- [ ] Parse explicit `@user` mentions for DM messages.
- [ ] Persist `MessageMention` rows with `dm_id`.
- [ ] If DM message model cannot support message metadata cleanly, keep index persistence as the required bar and document any metadata limitation.
- [ ] Run `go test ./internal/handlers -run TestPhase65ADMMessagePersistsUserMentions -count=1`.

### Task 5: Mentions API And Unified Feed Migration

**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Rewrite `GET /api/v1/mentions` to query `MessageMention`.
- [ ] Exclude self-mentions from the current user's mention feed.
- [ ] Extend unified activity feed `mention` rows to source from `MessageMention`.
- [ ] Preserve entity mention rows with `mention_kind=entity`.
- [ ] Add `mention_kind=user` and `mentioned_user_id` to feed `meta` for user mentions.
- [ ] Run `go test ./internal/handlers -run 'TestPhase65A(GetMentionsUsesMessageMention|UnifiedFeedReturnsUserMentionRows)' -count=1`.

### Task 6: Inbox Mention Alignment And Realtime

**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Replace the mention branch inside `GET /api/v1/inbox` with the same `MessageMention` source.
- [ ] Add websocket `mention.created`.
- [ ] Reuse unified feed-like payload shape where practical so Windsurf can append with minimal transform.
- [ ] Run `go test ./internal/handlers -run 'TestPhase65A(InboxMentionBranchUsesMessageMention|MentionCreatedBroadcast)' -count=1`.

### Task 7: Documentation And Collaboration Sync

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `docs/phase64-slack-core-convergence.md`
- Create: `docs/releases/v0.6.xx.md`
- Modify: `package.json`
- Modify: `apps/web/package.json`

- [ ] Document `MessageMention`, persisted `@user` mention semantics, and the new realtime event.
- [ ] Leave Windsurf a concrete handoff for `mention_kind=user|entity`, `mentioned_user_id`, and `mention.created`.
- [ ] Record how Phase 65A advances the remembered next-step `2` by strengthening the unified feed itself.

### Task 8: Verification And Release

**Files:**
- No additional files required unless verification exposes gaps.

- [ ] Run `go test ./internal/handlers -count=1`.
- [ ] Run `go test ./...`.
- [ ] Run `GOCACHE=$(pwd)/.cache/go-build go build ./...`.
- [ ] Run `pnpm --filter relay-agent-workspace lint`.
- [ ] Run `pnpm --filter relay-agent-workspace exec tsc --noEmit`.
- [ ] Run `git diff --check`.
- [ ] Commit the phase with a release-oriented message.
- [ ] Tag and push the new version.

---

## Execution Notes

- Implement in three internal cuts even if the release ships as one version:
  - `65A-A`: mention persistence on message create
  - `65A-B`: mentions endpoint + unified feed migration
  - `65A-C`: inbox mention alignment + realtime + docs
- Do not redesign the whole notification system in this phase.
- Keep parsing deterministic and workspace-scoped. Avoid fuzzy matching or display-name substring behavior.
- If DM metadata persistence becomes awkward, prefer preserving the `MessageMention` query index and documenting the limitation rather than over-refactoring DM storage in this slice.
