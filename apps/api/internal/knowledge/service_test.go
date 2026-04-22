package knowledge

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
)

func TestNormalizeFileChunkCitation(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	extractionID := ids.NewPrefixedUUID("fextract")
	database.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch.md", StoragePath: "launch.md", ContentType: "text/markdown", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.FileExtraction{ID: extractionID, FileID: "file-1", Status: "ready", Extractor: "plain_text", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.FileExtractionChunk{ID: 1, ExtractionID: extractionID, FileID: "file-1", ChunkIndex: 0, Text: "Launch checklist and rollout plan", LocatorType: "document", LocatorValue: "root", Heading: "Launch", CreatedAt: now})

	results, err := Lookup(database, LookupParams{Query: "launch"})
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(results) == 0 || results[0].EvidenceKind != "file_chunk" {
		t.Fatalf("expected file chunk citation, got %#v", results)
	}
}

func TestLookupAttachesEntityFieldsWhenEvidenceIsLinked(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	extractionID := ids.NewPrefixedUUID("fextract")
	database.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch.md", StoragePath: "launch.md", ContentType: "text/markdown", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.FileExtraction{ID: extractionID, FileID: "file-1", Status: "ready", Extractor: "plain_text", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.FileExtractionChunk{ID: 1, ExtractionID: extractionID, FileID: "file-1", ChunkIndex: 0, Text: "Launch checklist and rollout plan", LocatorType: "document", LocatorValue: "root", Heading: "Launch", CreatedAt: now})
	linkID := ids.NewPrefixedUUID("evidence")
	database.Create(&domain.KnowledgeEvidenceLink{ID: linkID, WorkspaceID: "ws-1", EvidenceKind: "file_chunk", EvidenceRefID: "1", SourceKind: "file", SourceRef: "file-1", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEvidenceEntityRef{ID: ids.NewPrefixedUUID("evidence-ref"), EvidenceID: linkID, EntityID: "entity-1", CreatedAt: now})

	results, err := Lookup(database, LookupParams{Query: "launch"})
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(results) == 0 || results[0].EntityID != "entity-1" || results[0].EntityTitle != "Launch Program" {
		t.Fatalf("expected entity binding, got %#v", results)
	}
}

func TestLookupHydratesEntityFromKnowledgeEntityRef(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program needs review", CreatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now})

	results, err := Lookup(database, LookupParams{Query: "launch"})
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(results) == 0 || results[0].EntityID != "entity-1" || results[0].EntityTitle != "Launch Program" {
		t.Fatalf("expected entity binding from knowledge entity ref, got %#v", results)
	}
}

func TestCreateKnowledgeEntityDefaults(t *testing.T) {
	database := setupKnowledgeTestDB(t)

	entity, err := CreateEntity(database, CreateEntityInput{
		WorkspaceID: "ws-1",
		Kind:        "project",
		Title:       "Launch Program",
	})
	if err != nil {
		t.Fatalf("create entity failed: %v", err)
	}
	if entity.ID == "" || entity.Status != "active" || entity.SourceKind != "manual" {
		t.Fatalf("expected defaulted entity, got %#v", entity)
	}
}

func TestAddEntityRefAppendsTimelineEvent(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()
	entity := domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now}
	database.Create(&entity)

	ref, err := AddEntityRef(database, "entity-1", AddEntityRefInput{RefKind: "file", RefID: "file-1", Role: "evidence"})
	if err != nil {
		t.Fatalf("add ref failed: %v", err)
	}
	if ref.ID == "" || ref.EntityID != "entity-1" || ref.Role != "evidence" {
		t.Fatalf("unexpected ref: %#v", ref)
	}

	events, err := ListEntityTimeline(database, "entity-1")
	if err != nil {
		t.Fatalf("timeline failed: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "file_linked" {
		t.Fatalf("expected file_linked event, got %#v", events)
	}
}

func TestEntityGraphPreviewIncludesRefsAndLinks(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now})
	database.Create(&domain.KnowledgeEntityLink{ID: "klink-1", WorkspaceID: "ws-1", FromEntityID: "entity-1", ToEntityID: "entity-2", Relation: "depends_on", CreatedAt: now})

	graph, err := BuildEntityGraph(database, "entity-1")
	if err != nil {
		t.Fatalf("graph failed: %v", err)
	}
	if len(graph.Nodes) < 3 || len(graph.Edges) < 2 {
		t.Fatalf("expected entity/ref graph nodes and edges, got %#v", graph)
	}
}

func setupKnowledgeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	if err := database.AutoMigrate(
		&domain.FileAsset{},
		&domain.FileExtraction{},
		&domain.FileExtractionChunk{},
		&domain.Message{},
		&domain.Artifact{},
		&domain.KnowledgeEvidenceLink{},
		&domain.KnowledgeEvidenceEntityRef{},
		&domain.KnowledgeEntity{},
		&domain.KnowledgeEntityRef{},
		&domain.KnowledgeEntityLink{},
		&domain.KnowledgeEvent{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return database
}
