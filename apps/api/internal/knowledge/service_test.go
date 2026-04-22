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

func TestGetChannelKnowledgeSummaryAggregatesTopEntitiesAndTrend(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-48 * time.Hour)})
	database.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Billing Service dependency", CreatedAt: now.Add(-24 * time.Hour)})
	database.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now.Add(-12 * time.Hour), UpdatedAt: now.Add(-12 * time.Hour)})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-48 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now.Add(-12 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-3", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-24 * time.Hour)})

	summary, err := GetChannelKnowledgeSummary(database, "ch-1", 5, 7)
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}
	if summary.ChannelID != "ch-1" || summary.TotalRefs != 3 || summary.RecentRefCount != 3 {
		t.Fatalf("unexpected summary envelope: %#v", summary)
	}
	if len(summary.TopEntities) != 2 {
		t.Fatalf("expected two top entities, got %#v", summary.TopEntities)
	}
	if summary.TopEntities[0].EntityID != "entity-1" || summary.TopEntities[0].RefCount != 2 || summary.TopEntities[0].MessageRefCount != 1 || summary.TopEntities[0].FileRefCount != 1 {
		t.Fatalf("expected entity-1 to lead aggregated counts, got %#v", summary.TopEntities[0])
	}
	if len(summary.TopEntities[0].Trend) != 7 {
		t.Fatalf("expected seven daily trend points, got %#v", summary.TopEntities[0].Trend)
	}
}

func TestSuggestEntitiesScopesToChannelWorkspaceAndRanksByChannelRefs(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	database.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-2", Name: "external", Type: "public"})
	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now})
	database.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "doc", Title: "Launch Checklist", Summary: "Pre-flight list", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-3", WorkspaceID: "ws-2", Kind: "project", Title: "Launch External", Summary: "Other workspace", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now})

	suggestions, err := SuggestEntities(database, SuggestEntitiesParams{Query: "launch", ChannelID: "ch-1", Limit: 5})
	if err != nil {
		t.Fatalf("suggest failed: %v", err)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected two scoped suggestions, got %#v", suggestions)
	}
	if suggestions[0].ID != "entity-1" || suggestions[0].ChannelRefCount != 2 {
		t.Fatalf("expected channel-ranked entity first, got %#v", suggestions[0])
	}
	for _, suggestion := range suggestions {
		if suggestion.ID == "entity-3" {
			t.Fatalf("expected workspace scoping to exclude entity-3, got %#v", suggestions)
		}
	}
}

func setupKnowledgeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open("file:"+ids.NewPrefixedUUID("test")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite test db: %v", err)
	}
	if err := database.AutoMigrate(
		&domain.Channel{},
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
