# Phase 47 Knowledge Live Events and Ingestion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add realtime knowledge entity updates, live event ingestion, richer graph payloads, and deterministic entity auto-linking from messages/files.

**Architecture:** Extend the existing `internal/knowledge` service and `handlers/knowledge.go` without introducing a graph database or full LLM entity extraction yet. Use explicit refs/links/events as the source of truth, broadcast websocket events from handlers, add an ingestion endpoint for future business streams, and implement deterministic auto-linking by matching entity titles against message/file text as the first AI-ready seam.

**Tech Stack:** Go, Gin, GORM, SQLite, existing realtime hub, existing `internal/knowledge`, existing file/message models.

---

## File Map

### Existing files to modify

- `apps/api/internal/knowledge/types.go`
  - add ingestion and auto-link request/response types
  - enrich graph node/edge types with metadata, weight, direction, and ref role
- `apps/api/internal/knowledge/service.go`
  - add `IngestEvent`, `AutoLinkEntitiesForMessage`, `AutoLinkEntitiesForFile`, and richer graph serialization
- `apps/api/internal/handlers/knowledge.go`
  - broadcast `knowledge.entity.created`, `knowledge.entity.updated`, `knowledge.entity.ref.created`, `knowledge.event.created`, `knowledge.link.created`
  - add ingestion and auto-link endpoints
- `apps/api/internal/handlers/collaboration.go`
  - trigger deterministic auto-link after new channel messages are persisted
- `apps/api/internal/handlers/files.go`
  - trigger deterministic auto-link after file upload/extraction is persisted
- `apps/api/internal/handlers/collaboration_test.go`
  - add websocket, ingestion, graph, message/file auto-link tests
- `apps/api/main.go`
  - register Phase 47 routes
- `README.md`
  - add Phase 47 API surface
- `CHANGELOG.md`
  - document `v0.5.87`
- `docs/AGENT-COLLAB.md`
  - add handoff for Windsurf
- `apps/web/components/agent-collab/agent-collab-data.ts`
  - update static fallback collaboration data
- `package.json`
  - bump version to `0.5.87`
- `apps/web/package.json`
  - bump version to `0.5.87`

### New files to create

- `docs/releases/v0.5.87.md`
  - release note and Windsurf handoff

## Task 1: Add realtime knowledge websocket broadcasts

**Files:**
- Modify: `apps/api/internal/handlers/knowledge.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing websocket test**

Add `TestKnowledgeEntityEndpointsBroadcastRealtimeEvents`.

Expected events:
- `knowledge.entity.created`
- `knowledge.entity.updated`
- `knowledge.entity.ref.created`
- `knowledge.event.created`
- `knowledge.link.created`

- [ ] **Step 2: Run test and verify RED**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEntityEndpointsBroadcastRealtimeEvents$'
```

Expected:
- FAIL because handlers do not broadcast these events yet.

- [ ] **Step 3: Implement broadcast helper**

Add a helper in `handlers/knowledge.go`:

```go
func broadcastKnowledgeRealtime(eventType, entityID string, payload any)
```

Use `realtime.Event{Type, EntityID, Payload}`.

- [ ] **Step 4: Broadcast from create/update/ref/event/link handlers**

- [ ] **Step 5: Run focused test and verify GREEN**

## Task 2: Add live event ingestion endpoint

**Files:**
- Modify: `apps/api/internal/knowledge/types.go`
- Modify: `apps/api/internal/knowledge/service.go`
- Modify: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/main.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing ingestion test**

Add `TestKnowledgeEventIngestCreatesTimelineEventAndBroadcasts`.

Endpoint:
- `POST /api/v1/knowledge/events/ingest`

Request:

```json
{
  "entity_id": "entity-1",
  "event_type": "live_update",
  "title": "CRM account updated",
  "body": "ACME moved to renewal risk",
  "source_kind": "live",
  "source_ref": "crm:account:acme"
}
```

Expected:
- creates `KnowledgeEvent`
- broadcasts `knowledge.event.created`
- returns `{ "event": ... }`

- [ ] **Step 2: Run test and verify RED**

- [ ] **Step 3: Implement `IngestEvent` service helper**

- [ ] **Step 4: Implement handler and route**

- [ ] **Step 5: Run focused test and verify GREEN**

## Task 3: Enrich graph preview payload

**Files:**
- Modify: `apps/api/internal/knowledge/types.go`
- Modify: `apps/api/internal/knowledge/service.go`
- Test: `apps/api/internal/knowledge/service_test.go`

- [ ] **Step 1: Write failing graph test**

Add `TestEntityGraphPreviewIncludesTypedRefsWeightsAndDirection`.

Expected graph payload:
- node fields:
  - `id`
  - `kind`
  - `title`
  - `source_kind`
  - `ref_kind`
- edge fields:
  - `id`
  - `from`
  - `to`
  - `relation`
  - `weight`
  - `direction`
  - `role`

- [ ] **Step 2: Run test and verify RED**

- [ ] **Step 3: Enrich `GraphNode` and `GraphEdge`**

- [ ] **Step 4: Update graph builder**

- [ ] **Step 5: Run package tests and verify GREEN**

## Task 4: Add deterministic entity auto-linking from messages/files

**Files:**
- Modify: `apps/api/internal/knowledge/types.go`
- Modify: `apps/api/internal/knowledge/service.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/internal/handlers/files.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing auto-link tests**

Add:
- `TestCreateMessageAutoLinksMentionedKnowledgeEntity`
- `TestUploadFileAutoLinksMentionedKnowledgeEntity`

Behavior:
- if message/file text contains an entity title, create a `KnowledgeEntityRef`
- message ref: `ref_kind=message`, `role=discussion`
- file ref: `ref_kind=file`, `role=evidence`
- create a timeline event
- broadcast `knowledge.entity.ref.created`

- [ ] **Step 2: Run tests and verify RED**

- [ ] **Step 3: Implement deterministic matching service helpers**

Match rules:
- case-insensitive
- entity title length must be at least 3 characters
- avoid duplicate refs for same `entity_id/ref_kind/ref_id`

- [ ] **Step 4: Call from message create and file upload flows**

- [ ] **Step 5: Run focused tests and verify GREEN**

## Task 5: Update docs and release

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `apps/web/components/agent-collab/agent-collab-data.ts`
- Modify: `package.json`
- Modify: `apps/web/package.json`
- Create: `docs/releases/v0.5.87.md`

- [ ] **Step 1: Update release docs**

Document:
- websocket events
- ingestion endpoint
- richer graph payload
- deterministic entity auto-linking

- [ ] **Step 2: Add Windsurf handoff**

Ask Windsurf to:
- listen for `knowledge.entity.*` and `knowledge.event.created`
- refresh knowledge list/detail/graph on events
- render richer graph edge weights and typed refs
- surface auto-linked refs in message/file cards

## Task 6: Full verification and release

- [ ] **Step 1: Run focused backend tests**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge ./internal/handlers
```

- [ ] **Step 2: Run full backend tests**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...
```

- [ ] **Step 3: Run backend build**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...
```

- [ ] **Step 4: Run frontend lint**

```bash
pnpm --filter relay-agent-workspace lint
```

- [ ] **Step 5: Commit, tag, and push**

```bash
git add ...
git commit -m "feat: add knowledge live events and ingestion"
git tag v0.5.87
git push origin main
git push origin v0.5.87
```
