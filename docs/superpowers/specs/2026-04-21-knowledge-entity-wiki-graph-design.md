# Phase 45-47: Knowledge Entities, Wiki, and Graph Design

## Summary

Relay Agent Workspace should not split knowledge into two disconnected systems for "static docs" and "dynamic events". The better model is a unified knowledge substrate with two primary collaboration axes and an explicit time layer:

- `artifact/file`
- `channel/message`
- time layer:
  - `snapshot`
  - `live_reference`
  - `event_history`

This design turns business entities into first-class objects that can connect channels, messages, files, artifacts, workflows, users, and later external live data sources. It gives the product a stable base for:

- AI citation lookup
- internal wiki pages
- relationship graph views
- real-time business data flowing into collaboration surfaces

The recommended implementation sequence is:

1. Phase 45: AI citation lookup with entity-aware evidence
2. Phase 46: knowledge entities and wiki substrate
3. Phase 47: graph API and live event ingestion

## Goals

- establish a first-class business-entity model for workspace knowledge
- let messages, files, artifacts, channels, workflows, and users attach to entities through explicit references
- support wiki-style knowledge pages without inventing a second storage model
- support graph-style relationship exploration without requiring the frontend to infer links
- prepare for real-time external business data to enter the workspace as entity updates instead of isolated message blobs
- keep the design compatible with existing file extraction and citation APIs

## Non-Goals

- building a complete BI or analytics platform
- replacing the existing message, file, or artifact stores
- shipping a full semantic knowledge graph engine in the first phase
- solving embeddings, vector retrieval, or advanced ontology management in this slice

## Why Entity-First Is Recommended

Three approaches were considered:

1. Entity-first, recommended
   Business entities such as projects, tasks, tickets, customers, incidents, services, and datasets become first-class knowledge objects. Events, messages, files, and AI outputs attach to these entities.

2. Event-first
   Real-time business data is first ingested as activity or message streams, with entities inferred later.

3. Full hybrid from day one
   Build entities, events, wiki, and graph together in a single large phase.

Entity-first is recommended because it produces a stable center of gravity for collaboration. The UI can still show real-time event flows, but those flows update or reference a durable knowledge object instead of becoming isolated chat noise. This fits Slack-like collaboration while also enabling AI-native retrieval and workspace knowledge organization.

## Knowledge Model

### Primary model

Use a unified knowledge layer above the existing channel, message, file, artifact, and workflow stores.

Core concepts:

- `knowledge_entity`
  A durable business or knowledge object.
- `knowledge_entity_link`
  A typed relationship between entities.
- `knowledge_entity_ref`
  A typed binding between an entity and an existing system object.
- `knowledge_event`
  A timeline event associated with an entity.

### `knowledge_entities`

Suggested fields:

- `id`
- `workspace_id`
- `kind`
  - `project`
  - `task`
  - `ticket`
  - `customer`
  - `incident`
  - `service`
  - `doc`
  - `dataset`
  - `custom`
- `title`
- `summary`
- `status`
- `owner_user_id`
- `source_kind`
  - `manual`
  - `imported`
  - `derived`
  - `live`
- `source_ref`
- `metadata_json`
- `created_at`
- `updated_at`

### `knowledge_entity_links`

Suggested fields:

- `id`
- `workspace_id`
- `from_entity_id`
- `to_entity_id`
- `relation`
  - `references`
  - `depends_on`
  - `blocks`
  - `relates_to`
  - `owned_by`
  - `documented_by`
  - `discussed_in`
  - `updated_by`
- `weight`
- `metadata_json`
- `created_at`

### `knowledge_entity_refs`

Suggested fields:

- `id`
- `workspace_id`
- `entity_id`
- `ref_kind`
  - `channel`
  - `message`
  - `thread`
  - `file`
  - `artifact`
  - `workflow`
  - `tool_run`
  - `user`
- `ref_id`
- `role`
  - `primary`
  - `evidence`
  - `discussion`
  - `output`
  - `owner`
- `metadata_json`
- `created_at`

### `knowledge_events`

Suggested fields:

- `id`
- `workspace_id`
- `entity_id`
- `event_type`
  - `created`
  - `status_changed`
  - `message_linked`
  - `file_linked`
  - `artifact_linked`
  - `workflow_run`
  - `live_update`
  - `ai_summary`
- `title`
- `body`
- `actor_user_id`
- `source_kind`
  - `manual`
  - `system`
  - `imported`
  - `ai`
- `source_ref`
- `occurred_at`
- `created_at`

## Time Layer

The system should not treat all knowledge as static. Each entity and evidence object can participate in one or more temporal modes:

- `snapshot`
  Stable state captured at a moment in time, such as a file extraction, artifact version, or generated summary.
- `live_reference`
  A pointer to something that changes over time, such as a connected business object or external feed.
- `event_history`
  A sequence of updates, discussions, and derived outputs that explain how the entity evolved.

This time layer is important because later business data ingestion should update entity timelines cleanly instead of forcing everything into plain messages.

## Citation and Evidence Substrate

Phase 45 should extend the existing citation model into a broader evidence layer.

Citation results should eventually support:

- `file_chunk`
- `message`
- `thread`
- `artifact_section`
- `entity_summary`

Recommended payload additions:

- `entity_id`
- `entity_title`
- `source_kind`
- `ref_kind`
- `locator`
- `snippet`
- `score`

This allows AI flows, wiki pages, and graph panels to share the same evidence substrate.

## Recommended Phase Breakdown

### Phase 45: AI Citation Lookup and Entity-Aware Evidence

Purpose:

- expand file citations into a unified evidence lookup layer
- prepare citation payloads for entity binding

Scope:

- citation lookup across file chunks, messages, threads, and artifact sections
- optional `entity_id` in citation results
- consistent evidence payloads for UI and AI execution layers

Key outputs:

- citation lookup endpoint
- evidence payload contract
- entity-aware response fields

### Phase 46: Knowledge Entities and Wiki Substrate

Purpose:

- make business entities first-class objects
- provide the API foundation for internal wiki pages

Scope:

- entity CRUD
- entity refs
- entity links
- entity timeline
- enough detail payload to power a first wiki page

Key outputs:

- `knowledge_entities`
- `knowledge_entity_links`
- `knowledge_entity_refs`
- `knowledge_events`
- entity details, refs, and timeline APIs

### Phase 47: Graph API and Live Event Ingestion

Purpose:

- connect entities to live business updates and relationship views

Scope:

- graph endpoint for nodes and edges
- live event ingestion
- realtime broadcasts for entity updates
- timeline updates from external or internal event sources

Key outputs:

- graph payload contract
- ingestion endpoint
- entity update realtime events

## API Design

### Entity APIs

- `GET /api/v1/knowledge/entities`
- `POST /api/v1/knowledge/entities`
- `GET /api/v1/knowledge/entities/:id`
- `PATCH /api/v1/knowledge/entities/:id`

These APIs should expose entity identity, summary, status, ownership, source metadata, and counts for refs, links, and timeline items when available.

### Timeline APIs

- `GET /api/v1/knowledge/entities/:id/timeline`
- `POST /api/v1/knowledge/entities/:id/events`

The timeline is the unified activity history for wiki surfaces, entity inspectors, and later business live-data views.

### Reference APIs

- `GET /api/v1/knowledge/entities/:id/refs`
- `POST /api/v1/knowledge/entities/:id/refs`

References explicitly connect an entity to existing system objects instead of forcing the frontend to guess relationships.

### Link APIs

- `GET /api/v1/knowledge/entities/:id/links`
- `POST /api/v1/knowledge/links`

These APIs expose entity-to-entity relationships and let graph views consume typed edges directly.

### Graph APIs

- `GET /api/v1/knowledge/entities/:id/graph`
- `GET /api/v1/knowledge/graph`

Recommended response shape:

```json
{
  "nodes": [
    {
      "id": "entity-...",
      "kind": "project",
      "title": "Q2 Launch"
    }
  ],
  "edges": [
    {
      "id": "link-...",
      "from": "entity-...",
      "to": "entity-...",
      "relation": "depends_on"
    }
  ]
}
```

### Ingestion API

- `POST /api/v1/knowledge/events/ingest`

This endpoint is for business-domain updates coming from adapters or future integrations. The first version should be simple and explicit:

- identify the target entity or enough source metadata to resolve one
- write a timeline event
- optionally create or update a reference
- optionally emit a realtime event

## Data Flow

Recommended operational flow:

1. A user creates or imports a `knowledge_entity`.
2. Messages, files, artifacts, channels, workflows, and users are attached through `knowledge_entity_refs`.
3. AI citation lookup returns evidence that may already bind to an entity or may be promoted into one.
4. Wiki pages read entity detail, refs, links, and timeline data from the same backend model.
5. Graph panels read nodes and edges from graph APIs rather than reconstructing relationships client-side.
6. External business systems later post updates into ingestion APIs, which create `knowledge_events` and keep the entity timeline current.

## Realtime

The backend should eventually broadcast entity-centric events in the same spirit as the current collaboration model.

Recommended events:

- `knowledge.entity.created`
- `knowledge.entity.updated`
- `knowledge.entity.ref.created`
- `knowledge.entity.timeline.updated`

This allows wiki panels, graph views, entity inspectors, and channel sidebars to stay synchronized.

## Error Handling

The knowledge layer should fail explicitly:

- unknown entity
  - `404 not_found`
- invalid reference target
  - `400 invalid_ref`
- unsupported relation kind
  - `400 invalid_relation`
- ingestion payload cannot resolve target
  - `422 unresolved_entity`
- duplicate primary reference where uniqueness is required
  - `409 duplicate_ref`

## Testing

Backend testing should cover:

- entity CRUD validation
- reference creation and duplicate handling
- link creation and graph serialization
- timeline append behavior
- ingestion behavior for known and unresolved targets
- citation payload compatibility when `entity_id` is present or absent

## Windsurf Collaboration Boundary

Codex should own:

- schema and persistence
- entity, refs, links, timeline, ingestion, and graph APIs
- citation payload contract changes
- realtime entity update broadcasts

Windsurf should own:

- wiki entity page
- entity inspector surfaces in channel/message/file contexts
- graph view rendering and interaction
- entity-aware citation cards and evidence panels

## Acceptance Criteria

This design is successful when:

- the system can represent durable business entities without abusing channels or messages as the only source of truth
- entity pages can show refs, links, and timeline data from backend APIs alone
- citation lookup can return evidence that is entity-aware
- graph views can render from explicit node and edge payloads
- future business live-data updates can append to entity timelines without redesigning the model

## Recommendation

Start with Phase 45, but reserve the Phase 46 and Phase 47 fields now:

- `entity_id`
- `source_kind`
- `ref_kind`
- `source_ref`

That keeps the next phases additive. The evidence layer becomes the substrate, the entity model becomes the wiki backbone, and the graph plus ingestion layer becomes the bridge to AI-native real-time business collaboration.
