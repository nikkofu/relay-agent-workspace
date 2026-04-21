# Phase 46 Knowledge Entities and Wiki Substrate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add first-class knowledge entities, refs, links, and timeline APIs so Relay can power wiki-style entity pages and hydrate citation `entity_title` fields.

**Architecture:** Extend the existing `internal/knowledge` package from citation lookup into a bounded knowledge substrate service. Persist `knowledge_entities`, `knowledge_entity_refs`, `knowledge_entity_links`, and `knowledge_events` with GORM, expose REST APIs under `/api/v1/knowledge`, and keep Phase 47 live ingestion/advanced graph expansion out of this slice.

**Tech Stack:** Go, Gin, GORM, SQLite, existing `internal/knowledge`, existing realtime hub, existing `ids.NewPrefixedUUID`.

---

## File Map

### Existing files to modify

- `apps/api/internal/domain/models.go`
  - add `KnowledgeEntity`, `KnowledgeEntityRef`, `KnowledgeEntityLink`, and `KnowledgeEvent`
- `apps/api/internal/db/db.go`
  - migrate the new knowledge models
- `apps/api/internal/knowledge/types.go`
  - add entity response, ref response, link response, timeline response, and graph preview types
- `apps/api/internal/knowledge/service.go`
  - add entity CRUD, ref/link/timeline helpers, graph preview, and citation entity-title hydration
- `apps/api/internal/knowledge/service_test.go`
  - add package tests for entity creation, refs, timeline, and citation title hydration
- `apps/api/internal/handlers/files.go`
  - keep citation lookup response behavior, now with hydrated `entity_title`
- `apps/api/internal/handlers/collaboration_test.go`
  - add HTTP API tests for knowledge entities, refs, timeline, and graph preview
- `apps/api/main.go`
  - register `/api/v1/knowledge/*` routes
- `README.md`
  - add Phase 46 API surface
- `CHANGELOG.md`
  - document `v0.5.86`
- `docs/AGENT-COLLAB.md`
  - update task board and Windsurf handoff
- `apps/web/components/agent-collab/agent-collab-data.ts`
  - update static fallback task board and comm log
- `package.json`
  - bump version to `0.5.86`
- `apps/web/package.json`
  - bump version to `0.5.86`

### New files to create

- `apps/api/internal/handlers/knowledge.go`
  - HTTP handlers for entity CRUD, refs, links, timeline, and graph preview
- `docs/releases/v0.5.86.md`
  - release note and Windsurf handoff

## Task 1: Add knowledge entity persistence models

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing persistence test**

Add `TestKnowledgeEntityModelsPersist` to `apps/api/internal/handlers/collaboration_test.go`.

Expected behavior:
- `KnowledgeEntity` persists with `id`, `workspace_id`, `kind`, `title`, `summary`, `status`
- `KnowledgeEntityRef` persists and points to a file/message/artifact
- `KnowledgeEntityLink` persists and relates two entities
- `KnowledgeEvent` persists with `event_type` and `occurred_at`

- [ ] **Step 2: Run focused test and verify RED**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEntityModelsPersist$'
```

Expected:
- FAIL because models do not exist.

- [ ] **Step 3: Add domain models**

Add:
- `KnowledgeEntity`
- `KnowledgeEntityRef`
- `KnowledgeEntityLink`
- `KnowledgeEvent`

Use prefixed string IDs for all primary keys:
- `entity-*`
- `kref-*`
- `klink-*`
- `kevent-*`

- [ ] **Step 4: Migrate models**

Update:
- `apps/api/internal/db/db.go`
- `setupTestDB` in `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 5: Run focused test and verify GREEN**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEntityModelsPersist$'
```

Expected:
- PASS

## Task 2: Add entity service and citation title hydration

**Files:**
- Modify: `apps/api/internal/knowledge/types.go`
- Modify: `apps/api/internal/knowledge/service.go`
- Modify: `apps/api/internal/knowledge/service_test.go`

- [ ] **Step 1: Write failing service tests**

Add tests:
- `TestCreateKnowledgeEntityDefaults`
- `TestAddEntityRefAppendsTimelineEvent`
- `TestLookupHydratesEntityTitle`
- `TestEntityGraphPreviewIncludesRefsAndLinks`

- [ ] **Step 2: Run package tests and verify RED**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge -run 'Test(CreateKnowledgeEntityDefaults|AddEntityRefAppendsTimelineEvent|LookupHydratesEntityTitle|EntityGraphPreviewIncludesRefsAndLinks)'
```

Expected:
- FAIL because service helpers do not exist.

- [ ] **Step 3: Implement service helpers**

Add helpers:
- `CreateEntity`
- `UpdateEntity`
- `ListEntities`
- `GetEntity`
- `AddEntityRef`
- `ListEntityRefs`
- `AddEntityLink`
- `ListEntityLinks`
- `AppendEntityEvent`
- `ListEntityTimeline`
- `BuildEntityGraph`

Also update citation lookup entity hydration:
- when `KnowledgeEvidenceEntityRef.EntityID` is present, load `KnowledgeEntity.Title`
- set `Citation.EntityTitle`

- [ ] **Step 4: Run package tests and verify GREEN**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge
```

Expected:
- PASS

## Task 3: Add knowledge REST APIs

**Files:**
- Create: `apps/api/internal/handlers/knowledge.go`
- Modify: `apps/api/main.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write failing API tests**

Add tests:
- `TestKnowledgeEntityCRUDEndpoints`
- `TestKnowledgeEntityRefsAndTimelineEndpoints`
- `TestKnowledgeEntityGraphEndpoint`

Expected API surface:
- `GET /api/v1/knowledge/entities`
- `POST /api/v1/knowledge/entities`
- `GET /api/v1/knowledge/entities/:id`
- `PATCH /api/v1/knowledge/entities/:id`
- `GET /api/v1/knowledge/entities/:id/refs`
- `POST /api/v1/knowledge/entities/:id/refs`
- `GET /api/v1/knowledge/entities/:id/timeline`
- `POST /api/v1/knowledge/entities/:id/events`
- `GET /api/v1/knowledge/entities/:id/links`
- `POST /api/v1/knowledge/links`
- `GET /api/v1/knowledge/entities/:id/graph`

- [ ] **Step 2: Run focused API tests and verify RED**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEntity(CRUD|Refs|Graph)'
```

Expected:
- FAIL because handlers/routes do not exist.

- [ ] **Step 3: Implement handlers**

Create `apps/api/internal/handlers/knowledge.go`.

Handler behavior:
- return `400` on missing required fields
- return `404` on unknown entity
- return `201` for create/ref/link/event append
- use `currentUser` as actor where available for timeline events
- keep response keys simple:
  - `{ "entities": [...] }`
  - `{ "entity": ... }`
  - `{ "refs": [...] }`
  - `{ "links": [...] }`
  - `{ "events": [...] }`
  - `{ "graph": { "nodes": [...], "edges": [...] } }`

- [ ] **Step 4: Register routes**

Update `apps/api/main.go` under `v1`.

- [ ] **Step 5: Run focused API tests and verify GREEN**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEntity(CRUD|Refs|Graph)'
```

Expected:
- PASS

## Task 4: Update release docs and collaboration handoff

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `apps/web/components/agent-collab/agent-collab-data.ts`
- Modify: `package.json`
- Modify: `apps/web/package.json`
- Create: `docs/releases/v0.5.86.md`

- [ ] **Step 1: Update version and release docs**

Set current release to `v0.5.86`.

- [ ] **Step 2: Add Windsurf handoff**

Document that Windsurf should build:
- entity detail/wiki page
- entity badge hydration from `entity_title`
- entity refs/timeline panels
- graph relationship preview badges

- [ ] **Step 3: Verify docs mention the new API list**

Ensure docs include all Phase 46 endpoints.

## Task 5: Full verification and release

**Files:**
- Verify only

- [ ] **Step 1: Run backend package tests**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge ./internal/handlers
```

Expected:
- PASS

- [ ] **Step 2: Run full backend tests**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...
```

Expected:
- PASS

- [ ] **Step 3: Run backend build**

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...
```

Expected:
- PASS

- [ ] **Step 4: Run frontend lint**

```bash
pnpm --filter relay-agent-workspace lint
```

Expected:
- PASS

- [ ] **Step 5: Commit, tag, and push**

```bash
git add ...
git commit -m "feat: add knowledge entity wiki substrate"
git tag v0.5.86
git push origin main
git push origin v0.5.86
```

Expected:
- remote `main` and tag `v0.5.86` point to the release commit.

## Notes

- Do not implement `POST /api/v1/knowledge/events/ingest` in Phase 46. That belongs to Phase 47.
- Keep graph API as a preview generated from explicit refs and links. Do not add a graph database.
- Keep wiki content on the entity summary/metadata/ref/timeline model for now. Do not introduce a separate wiki page table yet.
