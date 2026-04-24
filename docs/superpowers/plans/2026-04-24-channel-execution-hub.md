# Channel Execution Hub Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a channel-centered execution layer where lists and tool runs become native to channel context, Home gains cross-channel execution management blocks, and AI adds message-to-task suggestions plus channel execution summaries.

**Architecture:** Reuse existing list, tool-run, message, and Home aggregation surfaces. Extend backend contracts to persist message-linked list items, tool writeback targets, and execution summary aggregates. Keep Web delivery in Windsurf-owned surfaces that consume those contracts from the channel page and Home dashboard.

**Tech Stack:** Go, Gin, GORM, SQLite, existing structured workspace handlers, existing Home aggregation in `workspace.go`, Next.js, Zustand stores, existing channel/Home layouts.

**Spec:** `docs/superpowers/specs/2026-04-24-channel-execution-hub-design.md`

---

### Task 1: Backend Contract Tests For Channel Execution Hub

**Files:**
- Modify: `apps/api/internal/handlers/structured_workspace_test.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `TestChannelExecutionListItemPersistsMessageReference`.
- [ ] Add `TestChannelExecutionCreateListItemWithAIAssistFallsBackWithoutAI`.
- [ ] Add `TestChannelExecutionToolExecuteSupportsMessageWriteback`.
- [ ] Add `TestChannelExecutionToolExecuteSupportsListItemWriteback`.
- [ ] Add `TestChannelExecutionHomeIncludesOpenListWork`.
- [ ] Add `TestChannelExecutionHomeIncludesToolRunsNeedingAttention`.
- [ ] Add `TestChannelExecutionHomeIncludesChannelExecutionPulse`.
- [ ] Run `go test ./internal/handlers -run TestChannelExecution -count=1`.
- [ ] Confirm RED before implementation.

### Task 2: Persist Message References On Workspace List Items

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Modify: `apps/api/internal/handlers/structured_workspace.go`
- Modify: `apps/api/internal/handlers/structured_workspace_test.go`

- [ ] Extend `WorkspaceListItem` with nullable source-message fields for at least `message_id`, `channel_id`, and source snippet.
- [ ] Expose those fields through list-item hydration responses.
- [ ] Accept optional source-message input on `POST /api/v1/lists/:id/items`.
- [ ] Preserve existing manual list-item creation when no source message is present.
- [ ] Run `go test ./internal/handlers -run TestChannelExecutionListItemPersistsMessageReference -count=1`.

### Task 3: Add AI-Assisted Message-To-List Draft Endpoint

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/handlers/test_helpers_test.go`
- Modify: `apps/api/internal/handlers/ai_test.go`
- Modify: `apps/api/main.go`

- [ ] Add a narrow endpoint for message-to-list suggestions using existing AI gateway patterns.
- [ ] Input should include at minimum `message_id`, `list_id`, optional `channel_id`, and source text/context.
- [ ] Output should contain only suggestion fields needed for the form: normalized title, assignee candidate, due-date candidate, and rationale if already idiomatic.
- [ ] Make AI failure return a soft, non-blocking error shape so Web can fall back to manual entry.
- [ ] Run `go test ./internal/handlers -run TestChannelExecutionCreateListItemWithAIAssistFallsBackWithoutAI -count=1`.

### Task 4: Extend Tool Execution For Channel Writeback Targets

**Files:**
- Modify: `apps/api/internal/handlers/structured_workspace.go`
- Modify: `apps/api/internal/handlers/structured_workspace_test.go`
- Modify: `apps/api/internal/domain/models.go`

- [ ] Extend tool execution input to accept an explicit writeback target.
- [ ] Support exactly two first-phase targets: `message` and `list_item`.
- [ ] Define the request contract so `writeback_target=message` accepts the destination `channel_id` and optional `thread_id`, while `writeback_target=list_item` accepts the destination `list_id`.
- [ ] For `message`, persist a normal channel message whose metadata records the originating `tool_run_id` and `tool_id`, so the writeback is durable and visible in existing message surfaces.
- [ ] For `list_item`, create a new list item whose content is derived from the tool result and whose metadata/source fields retain the originating `tool_run_id`; do not introduce an "attach to existing item" mode in this phase.
- [ ] Fail invalid target combinations with a stable `400` contract.
- [ ] Run `go test ./internal/handlers -run 'TestChannelExecutionToolExecuteSupports(MessageWriteback|ListItemWriteback)' -count=1`.

### Task 5: Extend Home Aggregation With Execution Blocks

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `open_list_work[]` aggregation with channel, assignee, due-date, and status fields.
- [ ] Add `tool_runs_needing_attention[]` aggregation with running/failed/recent-review-worthy runs.
- [ ] Add `channel_execution_pulse[]` aggregation with open-item count, overdue count, recent tool activity, and summary text slot.
- [ ] Keep `GET /api/v1/home` backward-compatible for existing consumers.
- [ ] Run `go test ./internal/handlers -run 'TestChannelExecutionHomeIncludes(OpenListWork|ToolRunsNeedingAttention|ChannelExecutionPulse)' -count=1`.

### Task 6: Add Channel Execution Summary Aggregation

**Files:**
- Modify: `apps/api/internal/handlers/workspace.go`
- Modify: `apps/api/internal/handlers/structured_workspace.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add a lightweight summary builder that combines recent list-item changes, overdue state, and recent tool-run outcomes per channel.
- [ ] Prefer reusing Home/channel aggregate code paths rather than creating a heavyweight new subsystem.
- [ ] Ensure summary generation has stable behavior for empty channels, active channels, and failure-heavy channels.
- [ ] Surface summary text in Home pulse rows and leave room for channel-side consumption by Windsurf.
- [ ] Run `go test ./internal/handlers -run TestChannelExecutionHomeIncludesChannelExecutionPulse -count=1`.

### Task 7: Windsurf Web Track - Channel Execution Surface

**Files:**
- Modify: `apps/web/app/workspace/page.tsx`
- Create: `apps/web/components/channel/channel-execution-panel.tsx`
- Create: `apps/web/components/channel/channel-lists-panel.tsx`
- Create: `apps/web/components/channel/channel-tools-panel.tsx`
- Modify: `apps/web/stores/list-store.ts`
- Modify: `apps/web/stores/tool-store.ts`
- Modify: `apps/web/types/index.ts`

- [ ] Windsurf adds an `Execution` entry point in channel context alongside existing secondary panels.
- [ ] Windsurf renders `Lists` and `Tools` tabs without breaking message-area layout.
- [ ] Windsurf wires channel-scoped list fetching and tool-run fetching into the new panel.
- [ ] Windsurf adds channel-native `Create List` and `Run Tool` actions inside the execution panel rather than forcing users back to standalone pages.
- [ ] Windsurf consumes backend message-reference and tool-writeback fields without inventing alternate contracts.
- [ ] Windsurf verifies empty/loading/error states in the channel execution panel.

### Task 8: Windsurf Web Track - Message-To-List UX

**Files:**
- Modify: `apps/web/components/message/message-actions.tsx`
- Create: `apps/web/components/message/add-to-list-dialog.tsx`
- Modify: `apps/web/stores/list-store.ts`
- Modify: `apps/web/hooks/use-ai-chat.ts` or a more appropriate AI request hook if needed

- [ ] Windsurf adds `Add to List` to message actions.
- [ ] Windsurf allows selecting a current-channel list and requesting AI draft suggestions.
- [ ] Windsurf falls back to manual entry when the AI suggestion endpoint fails or is unavailable.
- [ ] Windsurf persists source-message metadata when creating the list item.
- [ ] Windsurf verifies the saved list item deep-links or visibly references the originating message.

### Task 9: Windsurf Web Track - Home Execution Blocks

**Files:**
- Modify: `apps/web/components/layout/home-dashboard.tsx`
- Modify: `apps/web/stores/workspace-store.ts`
- Modify: `apps/web/types/index.ts`

- [ ] Windsurf adds `Open List Work`, `Tool Runs Needing Attention`, and `Channel Execution Pulse` blocks to Home.
- [ ] Windsurf consumes new Home payload fields without regressing existing cards.
- [ ] Windsurf deep-links each card back into the correct channel/object context.
- [ ] Windsurf renders channel execution summary text where available and degrades cleanly when absent.

### Task 10: Documentation, Collaboration Handoff, And Verification

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Create: `docs/releases/v0.6.xx.md`

- [ ] Document message-linked list items, tool writeback targets, Home execution blocks, and AI-assist fallback behavior.
- [ ] Record the execution ownership split clearly: Windsurf owns Web, Codex owns backend/contracts/tests.
- [ ] Leave Windsurf a concise contract handoff covering list item source-message fields, AI suggestion response shape, Home aggregation fields, and tool writeback payloads.
- [ ] Run `go test ./internal/handlers -count=1`.
- [ ] Run `go test ./...`.
- [ ] Run `GOCACHE=$(pwd)/.cache/go-build go build ./...`.
- [ ] Run `git diff --check`.
- [ ] After Windsurf lands Web work, run the relevant Web verification commands they specify or the repo-standard `pnpm` checks.

---

## Execution Notes

- Treat this as three coordinated tracks with disjoint ownership:
  - `Gemini`: Tasks 1-6 and backend side of Task 10
  - `Windsurf`: Tasks 7-9 and Web side of Task 10
  - `Codex`: planning, cross-agent contract review, handoff generation, and final integration control only
- Keep the first release narrow. Do not introduce multi-step workflow orchestration, project hierarchies, or automatic AI state mutation.
- The AI layer is advisory only. Manual list-item creation and structured rendering must work without AI.
- Preserve Home backward compatibility so existing consumers continue to render while Windsurf adopts new blocks.
- Prefer `gpt-5.4-mini` subagents for isolated implementation and test slices owned by Gemini or Windsurf, while Codex stays out of overlapping write scopes.
