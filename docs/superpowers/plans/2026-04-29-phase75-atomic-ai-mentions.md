# Phase 75 Atomic Mentions And Channel AI Replies Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make channel `@user` mentions atomic and let `@AI Assistant` trigger a channel-native AI reply with DM-style thinking, tools, and token usage.

**Architecture:** Gemini adds durable user classification, structured mention parsing, and channel AI reply orchestration in the Go API. Windsurf upgrades the Tiptap composer to emit atomic mention nodes and consumes channel AI streaming/metadata in the Web. Codex owns integration sequencing and release audit.

**Tech Stack:** Go/Gin, GORM/SQLite, existing realtime hub, existing LLM gateway, Next.js 16, React 19, TypeScript, Tiptap/ProseMirror, Zustand, existing `metadata.ai_sidecar` renderer.

---

## Parallel Work Boundaries

- Gemini write scope: `apps/api/**`, backend tests, seed/model/migration-adjacent code only.
- Windsurf write scope: `apps/web/**`, Web tests/checks only.
- Codex write scope: `docs/**`, `CHANGELOG.md`, final integration review only.
- Do not let Gemini edit Tiptap UI files.
- Do not let Windsurf invent backend fields outside this contract.

## Task 0: Codex Contract Freeze

**Owner:** Codex
**Files:**
- Create: `docs/superpowers/specs/2026-04-29-phase75-atomic-ai-mentions-design.md`
- Create: `docs/superpowers/plans/2026-04-29-phase75-atomic-ai-mentions.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `CHANGELOG.md`
- Create: `docs/releases/v0.6.75.md`

- [ ] **Step 1: Publish the frozen split**

Record Gemini/Windsurf ownership, acceptance criteria, and the realtime contract.

- [ ] **Step 2: Verify the current baseline**

Run:

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase73BusinessApps|TestPhase74|TestCreateMessage|TestCreateDMMessage|TestPhase65CAISlashCommandAsk' -count=1
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- PASS before Phase 75 implementation starts.

- [ ] **Step 3: Commit and tag planning release**

```bash
git add docs/superpowers/specs/2026-04-29-phase75-atomic-ai-mentions-design.md docs/superpowers/plans/2026-04-29-phase75-atomic-ai-mentions.md docs/AGENT-COLLAB.md CHANGELOG.md docs/releases/v0.6.75.md
git commit -m "docs: plan phase 75 atomic ai mentions"
git tag v0.6.75
git push origin main
git push origin v0.6.75
```

## Task 1: Backend User Type Contract

**Owner:** Gemini
**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/seed.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing tests**

Cover:
- seeded `user-2` serializes with `user_type=ai`
- normal users default to `human` when empty
- `isAIUser` returns true for explicit `user_type=ai`
- `isAIUser` returns true for explicit `user_type=bot`
- legacy heuristic still works when `user_type` is empty
- explicit `user_type=human` does not trigger AI through heuristic unless empty fallback is needed

Run:

```bash
cd apps/api && go test ./internal/handlers -run TestPhase75AIUserType -count=1
```

Expected:
- FAIL before implementation.

- [ ] **Step 2: Implement user type**

Requirements:
- add `UserType string json:"user_type"` to `domain.User`
- seed `AI Assistant` as `ai`
- keep existing rows compatible by treating empty type as heuristic fallback
- expose `user_type` through existing user APIs without separate endpoint changes

- [ ] **Step 3: Re-run tests**

Expected:
- PASS.

- [ ] **Step 4: Commit**

```bash
git add apps/api/internal/domain/models.go apps/api/internal/db/seed.go apps/api/internal/handlers/collaboration.go apps/api/internal/handlers/collaboration_test.go
git commit -m "feat(api): add ai user type contract"
```

## Task 2: Backend Structured Mention Parsing

**Owner:** Gemini
**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing tests**

Cover:
- `CreateMessage` content containing `data-mention-user-id="user-2"` persists a `MessageMention`
- structured mention wins even if visible text varies
- duplicate structured + text mention persists one row
- text-only `@AI Assistant` fallback still works
- `message.metadata.user_mentions[]` contains `user_id`, `name`, and `mention_text`

Run:

```bash
cd apps/api && go test ./internal/handlers -run TestPhase75StructuredMentions -count=1
```

Expected:
- FAIL before implementation.

- [ ] **Step 2: Implement parser**

Requirements:
- parse known `data-mention-*` attributes from message HTML
- validate mentioned user belongs to the channel candidate list
- dedupe by `message_id + user_id + mention_kind`
- preserve existing `findExplicitUserMentions` text fallback
- avoid broad HTML parsing dependencies unless needed

- [ ] **Step 3: Re-run tests**

Expected:
- PASS.

- [ ] **Step 4: Commit**

```bash
git add apps/api/internal/handlers/collaboration.go apps/api/internal/handlers/collaboration_test.go
git commit -m "feat(api): parse structured user mentions"
```

## Task 3: Backend Channel AI Mention Replies

**Owner:** Gemini
**Files:**
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/ai.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`
- Test: `apps/api/internal/handlers/ai_test.go`

- [ ] **Step 1: Write failing tests**

Cover:
- channel message mentioning AI user triggers one AI reply
- channel message mentioning human user does not trigger AI
- AI user's own message does not trigger recursive AI
- final AI reply is saved as `domain.Message` with `UserID` equal to the mentioned AI user
- final AI reply metadata includes `ai_sidecar`
- final AI reply metadata includes `ai_mention_reply.trigger_message_id`
- `/ask` remains compatible

Run:

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase75ChannelAIMentionReply|TestPhase65CAISlashCommandAsk' -count=1
```

Expected:
- FAIL before implementation.

- [ ] **Step 2: Extract reusable AI reply helpers**

Requirements:
- keep `triggerAIDMReply` behavior intact
- share prompt construction or gateway streaming helpers where practical
- add channel context prompt using recent channel messages
- use the mentioned AI user as the sender
- no-op safely when `AIGateway == nil`

- [ ] **Step 3: Add channel stream events**

Requirements:
- broadcast `channel.ai.stream.chunk` or a documented generalized equivalent
- payload includes `temp_id`, `channel_id`, `trigger_message_id`, `ai_user_id`, `kind`, `chunk`, `is_final`
- final persisted `message.created` must include metadata

- [ ] **Step 4: Re-run tests**

Expected:
- PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/handlers/collaboration.go apps/api/internal/handlers/ai.go apps/api/internal/handlers/collaboration_test.go apps/api/internal/handlers/ai_test.go
git commit -m "feat(api): trigger channel ai replies from mentions"
```

## Task 4: Web User Types And Mention Target Shape

**Owner:** Windsurf
**Files:**
- Modify: `apps/web/types/index.ts`
- Modify: `apps/web/stores/user-store.ts`
- Modify: `apps/web/components/message/mention-popover.tsx`

- [ ] **Step 1: Write failing type checks**

Cover:
- API payload field is `user_type`
- Web canonical field is `userType`
- `user-store` maps `user_type` to `userType` exactly once at the API boundary
- AI users are identified by `userType=ai|bot` first
- legacy fallback still handles `user-2`, assistant name, or `ai@` email
- `MentionPopover` returns structured target objects, not a string

Run:

```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- FAIL until types and props are updated.

- [ ] **Step 2: Implement target shape**

Requirements:
- define `ComposerMentionTarget`
- update `MentionPopover` prop to `onSelect(target)`
- stop hardcoding AI-only behavior to `user.id === "user-2"` except as fallback styling
- keep `userType` as the only canonical Web field after store normalization
- keep user group rows functional and non-AI-triggering

- [ ] **Step 3: Re-run type check**

Expected:
- PASS.

- [ ] **Step 4: Commit**

```bash
git add apps/web/types/index.ts apps/web/stores/user-store.ts apps/web/components/message/mention-popover.tsx
git commit -m "feat(web): type composer mention targets"
```

## Task 5: Web Tiptap Atomic Mention Node

**Owner:** Windsurf
**Files:**
- Create: `apps/web/components/message/tiptap-user-mention.ts`
- Modify: `apps/web/components/message/message-composer.tsx`
- Modify: `apps/web/components/message/message-composer-demo.tsx`

- [ ] **Step 1: Write failing behavior checks**

If the repo has no editor unit harness, create a focused component-level test or document a manual verification script. Required checks:
- selecting `@AI Assistant` inserts one atom node
- serialized HTML contains all required attributes:
  - `data-mention-kind="user"`
  - `data-mention-user-id`
  - `data-mention-name`
  - `data-mention-user-type`
- the rendered chip is visually distinct from plain text
- copy/paste degrades to readable `@AI Assistant` text when atom metadata cannot be preserved
- caret after the chip + Backspace removes the whole chip
- entity mentions still insert through existing `@entity:` path
- slash commands still work

Run:

```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- FAIL until the extension is wired.

- [ ] **Step 2: Implement Tiptap extension**

Requirements:
- inline atom node
- selectable
- `contenteditable=false` render output
- attrs: `kind`, `userId`, `name`, `userType`, `mentionText`
- parse/render `data-mention-*` attributes
- render all four backend-required data attributes: `data-mention-kind`, `data-mention-user-id`, `data-mention-name`, `data-mention-user-type`
- include a distinct chip class/style hook so the mention is not indistinguishable from normal text
- preserve readable `@Name` text as the node's fallback text content
- command helper for insertion after deleting the trailing `@` trigger

- [ ] **Step 3: Wire composer insertion**

Requirements:
- replace trailing `@` trigger with the atom node plus trailing space
- do not insert only `name + " "`
- keep autosave/draft restore compatible with atom HTML
- keep upload/file attachment send path unchanged

- [ ] **Step 4: Re-run checks**

Expected:
- PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/web/components/message/tiptap-user-mention.ts apps/web/components/message/message-composer.tsx apps/web/components/message/message-composer-demo.tsx
git commit -m "feat(web): add atomic user mention chips"
```

## Task 6: Web Channel AI Streaming Consumption

**Owner:** Windsurf
**Files:**
- Modify: `apps/web/hooks/use-websocket.ts`
- Modify: `apps/web/stores/message-store.ts`
- Modify: `apps/web/components/message/message-item.tsx`
- Optional Create: `apps/web/components/message/channel-ai-streaming-message.tsx`

- [ ] **Step 1: Write failing checks**

Cover:
- realtime `message.created` keeps metadata when mapping to store
- `channel.ai.stream.chunk` creates/updates a temporary channel AI row
- final stream event removes the temporary row
- persisted AI message renders `ReasoningPanel`, `ToolTimeline`, and `UsageChip`

Run:

```bash
cd apps/web && pnpm exec tsc --noEmit
```

Expected:
- FAIL until store/event handling is updated.

- [ ] **Step 2: Preserve metadata in websocket messages**

Requirements:
- use the same mapping behavior as REST messages where possible
- do not drop `metadata`, `attachments`, or `reactions` from realtime `message.created`
- keep duplicate prevention by message id

- [ ] **Step 3: Add channel AI stream state**

Requirements:
- state keyed by `temp_id`
- include `channel_id`, `ai_user_id`, `trigger_message_id`, accumulated answer text, reasoning/tool/usage partials if available
- display a temporary AI thinking row in the active channel
- remove temp row when final persisted message arrives

- [ ] **Step 4: Re-run checks**

Expected:
- PASS.

- [ ] **Step 5: Commit**

```bash
git add apps/web/hooks/use-websocket.ts apps/web/stores/message-store.ts apps/web/components/message/message-item.tsx
git commit -m "feat(web): render channel ai mention streams"
```

## Task 7: Integration Verification And Release

**Owner:** Codex
**Files:**
- Modify: `CHANGELOG.md`
- Create: `docs/releases/v0.6.76.md` if implementation ships after this planning release
- Modify: `docs/AGENT-COLLAB.md`

- [ ] **Step 1: Fetch latest worker commits**

```bash
git fetch origin main --no-tags
git merge origin/main
```

- [ ] **Step 2: Run full focused checks**

```bash
cd apps/api && go test ./internal/handlers -run 'TestPhase75|TestCreateMessage|TestCreateDMMessage|TestPhase65CAISlashCommandAsk' -count=1
cd apps/web && pnpm exec tsc --noEmit
cd apps/web && pnpm lint
```

Expected:
- PASS.

- [ ] **Step 3: Manual acceptance**

Verify in the running app:
- `@AI Assistant` inserts as one chip
- Backspace deletes the full chip
- sent channel message persists `user_mentions`
- channel AI reply appears from AI Assistant
- reasoning/tools/token usage renders in the channel message list
- DM AI replies still work
- `/ask` still works

- [ ] **Step 4: Publish implementation release**

Only after Gemini and Windsurf implementation commits are merged:

```bash
git tag v0.6.76
git push origin main
git push origin v0.6.76
```
