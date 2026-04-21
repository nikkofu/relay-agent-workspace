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
	database.Create(&domain.KnowledgeEvidenceEntityRef{ID: ids.NewPrefixedUUID("evidence-ref"), EvidenceID: linkID, EntityID: "entity-1", CreatedAt: now})

	results, err := Lookup(database, LookupParams{Query: "launch"})
	if err != nil {
		t.Fatalf("lookup failed: %v", err)
	}
	if len(results) == 0 || results[0].EntityID != "entity-1" {
		t.Fatalf("expected entity binding, got %#v", results)
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
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return database
}
