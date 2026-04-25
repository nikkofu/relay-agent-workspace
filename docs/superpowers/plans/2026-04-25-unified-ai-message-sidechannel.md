# Unified AI Message Side-Channel Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship one durable `metadata.ai_sidecar` contract for AI-generated messages across AI DM, channel `/ask`, and canvas AI, with normalized reasoning/tool/usage data that works during streaming and survives replay after reload.

**Architecture:** Gemini owns additive backend normalization, streaming envelope, persistence, and compatibility rollout. Windsurf owns multi-surface rendering and stream-to-persisted handoff logic. Legacy flat metadata fields remain readable during rollout, while `metadata.ai_sidecar` becomes the canonical shape.

**Tech Stack:** Go, Gin, GORM, existing LLM gateway and SSE streaming, existing DM/channel/canvas message surfaces, Next.js, Zustand, existing websocket and SSE consumers.

**Spec:** `docs/superpowers/specs/2026-04-25-unified-ai-message-sidechannel-design.md`

---

### Task 1: Gemini Backend Contract Tests For Unified AI Sidecar

**Files:**
- Modify: `apps/api/internal/handlers/ai_test.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`

- [ ] Add `TestUnifiedAISidecarExecuteAIStreamsNormativeKinds`.
- [ ] Add `TestUnifiedAISidecarPersistsCanonicalMetadata`.
- [ ] Add `TestUnifiedAISidecarDualReadsLegacyFlatFields`.
- [ ] Add `TestUnifiedAISidecarChannelAskPersistsReplayableSidecar`.
- [ ] Add `TestUnifiedAISidecarCanvasAIReplyPersistsReplayableSidecar`.
- [ ] Run `cd apps/api && go test ./internal/handlers -run TestUnifiedAISidecar -count=1`.
- [ ] Confirm RED before implementation.

### Task 2: Gemini Backend - Canonical `metadata.ai_sidecar` Normalization

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/domain/models.go` if helper structs are warranted
- Modify: `apps/api/internal/handlers/ai_test.go`

- [ ] Define the canonical persisted shape:
  - `metadata.ai_sidecar.reasoning`
  - `metadata.ai_sidecar.tool_calls`
  - `metadata.ai_sidecar.usage`
- [ ] Normalize provider/runtime output into that shape for AI-generated messages.
- [ ] Ensure missing optional sections do not prevent message persistence.
- [ ] Keep `metadata.ai_sidecar` as the source of truth whenever present.
- [ ] Run `cd apps/api && go test ./internal/handlers -run 'TestUnifiedAISidecar(PersistsCanonicalMetadata|ChannelAskPersistsReplayableSidecar|CanvasAIReplyPersistsReplayableSidecar)' -count=1`.

### Task 3: Gemini Backend - Normative Streaming Envelope

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/ai_test.go`

- [ ] Freeze the stream envelope as:
  - `{"kind":"reasoning|tool_call|usage|answer","message_id":"...","payload":{}}`
- [ ] Emit side-channel streaming items for AI DM, channel `/ask`, and canvas AI where data exists.
- [ ] Preserve answer streaming while adding side-channel kinds additively.
- [ ] Document or implement `message_id` behavior for provisional vs final message ids consistently.
- [ ] Run `cd apps/api && go test ./internal/handlers -run TestUnifiedAISidecarExecuteAIStreamsNormativeKinds -count=1`.

### Task 4: Gemini Backend - Mixed-Version Compatibility

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/ai_test.go`

- [ ] Dual-read legacy flat metadata fields:
  - `metadata.reasoning`
  - `metadata.tool_calls`
  - `metadata.usage`
- [ ] Synthesize equivalent sidecar views when `metadata.ai_sidecar` is absent.
- [ ] Dual-write legacy flat fields only if required for short-term client compatibility during rollout.
- [ ] Keep `ai_sidecar` canonical even when dual-writing.
- [ ] Run `cd apps/api && go test ./internal/handlers -run TestUnifiedAISidecarDualReadsLegacyFlatFields -count=1`.

### Task 5: Windsurf Web - AI DM Sidecar Consumption

**Files:**
- Modify: `apps/web/app/workspace/dms/[id]/page.tsx`
- Modify: `apps/web/stores/message-store.ts`
- Modify: `apps/web/types/index.ts`

- [ ] Prefer `metadata.ai_sidecar` over flat legacy metadata when rendering AI DM messages.
- [ ] Use legacy flat fields only as fallback when sidecar is absent.
- [ ] Replace token heuristics with real `usage` when present, while keeping heuristics as fallback only.
- [ ] Render reasoning/tool/usage from persisted messages after reload, not just live stream state.

### Task 6: Windsurf Web - Channel `/ask` Sidecar Consumption

**Files:**
- Modify: `apps/web/components/message/message-item.tsx`
- Modify: `apps/web/components/message/message-list.tsx`
- Modify: `apps/web/components/message/message-composer.tsx`
- Modify: `apps/web/types/index.ts`

- [ ] Render AI sidecar metadata on `/ask` replies using the same interpretation as AI DM.
- [ ] Ensure streamed `/ask` side-channel data hands off cleanly to persisted message metadata when the final message arrives.
- [ ] Keep non-AI human messages visually unaffected.

### Task 7: Windsurf Web - Canvas AI Sidecar Consumption

**Files:**
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`
- Modify: `apps/web/components/layout/canvas-panel.tsx`
- Modify: `apps/web/types/index.ts`

- [ ] Consume the normative stream envelope kinds in the canvas AI dock.
- [ ] Render reasoning/tool/usage from persisted `metadata.ai_sidecar` when reopening or reloading canvas AI history.
- [ ] Remove or demote UI heuristics wherever authoritative sidecar data exists.

### Task 8: Windsurf Web - Stream-To-Persisted Handoff Consistency

**Files:**
- Modify: `apps/web/hooks/use-websocket.ts` if needed
- Modify: `apps/web/stores/message-store.ts`
- Modify: `apps/web/components/dm/dm-conversation-page.tsx` or current DM renderer
- Modify: `apps/web/components/layout/canvas-ai-dock.tsx`

- [ ] Treat stream state as temporary preview only.
- [ ] Once the final persisted message arrives, replace transient reasoning/tool/usage state with stored `metadata.ai_sidecar`.
- [ ] Prevent duplicate tool rows or duplicate reasoning text when both stream events and final message metadata are present.

### Task 9: Codex Coordination, Docs, And Rollout Control

**Files:**
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `CHANGELOG.md`
- Create: `docs/releases/v0.6.51.md`

- [ ] Freeze the canonical sidecar contract and rollout rule in `docs/AGENT-COLLAB.md`.
- [ ] Record the mixed-version rule clearly:
  - dual-read immediately
  - optional dual-write during rollout
  - `ai_sidecar` remains canonical
- [ ] Publish the execution split:
  - Gemini Tasks 1-4
  - Windsurf Tasks 5-8
- [ ] After Gemini reports completion, verify backend with:
  - `cd apps/api && go test ./internal/handlers -run TestUnifiedAISidecar -count=1`
- [ ] After Windsurf reports completion, verify the Web checks they provide or repo-standard checks if supplied.
- [ ] Run `git diff --check`.

---

## Execution Notes

- Treat this as one contract phase with two implementation tracks and one control track:
  - `Gemini`: Tasks 1-4
  - `Windsurf`: Tasks 5-8
  - `Codex`: Task 9
- Preferred sequencing:
  - Gemini should freeze the canonical sidecar shape and stream envelope first
  - Windsurf can begin fallback-aware rendering refactors in parallel, but should not finalize stream parsers before Gemini lands the normative envelope
- Do not create surface-specific sidecar variants in this phase.
- Keep rollout additive; do not require a blocking migration of old AI messages before ship.
