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
	if summary.Velocity.RecentWindowDays == 0 || summary.Velocity.RecentRefCount == 0 || summary.Velocity.Delta == 0 {
		t.Fatalf("expected summary velocity to be populated, got %#v", summary.Velocity)
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

func TestFindMentionedEntitiesResolvesExplicitEntityMentions(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "doc", Title: "Launch", Status: "active", SourceKind: "manual", CreatedAt: now, UpdatedAt: now})

	mentions, err := FindMentionedEntities(database, "ws-1", "<p>@Launch Program is ready, but Launch without @ should not count.</p>")
	if err != nil {
		t.Fatalf("find mentioned entities failed: %v", err)
	}
	if len(mentions) != 1 {
		t.Fatalf("expected one explicit mention, got %#v", mentions)
	}
	if mentions[0].EntityID != "entity-1" || mentions[0].MentionText != "@Launch Program" {
		t.Fatalf("expected longest explicit title match, got %#v", mentions[0])
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

func TestFollowEntityAndListFollowedEntities(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()
	database.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", CreatedAt: now, UpdatedAt: now})

	follow, err := FollowEntity(database, "entity-1", "user-1")
	if err != nil {
		t.Fatalf("failed to follow entity: %v", err)
	}
	if follow.EntityID != "entity-1" || follow.UserID != "user-1" || follow.WorkspaceID != "ws-1" {
		t.Fatalf("unexpected follow payload: %#v", follow)
	}

	if _, err := FollowEntity(database, "entity-1", "user-1"); err != nil {
		t.Fatalf("idempotent follow failed: %v", err)
	}

	items, err := ListFollowedEntities(database, "user-1")
	if err != nil {
		t.Fatalf("failed to list followed entities: %v", err)
	}
	if len(items) != 1 || items[0].Entity.ID != "entity-1" || !items[0].IsFollowing {
		t.Fatalf("unexpected followed entities: %#v", items)
	}

	if err := UnfollowEntity(database, "entity-1", "user-1"); err != nil {
		t.Fatalf("failed to unfollow entity: %v", err)
	}
	items, err = ListFollowedEntities(database, "user-1")
	if err != nil {
		t.Fatalf("failed to list after unfollow: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no followed entities after unfollow, got %#v", items)
	}
}

func TestUpdateFollowNotificationLevelAndDetectSpikeAlerts(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()
	database.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	database.Create(&domain.User{ID: "user-2", Name: "Jane Smith", Email: "jane@example.com"})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program is accelerating", CreatedAt: now.Add(-2 * time.Hour)})
	database.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program needs QA", CreatedAt: now.Add(-90 * time.Minute)})
	database.Create(&domain.Message{ID: "msg-3", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program rollout confirmed", CreatedAt: now.Add(-30 * time.Minute)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-90 * time.Minute)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-3", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-3", Role: "decision", CreatedAt: now.Add(-30 * time.Minute)})

	follow, err := FollowEntity(database, "entity-1", "user-1")
	if err != nil {
		t.Fatalf("failed to follow entity for user-1: %v", err)
	}
	if _, err := FollowEntity(database, "entity-1", "user-2"); err != nil {
		t.Fatalf("failed to follow entity for user-2: %v", err)
	}

	updated, err := UpdateFollowNotificationLevel(database, "entity-1", "user-2", "digest_only")
	if err != nil {
		t.Fatalf("failed to update notification level: %v", err)
	}
	if updated.NotificationLevel != "digest_only" {
		t.Fatalf("expected digest_only level, got %#v", updated)
	}

	alerts, err := DetectEntitySpikeAlerts(database, "entity-1", "ch-1", now)
	if err != nil {
		t.Fatalf("failed to detect spike alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected one alert, got %#v", alerts)
	}
	if alerts[0].Entity.ID != "entity-1" || alerts[0].RecentRefCount != 3 || len(alerts[0].UserIDs) != 1 || alerts[0].UserIDs[0] != "user-1" {
		t.Fatalf("unexpected alert payload: %#v", alerts[0])
	}

	var refreshed domain.KnowledgeEntityFollow
	if err := database.First(&refreshed, "id = ?", follow.ID).Error; err != nil {
		t.Fatalf("failed to reload follow: %v", err)
	}
	if refreshed.LastAlertedAt == nil {
		t.Fatalf("expected last_alerted_at to be populated, got %#v", refreshed)
	}
}

func TestMatchEntitiesInTextReturnsLongestNonOverlappingSpans(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Summary: "Main launch initiative", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "doc", Title: "Launch", Summary: "Generic launch", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-3", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Summary: "Payments", Status: "active", CreatedAt: now, UpdatedAt: now})

	matches, err := MatchEntitiesInText(database, MatchEntitiesInput{
		WorkspaceID: "ws-1",
		Text:        "Can we align Launch Program with Billing Service today?",
		Limit:       8,
	})
	if err != nil {
		t.Fatalf("failed to match entities: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected two entity matches, got %#v", matches)
	}
	if matches[0].EntityID != "entity-1" || matches[0].MatchedText != "Launch Program" {
		t.Fatalf("expected longest Launch Program match first, got %#v", matches[0])
	}
	if matches[1].EntityID != "entity-3" || matches[1].MatchedText != "Billing Service" {
		t.Fatalf("expected billing match second, got %#v", matches[1])
	}
	if matches[0].Start >= matches[0].End || matches[1].Start >= matches[1].End {
		t.Fatalf("expected valid spans, got %#v", matches)
	}
}

func TestGetEntityHoverSummaryAggregatesLiveRefStats(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	database.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "exec", Type: "public"})
	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-12 * time.Hour)})
	database.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-2", UserID: "user-1", Content: "Launch Program escalated", CreatedAt: now.Add(-2 * time.Hour)})
	database.Create(&domain.FileAsset{ID: "file-1", ChannelID: "ch-1", UploaderID: "user-1", Name: "launch-plan.md", StoragePath: "launch-plan.md", ContentType: "text/markdown", CreatedAt: now.Add(-6 * time.Hour), UpdatedAt: now.Add(-6 * time.Hour)})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-12 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "file", RefID: "file-1", Role: "evidence", CreatedAt: now.Add(-6 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-3", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})

	hover, err := GetEntityHoverSummary(database, "entity-1", "ch-1", 7)
	if err != nil {
		t.Fatalf("hover summary failed: %v", err)
	}
	if hover.EntityID != "entity-1" || hover.RefCount != 3 || hover.ChannelRefCount != 2 {
		t.Fatalf("unexpected hover envelope: %#v", hover)
	}
	if hover.MessageRefCount != 2 || hover.FileRefCount != 1 || hover.RecentRefCount != 3 {
		t.Fatalf("expected split ref counts, got %#v", hover)
	}
	if len(hover.RelatedChannels) != 2 || hover.RelatedChannels[0].ChannelID != "ch-1" {
		t.Fatalf("expected related channels ranked by ref count, got %#v", hover.RelatedChannels)
	}
}

func TestBuildChannelKnowledgeDigestRanksRecentMovements(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-48 * time.Hour)})
	database.Create(&domain.Message{ID: "msg-2", ChannelID: "ch-1", UserID: "user-1", Content: "Billing Service dependency", CreatedAt: now.Add(-6 * time.Hour)})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntity{ID: "entity-2", WorkspaceID: "ws-1", Kind: "service", Title: "Billing Service", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-48 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-2", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-6 * time.Hour)})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-3", WorkspaceID: "ws-1", EntityID: "entity-2", RefKind: "message", RefID: "msg-2", Role: "discussion", CreatedAt: now.Add(-2 * time.Hour)})

	digest, err := BuildChannelKnowledgeDigest(database, "ch-1", "weekly", 5)
	if err != nil {
		t.Fatalf("build digest failed: %v", err)
	}
	if digest.ChannelID != "ch-1" || digest.Window != "weekly" || digest.WindowDays != 7 {
		t.Fatalf("unexpected digest envelope: %#v", digest)
	}
	if digest.TotalRefs != 3 || len(digest.TopMovements) != 2 {
		t.Fatalf("expected digest movement list, got %#v", digest)
	}
	if digest.TopMovements[0].EntityID != "entity-2" || digest.TopMovements[0].RecentRefCount != 2 {
		t.Fatalf("expected hottest recent entity first, got %#v", digest.TopMovements)
	}
	if digest.Headline == "" || digest.Summary == "" {
		t.Fatalf("expected digest narrative text, got %#v", digest)
	}
}

func TestUpsertDigestScheduleAndComputeNextRunAt(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Date(2026, 4, 22, 8, 30, 0, 0, time.UTC)

	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})

	schedule, err := UpsertDigestSchedule(database, UpsertDigestScheduleInput{
		ChannelID: "ch-1",
		CreatedBy: "user-1",
		Window:    "weekly",
		Timezone:  "Asia/Shanghai",
		DayOfWeek: 0,
		Hour:      9,
		Minute:    0,
		Limit:     5,
		Pin:       true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("upsert digest schedule failed: %v", err)
	}
	if schedule.ChannelID != "ch-1" || schedule.Window != "weekly" || !schedule.Pin {
		t.Fatalf("unexpected stored schedule: %#v", schedule)
	}

	nextRunAt, err := ComputeDigestScheduleNextRunAt(schedule, now)
	if err != nil {
		t.Fatalf("compute next run failed: %v", err)
	}
	if nextRunAt == nil || nextRunAt.Location().String() != "Asia/Shanghai" {
		t.Fatalf("expected localized next run time, got %#v", nextRunAt)
	}
	if nextRunAt.Weekday() != time.Sunday || nextRunAt.Hour() != 9 || nextRunAt.Minute() != 0 {
		t.Fatalf("expected next run on Sunday 09:00 Asia/Shanghai, got %v", nextRunAt)
	}
}

func TestProcessDigestSchedulesPublishesDueDigestOnce(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Date(2026, 4, 26, 1, 5, 0, 0, time.UTC)

	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public"})
	database.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	database.Create(&domain.Message{ID: "msg-1", ChannelID: "ch-1", UserID: "user-1", Content: "Launch Program kickoff", CreatedAt: now.Add(-2 * time.Hour)})
	database.Create(&domain.KnowledgeEntity{ID: "entity-1", WorkspaceID: "ws-1", Kind: "project", Title: "Launch Program", Status: "active", CreatedAt: now, UpdatedAt: now})
	database.Create(&domain.KnowledgeEntityRef{ID: "kref-1", WorkspaceID: "ws-1", EntityID: "entity-1", RefKind: "message", RefID: "msg-1", Role: "discussion", CreatedAt: now.Add(-90 * time.Minute)})

	_, err := UpsertDigestSchedule(database, UpsertDigestScheduleInput{
		ChannelID: "ch-1",
		CreatedBy: "user-1",
		Window:    "weekly",
		Timezone:  "Asia/Shanghai",
		DayOfWeek: 0,
		Hour:      9,
		Minute:    0,
		Limit:     5,
		Pin:       true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("upsert digest schedule failed: %v", err)
	}

	published, err := ProcessDigestSchedules(database, now)
	if err != nil {
		t.Fatalf("process digest schedules failed: %v", err)
	}
	if len(published) != 1 || published[0].Message.ChannelID != "ch-1" || !published[0].Message.IsPinned {
		t.Fatalf("expected one pinned digest publish, got %#v", published)
	}

	var count int64
	database.Model(&domain.Message{}).Where("channel_id = ? AND is_pinned = ?", "ch-1", true).Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly one published digest message, got %d", count)
	}

	published, err = ProcessDigestSchedules(database, now.Add(10*time.Minute))
	if err != nil {
		t.Fatalf("re-process digest schedules failed: %v", err)
	}
	if len(published) != 0 {
		t.Fatalf("expected second pass in same slot to skip duplicate publish, got %#v", published)
	}
}

func TestListKnowledgeInboxReturnsDigestMessagesAndReadState(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	now := time.Now().UTC()

	database.Create(&domain.User{ID: "user-1", Name: "Nikko Fu", Email: "nikko@example.com"})
	database.Create(&domain.Channel{ID: "ch-1", WorkspaceID: "ws-1", Name: "launch", Type: "public", IsStarred: true})
	database.Create(&domain.Channel{ID: "ch-2", WorkspaceID: "ws-1", Name: "ops", Type: "public"})
	database.Create(&domain.ChannelMember{ChannelID: "ch-1", UserID: "user-1", Role: "member"})
	database.Create(&domain.ChannelMember{ChannelID: "ch-2", UserID: "user-1", Role: "member"})

	firstDigest, err := PublishChannelDigest(database, PublishChannelDigestInput{
		ChannelID:  "ch-1",
		UserID:     "user-1",
		Window:     "weekly",
		Limit:      5,
		Pin:        true,
		OccurredAt: now.Add(-2 * time.Hour),
	})
	if err != nil {
		t.Fatalf("publish digest failed: %v", err)
	}
	secondDigest, err := PublishChannelDigest(database, PublishChannelDigestInput{
		ChannelID:  "ch-2",
		UserID:     "user-1",
		Window:     "daily",
		Limit:      5,
		Pin:        false,
		OccurredAt: now,
	})
	if err != nil {
		t.Fatalf("publish second digest failed: %v", err)
	}

	database.Create(&domain.NotificationRead{UserID: "user-1", ItemID: "knowledge-digest-" + firstDigest.Message.ID, ReadAt: now})

	items, err := ListKnowledgeInbox(database, KnowledgeInboxParams{UserID: "user-1", Scope: "all", Limit: 10})
	if err != nil {
		t.Fatalf("list knowledge inbox failed: %v", err)
	}
	if len(items) != 2 || items[0].Message.ID != secondDigest.Message.ID {
		t.Fatalf("expected newest digest first, got %#v", items)
	}
	if items[0].IsRead || !items[1].IsRead {
		t.Fatalf("expected read state from notification reads, got %#v", items)
	}

	starredOnly, err := ListKnowledgeInbox(database, KnowledgeInboxParams{UserID: "user-1", Scope: "starred", Limit: 10})
	if err != nil {
		t.Fatalf("list starred knowledge inbox failed: %v", err)
	}
	if len(starredOnly) != 1 || starredOnly[0].Channel.ID != "ch-1" {
		t.Fatalf("expected starred scope to keep only starred-channel digest, got %#v", starredOnly)
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
		&domain.KnowledgeEntityFollow{},
		&domain.KnowledgeDigestSchedule{},
		&domain.NotificationRead{},
		&domain.User{},
		&domain.ChannelMember{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return database
}
