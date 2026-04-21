# Phase 45 AI Citation Lookup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a unified AI citation lookup API that returns entity-aware evidence from file chunks, messages, threads, and artifact sections without changing the existing Phase 44 file extraction contract.

**Architecture:** Keep handlers thin and introduce a focused `apps/api/internal/knowledge` package for evidence lookup, citation normalization, and future entity integration seams. Persist only the minimum new data needed in Phase 45, reuse existing file extraction chunks and message/artifact stores, and reserve `entity_id`, `source_kind`, `ref_kind`, and `source_ref` in the citation payload so Phase 46 and Phase 47 stay additive.

**Tech Stack:** Go, Gin, GORM, SQLite, existing `internal/fileindex` package, existing artifact/message/file models, existing realtime hub, new `internal/knowledge` package.

---

## File Map

### Existing files to modify

- `apps/api/internal/domain/models.go`
  - add the minimal Phase 45 persistence models for evidence links and optional entity references
- `apps/api/internal/db/db.go`
  - migrate the new Phase 45 models
- `apps/api/internal/handlers/files.go`
  - keep file citation responses aligned with the new unified payload contract
- `apps/api/internal/handlers/artifacts.go`
  - add artifact-section evidence lookup helpers and artifact citation endpoint if needed
- `apps/api/internal/handlers/collaboration.go`
  - reuse message/thread hydration helpers for citation evidence assembly
- `apps/api/internal/handlers/ai.go`
  - reserve a clean integration point so AI execute can consume citation lookup later without reworking payloads
- `apps/api/internal/handlers/collaboration_test.go`
  - add end-to-end API tests for unified citation lookup and entity-aware response fields
- `apps/api/main.go`
  - register new citation lookup routes

### New files to create

- `apps/api/internal/knowledge/types.go`
  - shared types for evidence kinds, locators, citation records, and lookup responses
- `apps/api/internal/knowledge/service.go`
  - orchestration for cross-source citation lookup and normalization
- `apps/api/internal/knowledge/service_test.go`
  - package-level ranking and normalization tests

### Existing files to inspect while implementing

- `apps/api/internal/fileindex/service.go`
  - reuse file chunk access patterns instead of inventing new file parsing work
- `apps/api/internal/handlers/test_helpers_test.go`
  - keep API tests consistent with existing helpers and ID expectations
- `docs/superpowers/specs/2026-04-21-knowledge-entity-wiki-graph-design.md`
  - source of truth for Phase 45 payload reservations and future boundaries
- `docs/AGENT-COLLAB.md`
  - handoff notes for Windsurf after the backend release

## Task 1: Add Phase 45 evidence persistence models

**Files:**
- Modify: `apps/api/internal/domain/models.go`
- Modify: `apps/api/internal/db/db.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write the failing persistence test**

Add a test in `apps/api/internal/handlers/collaboration_test.go` that:
- runs `setupTestDB(t)`
- creates one `domain.KnowledgeEvidenceLink`
- creates one `domain.KnowledgeEvidenceEntityRef`
- verifies both rows persist and can be queried back

Suggested code sketch:

```go
func TestKnowledgeEvidenceModelsPersist(t *testing.T) {
	setupTestDB(t)

	link := domain.KnowledgeEvidenceLink{
		ID:         ids.NewPrefixedUUID("evidence"),
		WorkspaceID: "ws-1",
		EvidenceKind: "file_chunk",
		EvidenceRefID: "chunk-1",
		SourceKind: "file",
		SourceRef: "file-1",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := db.DB.Create(&link).Error; err != nil {
		t.Fatalf("failed to create evidence link: %v", err)
	}

	ref := domain.KnowledgeEvidenceEntityRef{
		ID: ids.NewPrefixedUUID("evidence-ref"),
		EvidenceID: link.ID,
		EntityID: "entity-1",
		CreatedAt: time.Now().UTC(),
	}
	if err := db.DB.Create(&ref).Error; err != nil {
		t.Fatalf("failed to create evidence entity ref: %v", err)
	}
}
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEvidenceModelsPersist$'
```

Expected:
- FAIL because `KnowledgeEvidenceLink` and `KnowledgeEvidenceEntityRef` do not exist yet.

- [ ] **Step 3: Add the new domain models**

Modify `apps/api/internal/domain/models.go` to add:

```go
type KnowledgeEvidenceLink struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	WorkspaceID   string    `gorm:"index" json:"workspace_id"`
	EvidenceKind  string    `json:"evidence_kind"`
	EvidenceRefID string    `gorm:"index" json:"evidence_ref_id"`
	SourceKind    string    `json:"source_kind"`
	SourceRef     string    `json:"source_ref"`
	RefKind       string    `json:"ref_kind"`
	Locator       string    `json:"locator"`
	Snippet       string    `json:"snippet"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type KnowledgeEvidenceEntityRef struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	EvidenceID string    `gorm:"index" json:"evidence_id"`
	EntityID   string    `gorm:"index" json:"entity_id"`
	CreatedAt  time.Time `json:"created_at"`
}
```

These tables are intentionally minimal. They reserve the join points Phase 46 needs without forcing full entity CRUD into Phase 45.

- [ ] **Step 4: Migrate the new models**

Modify `apps/api/internal/db/db.go` and `setupTestDB` so `AutoMigrate` includes:

```go
&domain.KnowledgeEvidenceLink{},
&domain.KnowledgeEvidenceEntityRef{},
```

- [ ] **Step 5: Run the focused test and verify it passes**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestKnowledgeEvidenceModelsPersist$'
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add apps/api/internal/domain/models.go apps/api/internal/db/db.go apps/api/internal/handlers/collaboration_test.go
git commit -m "feat: add knowledge evidence persistence models"
```

## Task 2: Introduce the knowledge citation service

**Files:**
- Create: `apps/api/internal/knowledge/types.go`
- Create: `apps/api/internal/knowledge/service.go`
- Create: `apps/api/internal/knowledge/service_test.go`

- [ ] **Step 1: Write the failing package tests**

Create `apps/api/internal/knowledge/service_test.go` with tests for:

- file chunk citation normalization
- message citation normalization
- artifact section citation normalization
- optional entity binding passthrough
- mixed evidence ranking order

Suggested test names:

- `TestNormalizeFileChunkCitation`
- `TestNormalizeMessageCitation`
- `TestNormalizeArtifactSectionCitation`
- `TestCitationKeepsOptionalEntityFields`
- `TestLookupRanksExactSnippetMatchesAheadOfLooseMatches`

- [ ] **Step 2: Run package tests and verify they fail**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge
```

Expected:
- FAIL because the package does not exist yet.

- [ ] **Step 3: Add shared citation types**

Create `apps/api/internal/knowledge/types.go` with:

```go
package knowledge

type Citation struct {
	ID          string  `json:"id"`
	EvidenceKind string `json:"evidence_kind"`
	SourceKind  string  `json:"source_kind"`
	SourceRef   string  `json:"source_ref"`
	RefKind     string  `json:"ref_kind"`
	Locator     string  `json:"locator,omitempty"`
	Snippet     string  `json:"snippet"`
	Title       string  `json:"title,omitempty"`
	Score       float64 `json:"score"`
	EntityID    string  `json:"entity_id,omitempty"`
	EntityTitle string  `json:"entity_title,omitempty"`
}

type LookupParams struct {
	Query       string
	ChannelID   string
	EntityID    string
	Limit       int
	IncludeKinds map[string]bool
}
```

- [ ] **Step 4: Implement the minimal lookup service**

Create `apps/api/internal/knowledge/service.go` with:
- helpers to read file chunks from `domain.FileExtractionChunk`
- helpers to read messages and thread replies from `domain.Message`
- helpers to read artifact content from `domain.Artifact`
- a normalizer that emits one `Citation` shape for all sources
- a simple ranking strategy:
  - exact substring match first
  - prefix match second
  - fallback loose `LIKE` matches last

Keep it deterministic and SQLite-friendly. Do not add embeddings or semantic ranking here.

- [ ] **Step 5: Run the package tests and verify they pass**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add apps/api/internal/knowledge/types.go apps/api/internal/knowledge/service.go apps/api/internal/knowledge/service_test.go
git commit -m "feat: add unified knowledge citation service"
```

## Task 3: Add the unified citation lookup API

**Files:**
- Modify: `apps/api/internal/handlers/files.go`
- Modify: `apps/api/internal/handlers/artifacts.go`
- Modify: `apps/api/internal/handlers/collaboration.go`
- Modify: `apps/api/main.go`
- Test: `apps/api/internal/handlers/collaboration_test.go`

- [ ] **Step 1: Write the failing API tests**

Add handler tests in `apps/api/internal/handlers/collaboration_test.go` for:

- `GET /api/v1/citations/lookup?q=launch`
- channel-scoped lookups
- artifact evidence in results
- message evidence in results
- file chunk evidence in results
- optional `entity_id` field presence when evidence is linked

Suggested test names:

- `TestCitationLookupReturnsMixedEvidenceKinds`
- `TestCitationLookupHonorsChannelScope`
- `TestCitationLookupIncludesOptionalEntityFields`

- [ ] **Step 2: Run the focused API tests and verify they fail**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestCitationLookup'
```

Expected:
- FAIL because the route and handler do not exist yet.

- [ ] **Step 3: Implement the new lookup handler**

Add a new handler, preferably in `apps/api/internal/handlers/files.go` or a small dedicated section near other knowledge endpoints:

```go
func LookupCitations(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	params := knowledge.LookupParams{
		Query:     q,
		ChannelID: strings.TrimSpace(c.Query("channel_id")),
		EntityID:  strings.TrimSpace(c.Query("entity_id")),
		Limit:     parsePositiveInt(c.DefaultQuery("limit", "20"), 20),
	}

	results, err := knowledge.Lookup(db.DB, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup citations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"citations": results})
}
```

Keep the response contract stable:
- `id`
- `evidence_kind`
- `source_kind`
- `source_ref`
- `ref_kind`
- `locator`
- `snippet`
- `title`
- `score`
- `entity_id`
- `entity_title`

- [ ] **Step 4: Register the route**

Modify `apps/api/main.go` to add:

```go
v1.GET("/citations/lookup", handlers.LookupCitations)
```

- [ ] **Step 5: Keep file citations aligned with the unified contract**

Update `GetFileCitations` in `apps/api/internal/handlers/files.go` so the file-specific endpoint returns the same field names where possible. This avoids Windsurf building two incompatible citation card renderers.

- [ ] **Step 6: Run the focused API tests and verify they pass**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestCitationLookup'
```

Expected:
- PASS

- [ ] **Step 7: Commit**

```bash
git add apps/api/internal/handlers/files.go apps/api/internal/handlers/artifacts.go apps/api/internal/handlers/collaboration.go apps/api/main.go apps/api/internal/handlers/collaboration_test.go
git commit -m "feat: add unified citation lookup api"
```

## Task 4: Add entity-aware evidence joins

**Files:**
- Modify: `apps/api/internal/knowledge/service.go`
- Modify: `apps/api/internal/handlers/collaboration_test.go`
- Test: `apps/api/internal/knowledge/service_test.go`

- [ ] **Step 1: Write the failing tests for entity-aware evidence**

Add tests that:
- seed a `KnowledgeEvidenceLink`
- seed a `KnowledgeEvidenceEntityRef`
- verify `Lookup` returns `entity_id`
- verify unbound evidence still returns cleanly with empty entity fields

Suggested test names:

- `TestLookupAttachesEntityFieldsWhenEvidenceIsLinked`
- `TestLookupLeavesEntityFieldsEmptyWhenNoLinkExists`

- [ ] **Step 2: Run the focused tests and verify they fail**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge -run 'TestLookup(Attaches|Leaves)'
```

Expected:
- FAIL because entity joins are not wired into lookup yet.

- [ ] **Step 3: Implement the minimal entity join**

Update `apps/api/internal/knowledge/service.go` so normalization:
- looks up matching `KnowledgeEvidenceLink` by `evidence_kind`, `source_kind`, and `source_ref`
- joins optional `KnowledgeEvidenceEntityRef`
- populates `entity_id`

Do not implement full entity title hydration yet unless it is trivial. If needed, leave `entity_title` empty in Phase 45 and document that Phase 46 will backfill it.

- [ ] **Step 4: Run the focused tests and verify they pass**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge -run 'TestLookup(Attaches|Leaves)'
```

Expected:
- PASS

- [ ] **Step 5: Commit**

```bash
git add apps/api/internal/knowledge/service.go apps/api/internal/handlers/collaboration_test.go apps/api/internal/knowledge/service_test.go
git commit -m "feat: attach entity references to citation lookup"
```

## Task 5: Reserve AI execution integration and update docs

**Files:**
- Modify: `apps/api/internal/handlers/ai.go`
- Modify: `docs/AGENT-COLLAB.md`
- Modify: `CHANGELOG.md`
- Create: `docs/releases/v0.5.xx.md`

- [ ] **Step 1: Write the failing AI handler contract test**

Add or extend a test in `apps/api/internal/handlers/ai_test.go` that verifies the AI execution layer can accept a future `citations` field in request payloads without rejecting the request shape.

Suggested code sketch:

```go
func TestExecuteAIAllowsCitationPayloadField(t *testing.T) {
	// bind request with citations field and assert handler does not 400 on shape alone
}
```

- [ ] **Step 2: Run the focused AI test and verify it fails**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestExecuteAIAllowsCitationPayloadField$'
```

Expected:
- FAIL because the request binding does not include the field yet.

- [ ] **Step 3: Add the non-breaking request field**

Modify `apps/api/internal/handlers/ai.go` request binding struct to accept:

```go
Citations []map[string]any `json:"citations"`
```

Do not wire execution behavior in this phase. This is only a compatibility seam.

- [ ] **Step 4: Update collaboration and release docs**

After implementation, update:
- `docs/AGENT-COLLAB.md`
  - note the new citation lookup contract and what Windsurf should consume
- `CHANGELOG.md`
  - document the new endpoint and payload fields
- `docs/releases/v0.5.xx.md`
  - summarize Phase 45 backend delivery and frontend ask

- [ ] **Step 5: Run the focused AI test and verify it passes**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/handlers -run 'TestExecuteAIAllowsCitationPayloadField$'
```

Expected:
- PASS

- [ ] **Step 6: Commit**

```bash
git add apps/api/internal/handlers/ai.go apps/api/internal/handlers/ai_test.go docs/AGENT-COLLAB.md CHANGELOG.md docs/releases/v0.5.xx.md
git commit -m "docs: document phase 45 citation lookup contract"
```

## Task 6: Full verification

**Files:**
- Verify only

- [ ] **Step 1: Run package tests**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./internal/knowledge ./internal/handlers
```

Expected:
- PASS

- [ ] **Step 2: Run full backend tests**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go test ./...
```

Expected:
- PASS

- [ ] **Step 3: Run backend build**

Run:

```bash
cd apps/api && GOCACHE=$(pwd)/.cache/go-build go build ./...
```

Expected:
- PASS

- [ ] **Step 4: Commit verification-only metadata if docs changed**

```bash
git status
```

Expected:
- clean working tree or only intentional release metadata changes

## Notes for Execution

- Keep Phase 45 intentionally narrow. Do not pull full `knowledge_entities` CRUD into this implementation.
- Prefer reusing existing file, message, thread, and artifact query helpers rather than building parallel hydration paths.
- Keep citation ranking simple and deterministic.
- If an entity title is not available yet, return `entity_id` and leave `entity_title` empty instead of inventing placeholder joins.
- Preserve the existing `GET /api/v1/files/:id/citations` endpoint, but normalize its payload toward the new shared contract.
- Avoid schema churn beyond what Phase 45 needs. Phase 46 will carry the heavier entity model.
