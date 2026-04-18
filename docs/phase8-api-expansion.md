# Phase 8: Backend Expansion Target

## 1. Current Reality

`Relay Agent Workspace` is no longer a pure mock-only UI. The project now ships a working monorepo with:

- `apps/web`: Next.js workspace UI
- `apps/api`: Go + Gin backend
- SQLite persistence through GORM
- WebSocket realtime fanout
- LLM gateway with:
  - `openai`
  - `openai-compatible`
  - `openrouter`
  - `gemini`

Already delivered backend capabilities include:

- current user and users list
- organizations, teams, workspaces, channels
- channel messages and thread replies
- AI SSE execution
- AI config and user AI settings
- persisted reactions, pinning, later, unread, and AI feedback
- realtime events for:
  - `message.created`
  - `reaction.updated`
  - `message.updated`
  - `message.deleted`
  - `agent_collab.sync`

So the purpose of this document is no longer “what do we need to start backend work at all”.  
It is now: **what still needs to be added to reach the broader Phase 8 dynamic product target**.

## 2. UI Surfaces Driving The Remaining API Work

The remaining backend work is still driven directly by the shipped UI in `apps/web`:

- Workspace shell:
  - `apps/web/app/workspace/layout.tsx`
  - `apps/web/app/workspace/page.tsx`
- Message experience:
  - `apps/web/components/layout/message-area.tsx`
  - `apps/web/components/layout/thread-panel.tsx`
  - `apps/web/components/message/message-composer.tsx`
  - `apps/web/components/message/message-item.tsx`
  - `apps/web/components/message/message-actions.tsx`
- AI surfaces:
  - `apps/web/components/ai-chat/ai-chat-panel.tsx`
  - `apps/web/components/ai-chat/ai-thread-summary.tsx`
  - `apps/web/components/ai-chat/ai-slash-command.tsx`
  - `apps/web/hooks/use-ai-chat.ts`
- Navigation and productivity views:
  - `apps/web/app/workspace/activity/page.tsx`
  - `apps/web/app/workspace/dms/page.tsx`
  - `apps/web/app/workspace/later/page.tsx`
  - `apps/web/components/search/search-dialog.tsx`
  - `apps/web/components/layout/canvas-panel.tsx`

## 3. Shipped API Baseline

Current backend surface already available:

- `GET /ping`
- `GET /api/v1/me`
- `PATCH /api/v1/me/settings`
- `GET /api/v1/users`
- `GET /api/v1/orgs`
- `GET /api/v1/orgs/:id/teams`
- `POST /api/v1/orgs/:id/agents`
- `GET /api/v1/workspaces`
- `GET /api/v1/workspaces/:id/invites`
- `POST /api/v1/workspaces/:id/invites`
- `GET /api/v1/channels`
- `GET /api/v1/channels/:id/members`
- `POST /api/v1/channels/:id/members`
- `DELETE /api/v1/channels/:id/members/:userId`
- `PATCH /api/v1/channels/:id`
- `GET /api/v1/dms`
- `POST /api/v1/dms`
- `GET /api/v1/dms/:id/messages`
- `POST /api/v1/dms/:id/messages`
- `GET /api/v1/activity`
- `GET /api/v1/inbox`
- `GET /api/v1/mentions`
- `POST /api/v1/notifications/read`
- `GET /api/v1/later`
- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`
- `GET /api/v1/starred`
- `POST /api/v1/channels/:id/star`
- `GET /api/v1/pins`
- `GET /api/v1/drafts`
- `PUT /api/v1/drafts/:scope`
- `GET /api/v1/search`
- `GET /api/v1/messages`
- `GET /api/v1/messages/:id/thread`
- `POST /api/v1/messages`
- `DELETE /api/v1/messages/:id`
- `POST /api/v1/messages/:id/reactions`
- `POST /api/v1/messages/:id/pin`
- `POST /api/v1/messages/:id/later`
- `POST /api/v1/messages/:id/unread`
- `GET /api/v1/ai/config`
- `POST /api/v1/ai/execute`
- `POST /api/v1/ai/feedback`
- `GET /api/v1/realtime`

## 4. Remaining Backend Targets

### 4.1 Presence And Typing

Baseline support now exists for:

- `GET /api/v1/presence`
- `POST /api/v1/presence`
- `POST /api/v1/typing`
- realtime events:
  - `presence.updated`
  - `typing.updated`

Likely follow-ups:

- heartbeat / expiry semantics
- scoped presence subscriptions
- debounce and timeout conventions for typing indicators

### 4.2 AI Conversation State

The current AI path is execution-oriented and streaming-first. To support richer AI collaboration, we still need:

- AI conversation persistence
- channel or thread scoped AI history
- summaries and artifact outputs

Recommended additions:

- `GET /api/v1/ai/conversations`
- `GET /api/v1/ai/conversations/:id`
- `POST /api/v1/ai/threads/:id/summary`
- `POST /api/v1/ai/channels/:id/summary`
- `POST /api/v1/ai/commands/execute`

### 4.3 Canvas / Artifact Lifecycle

The canvas panel is already a meaningful product surface, but backend artifact lifecycle is still missing.

Recommended additions:

- `GET /api/v1/artifacts`
- `POST /api/v1/artifacts`
- `GET /api/v1/artifacts/:id`
- `PATCH /api/v1/artifacts/:id`
- `POST /api/v1/ai/canvas/generate`

### 4.4 Files

If message attachments and artifact references are to become real:

- `POST /api/v1/files/upload`
- `GET /api/v1/files/:id`

### 4.5 Slack Parity Layer

The current workspace is already functionally dynamic, but there is still a layer of classic Slack-style collaboration APIs that should be planned explicitly instead of being discovered ad hoc.

Recommended additions:

- channel membership and directory:
  - `GET /api/v1/channels/:id/members`
  - `POST /api/v1/channels/:id/members`
  - `DELETE /api/v1/channels/:id/members/:userId`
- channel metadata and admin actions:
  - `PATCH /api/v1/channels/:id`
  - fields such as:
    - `topic`
    - `purpose`
    - `is_archived`
- invitations:
  - `POST /api/v1/workspaces/:id/invites`
  - `GET /api/v1/workspaces/:id/invites`
- sidebar and discovery surfaces:
  - shipped baseline:
    - `GET /api/v1/starred`
    - `POST /api/v1/channels/:id/star`
    - `GET /api/v1/pins`
  - likely follow-up:
    - pinned filters by channel or user
    - star sorting and manual ordering
- notification and inbox surfaces:
  - shipped baseline:
    - `GET /api/v1/inbox`
    - `GET /api/v1/mentions`
    - `POST /api/v1/notifications/read`
  - likely follow-up:
    - notification preferences
    - mute rules
    - read-state websocket sync
- drafts:
  - shipped baseline:
    - `GET /api/v1/drafts`
    - `PUT /api/v1/drafts/:scope`
  - likely follow-up:
    - `DELETE /api/v1/drafts/:scope`

## 5. Realtime Target State

Current websocket support is enough for message creation and interaction sync.  
The broader target should cover:

- `message.created`
- `message.updated`
- `message.deleted`
- `reaction.updated`
- `presence.updated`
- `typing.updated`
- `thread.updated`
- `ai.delta`
- `ai.reasoning`
- `agent_collab.sync`
- later agent runtime events such as:
  - `agent.run.updated`
  - `agent.step.updated`
  - `artifact.updated`

## 6. Recommended Next Phases

Recommended sequence from here:

1. Expand realtime into presence, typing, and richer thread updates.
2. Add the Slack parity layer:
   - members
   - invites
   - channel metadata
   - inbox and mentions
   - drafts and stars
3. Persist AI conversation state and summaries.
4. Introduce artifact and file lifecycle APIs.
5. Add richer search layers such as suggestions and semantic retrieval.
6. Move toward explicit agent runtime APIs once the collaboration foundation is stable.

## 7. Summary

Phase 8 is no longer about “starting a backend”. That phase is already underway and shipped.  
The broader target now is to evolve Relay from:

- a realtime AI-native workspace with foundational APIs

into:

- a full collaboration platform for humans and agents, with search, history, artifacts, agent execution, and richer realtime semantics.
