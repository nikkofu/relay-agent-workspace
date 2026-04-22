package knowledge

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
)

func Lookup(database *gorm.DB, params LookupParams) ([]Citation, error) {
	query := strings.TrimSpace(params.Query)
	if query == "" {
		return []Citation{}, nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	citations := make([]Citation, 0, limit)

	fileCitations, err := lookupFileChunkCitations(database, params)
	if err != nil {
		return nil, err
	}
	citations = append(citations, fileCitations...)

	messageCitations, err := lookupMessageCitations(database, params)
	if err != nil {
		return nil, err
	}
	citations = append(citations, messageCitations...)

	artifactCitations, err := lookupArtifactCitations(database, params)
	if err != nil {
		return nil, err
	}
	citations = append(citations, artifactCitations...)

	sort.SliceStable(citations, func(i, j int) bool {
		if citations[i].Score == citations[j].Score {
			return citations[i].ID < citations[j].ID
		}
		return citations[i].Score > citations[j].Score
	})

	if len(citations) > limit {
		citations = citations[:limit]
	}

	return citations, nil
}

func lookupFileChunkCitations(database *gorm.DB, params LookupParams) ([]Citation, error) {
	var chunks []domain.FileExtractionChunk
	query := database.Model(&domain.FileExtractionChunk{}).
		Order("chunk_index asc, id asc")
	if params.ChannelID != "" {
		query = query.Joins("JOIN file_assets ON file_assets.id = file_extraction_chunks.file_id").
			Where("file_assets.channel_id = ?", params.ChannelID)
	}
	if err := query.Find(&chunks).Error; err != nil {
		return nil, err
	}

	results := make([]Citation, 0, len(chunks))
	for _, chunk := range chunks {
		score, ok := scoreText(chunk.Text, params.Query)
		if !ok || !includeKind(params, "file_chunk") {
			continue
		}

		citation := Citation{
			ID:           fmt.Sprintf("file-chunk-%d", chunk.ID),
			EvidenceKind: "file_chunk",
			SourceKind:   "file",
			SourceRef:    chunk.FileID,
			RefKind:      "file_chunk",
			Locator:      buildLocator(chunk.LocatorType, chunk.LocatorValue),
			Snippet:      bestSnippet(chunk.Text, params.Query),
			Title:        chunk.Heading,
			Score:        score,
		}
		if err := attachEntity(database, &citation, "file_chunk", strconv.FormatUint(uint64(chunk.ID), 10), "file", chunk.FileID); err != nil {
			return nil, err
		}
		if params.EntityID != "" && citation.EntityID != params.EntityID {
			continue
		}
		results = append(results, citation)
	}
	return results, nil
}

func lookupMessageCitations(database *gorm.DB, params LookupParams) ([]Citation, error) {
	var messages []domain.Message
	query := database.Model(&domain.Message{}).
		Order("created_at desc")
	if params.ChannelID != "" {
		query = query.Where("channel_id = ?", params.ChannelID)
	}
	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}

	results := make([]Citation, 0, len(messages))
	for _, message := range messages {
		score, ok := scoreText(message.Content, params.Query)
		if !ok {
			continue
		}
		evidenceKind := "message"
		refKind := "message"
		if message.ThreadID != "" {
			evidenceKind = "thread"
			refKind = "thread_message"
		}
		if !includeKind(params, evidenceKind) {
			continue
		}

		citation := Citation{
			ID:           message.ID,
			EvidenceKind: evidenceKind,
			SourceKind:   "message",
			SourceRef:    message.ID,
			RefKind:      refKind,
			Locator:      message.ChannelID,
			Snippet:      bestSnippet(message.Content, params.Query),
			Title:        firstLine(message.Content),
			Score:        score,
		}
		if err := attachEntity(database, &citation, evidenceKind, message.ID, "message", message.ID); err != nil {
			return nil, err
		}
		if params.EntityID != "" && citation.EntityID != params.EntityID {
			continue
		}
		results = append(results, citation)
	}
	return results, nil
}

func lookupArtifactCitations(database *gorm.DB, params LookupParams) ([]Citation, error) {
	var artifacts []domain.Artifact
	query := database.Model(&domain.Artifact{}).
		Order("updated_at desc")
	if params.ChannelID != "" {
		query = query.Where("channel_id = ?", params.ChannelID)
	}
	if err := query.Find(&artifacts).Error; err != nil {
		return nil, err
	}

	results := make([]Citation, 0, len(artifacts))
	for _, artifact := range artifacts {
		sections := splitArtifactSections(artifact.Content)
		for idx, section := range sections {
			score, ok := scoreText(section, params.Query)
			if !ok || !includeKind(params, "artifact_section") {
				continue
			}
			locator := fmt.Sprintf("section:%d", idx+1)
			citation := Citation{
				ID:           fmt.Sprintf("%s#%d", artifact.ID, idx+1),
				EvidenceKind: "artifact_section",
				SourceKind:   "artifact",
				SourceRef:    artifact.ID,
				RefKind:      "artifact_section",
				Locator:      locator,
				Snippet:      bestSnippet(section, params.Query),
				Title:        artifact.Title,
				Score:        score,
			}
			if err := attachEntity(database, &citation, "artifact_section", citation.ID, "artifact", artifact.ID); err != nil {
				return nil, err
			}
			if params.EntityID != "" && citation.EntityID != params.EntityID {
				continue
			}
			results = append(results, citation)
		}
	}
	return results, nil
}

func attachEntity(database *gorm.DB, citation *Citation, evidenceKind, evidenceRefID, sourceKind, sourceRef string) error {
	var link domain.KnowledgeEvidenceLink
	err := database.Where(
		"evidence_kind = ? AND evidence_ref_id = ? AND source_kind = ? AND source_ref = ?",
		evidenceKind,
		evidenceRefID,
		sourceKind,
		sourceRef,
	).First(&link).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return attachEntityFromKnowledgeRef(database, citation, evidenceKind, evidenceRefID, sourceKind, sourceRef)
		}
		return err
	}

	var ref domain.KnowledgeEvidenceEntityRef
	err = database.Where("evidence_id = ?", link.ID).First(&ref).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	citation.EntityID = ref.EntityID
	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", ref.EntityID).Error; err == nil {
		citation.EntityTitle = entity.Title
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func attachEntityFromKnowledgeRef(database *gorm.DB, citation *Citation, evidenceKind, evidenceRefID, sourceKind, sourceRef string) error {
	refKinds := []string{evidenceKind}
	if sourceKind != "" && sourceKind != evidenceKind {
		refKinds = append(refKinds, sourceKind)
	}

	refIDs := []string{evidenceRefID}
	if sourceRef != "" && sourceRef != evidenceRefID {
		refIDs = append(refIDs, sourceRef)
	}

	var ref domain.KnowledgeEntityRef
	err := database.Where("ref_kind IN ? AND ref_id IN ?", refKinds, refIDs).Order("created_at desc").First(&ref).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	citation.EntityID = ref.EntityID
	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", ref.EntityID).Error; err == nil {
		citation.EntityTitle = entity.Title
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	return nil
}

func CreateEntity(database *gorm.DB, input CreateEntityInput) (domain.KnowledgeEntity, error) {
	now := time.Now().UTC()
	entity := domain.KnowledgeEntity{
		ID:          newKnowledgeID("entity"),
		WorkspaceID: strings.TrimSpace(input.WorkspaceID),
		Kind:        defaultString(strings.TrimSpace(input.Kind), "custom"),
		Title:       strings.TrimSpace(input.Title),
		Summary:     strings.TrimSpace(input.Summary),
		Status:      defaultString(strings.TrimSpace(input.Status), "active"),
		OwnerUserID: strings.TrimSpace(input.OwnerUserID),
		SourceKind:  defaultString(strings.TrimSpace(input.SourceKind), "manual"),
		SourceRef:   strings.TrimSpace(input.SourceRef),
		Metadata:    strings.TrimSpace(input.Metadata),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if entity.WorkspaceID == "" || entity.Title == "" {
		return entity, errors.New("workspace_id and title are required")
	}
	if err := database.Create(&entity).Error; err != nil {
		return entity, err
	}
	_, _ = AppendEntityEvent(database, entity.ID, AddEntityEventInput{
		EventType:  "created",
		Title:      "Created " + entity.Title,
		SourceKind: "system",
	})
	return entity, nil
}

func UpdateEntity(database *gorm.DB, id string, input UpdateEntityInput) (domain.KnowledgeEntity, error) {
	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", id).Error; err != nil {
		return entity, err
	}
	if strings.TrimSpace(input.Kind) != "" {
		entity.Kind = strings.TrimSpace(input.Kind)
	}
	if strings.TrimSpace(input.Title) != "" {
		entity.Title = strings.TrimSpace(input.Title)
	}
	if input.Summary != "" {
		entity.Summary = strings.TrimSpace(input.Summary)
	}
	if strings.TrimSpace(input.Status) != "" {
		entity.Status = strings.TrimSpace(input.Status)
	}
	if input.OwnerUserID != "" {
		entity.OwnerUserID = strings.TrimSpace(input.OwnerUserID)
	}
	if input.SourceKind != "" {
		entity.SourceKind = strings.TrimSpace(input.SourceKind)
	}
	if input.SourceRef != "" {
		entity.SourceRef = strings.TrimSpace(input.SourceRef)
	}
	if input.Metadata != "" {
		entity.Metadata = strings.TrimSpace(input.Metadata)
	}
	entity.UpdatedAt = time.Now().UTC()
	if err := database.Save(&entity).Error; err != nil {
		return entity, err
	}
	_, _ = AppendEntityEvent(database, entity.ID, AddEntityEventInput{
		EventType:  "updated",
		Title:      "Updated " + entity.Title,
		SourceKind: "system",
	})
	return entity, nil
}

func ListEntities(database *gorm.DB, workspaceID string) ([]domain.KnowledgeEntity, error) {
	var entities []domain.KnowledgeEntity
	query := database.Order("updated_at desc")
	if strings.TrimSpace(workspaceID) != "" {
		query = query.Where("workspace_id = ?", strings.TrimSpace(workspaceID))
	}
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func GetEntity(database *gorm.DB, id string) (domain.KnowledgeEntity, error) {
	var entity domain.KnowledgeEntity
	err := database.First(&entity, "id = ?", id).Error
	return entity, err
}

func AddEntityRef(database *gorm.DB, entityID string, input AddEntityRefInput) (domain.KnowledgeEntityRef, error) {
	entity, err := GetEntity(database, entityID)
	if err != nil {
		return domain.KnowledgeEntityRef{}, err
	}
	now := time.Now().UTC()
	ref := domain.KnowledgeEntityRef{
		ID:          newKnowledgeID("kref"),
		WorkspaceID: entity.WorkspaceID,
		EntityID:    entity.ID,
		RefKind:     strings.TrimSpace(input.RefKind),
		RefID:       strings.TrimSpace(input.RefID),
		Role:        defaultString(strings.TrimSpace(input.Role), "evidence"),
		Metadata:    strings.TrimSpace(input.Metadata),
		CreatedAt:   now,
	}
	if ref.RefKind == "" || ref.RefID == "" {
		return ref, errors.New("ref_kind and ref_id are required")
	}
	if err := database.Create(&ref).Error; err != nil {
		return ref, err
	}
	_, _ = AppendEntityEvent(database, entity.ID, AddEntityEventInput{
		EventType:  ref.RefKind + "_linked",
		Title:      "Linked " + ref.RefKind,
		SourceKind: "system",
		SourceRef:  ref.RefID,
	})
	return ref, nil
}

func ListEntityRefs(database *gorm.DB, entityID string) ([]domain.KnowledgeEntityRef, error) {
	var refs []domain.KnowledgeEntityRef
	if err := database.Where("entity_id = ?", entityID).Order("created_at desc").Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

func AddEntityLink(database *gorm.DB, input AddEntityLinkInput) (domain.KnowledgeEntityLink, error) {
	from, err := GetEntity(database, input.FromEntityID)
	if err != nil {
		return domain.KnowledgeEntityLink{}, err
	}
	if _, err := GetEntity(database, input.ToEntityID); err != nil {
		return domain.KnowledgeEntityLink{}, err
	}
	link := domain.KnowledgeEntityLink{
		ID:           newKnowledgeID("klink"),
		WorkspaceID:  defaultString(strings.TrimSpace(input.WorkspaceID), from.WorkspaceID),
		FromEntityID: strings.TrimSpace(input.FromEntityID),
		ToEntityID:   strings.TrimSpace(input.ToEntityID),
		Relation:     defaultString(strings.TrimSpace(input.Relation), "relates_to"),
		Weight:       input.Weight,
		Metadata:     strings.TrimSpace(input.Metadata),
		CreatedAt:    time.Now().UTC(),
	}
	if link.Weight == 0 {
		link.Weight = 1
	}
	if err := database.Create(&link).Error; err != nil {
		return link, err
	}
	return link, nil
}

func ListEntityLinks(database *gorm.DB, entityID string) ([]domain.KnowledgeEntityLink, error) {
	var links []domain.KnowledgeEntityLink
	if err := database.Where("from_entity_id = ? OR to_entity_id = ?", entityID, entityID).Order("created_at desc").Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func AppendEntityEvent(database *gorm.DB, entityID string, input AddEntityEventInput) (domain.KnowledgeEvent, error) {
	entity, err := GetEntity(database, entityID)
	if err != nil {
		return domain.KnowledgeEvent{}, err
	}
	now := time.Now().UTC()
	event := domain.KnowledgeEvent{
		ID:          newKnowledgeID("kevent"),
		WorkspaceID: entity.WorkspaceID,
		EntityID:    entity.ID,
		EventType:   defaultString(strings.TrimSpace(input.EventType), "updated"),
		Title:       strings.TrimSpace(input.Title),
		Body:        strings.TrimSpace(input.Body),
		ActorUserID: strings.TrimSpace(input.ActorUserID),
		SourceKind:  defaultString(strings.TrimSpace(input.SourceKind), "system"),
		SourceRef:   strings.TrimSpace(input.SourceRef),
		OccurredAt:  now,
		CreatedAt:   now,
	}
	if event.Title == "" {
		event.Title = event.EventType
	}
	if err := database.Create(&event).Error; err != nil {
		return event, err
	}
	return event, nil
}

func IngestEvent(database *gorm.DB, input IngestEventInput) (domain.KnowledgeEvent, error) {
	return AppendEntityEvent(database, strings.TrimSpace(input.EntityID), AddEntityEventInput{
		EventType:   input.EventType,
		Title:       input.Title,
		Body:        input.Body,
		ActorUserID: input.ActorUserID,
		SourceKind:  defaultString(strings.TrimSpace(input.SourceKind), "live"),
		SourceRef:   input.SourceRef,
	})
}

func AutoLinkEntitiesForMessage(database *gorm.DB, workspaceID, messageID, content string) ([]domain.KnowledgeEntityRef, error) {
	return autoLinkEntities(database, workspaceID, "message", messageID, "discussion", content)
}

func AutoLinkEntitiesForFile(database *gorm.DB, workspaceID, fileID, content string) ([]domain.KnowledgeEntityRef, error) {
	return autoLinkEntities(database, workspaceID, "file", fileID, "evidence", content)
}

func autoLinkEntities(database *gorm.DB, workspaceID, refKind, refID, role, content string) ([]domain.KnowledgeEntityRef, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	refID = strings.TrimSpace(refID)
	if workspaceID == "" || refID == "" || strings.TrimSpace(content) == "" {
		return []domain.KnowledgeEntityRef{}, nil
	}

	haystack := normalizeKnowledgeText(content)
	if haystack == "" {
		return []domain.KnowledgeEntityRef{}, nil
	}

	var entities []domain.KnowledgeEntity
	if err := database.Where("workspace_id = ? AND status <> ?", workspaceID, "archived").Order("title asc").Find(&entities).Error; err != nil {
		return nil, err
	}

	created := make([]domain.KnowledgeEntityRef, 0)
	for _, entity := range entities {
		needle := normalizeKnowledgeText(entity.Title)
		if len(needle) < 3 || !strings.Contains(haystack, needle) {
			continue
		}

		var count int64
		if err := database.Model(&domain.KnowledgeEntityRef{}).
			Where("entity_id = ? AND ref_kind = ? AND ref_id = ?", entity.ID, refKind, refID).
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			continue
		}

		ref, err := AddEntityRef(database, entity.ID, AddEntityRefInput{
			RefKind: refKind,
			RefID:   refID,
			Role:    role,
		})
		if err != nil {
			return nil, err
		}
		created = append(created, ref)
	}

	return created, nil
}

func ListEntityTimeline(database *gorm.DB, entityID string) ([]domain.KnowledgeEvent, error) {
	var events []domain.KnowledgeEvent
	if err := database.Where("entity_id = ?", entityID).Order("occurred_at desc, created_at desc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func GetChannelKnowledgeContext(database *gorm.DB, channelID string, limit int) (ChannelKnowledgeContext, error) {
	channelID = strings.TrimSpace(channelID)
	context := ChannelKnowledgeContext{ChannelID: channelID, Refs: []ChannelKnowledgeRef{}}
	if channelID == "" {
		return context, errors.New("channel_id is required")
	}
	if limit <= 0 {
		limit = 20
	}

	messageRefs, err := listChannelMessageKnowledgeRefs(database, channelID)
	if err != nil {
		return context, err
	}
	context.Refs = append(context.Refs, messageRefs...)

	fileRefs, err := listChannelFileKnowledgeRefs(database, channelID)
	if err != nil {
		return context, err
	}
	context.Refs = append(context.Refs, fileRefs...)

	sort.SliceStable(context.Refs, func(i, j int) bool {
		if context.Refs[i].CreatedAt.Equal(context.Refs[j].CreatedAt) {
			return context.Refs[i].ID < context.Refs[j].ID
		}
		return context.Refs[i].CreatedAt.After(context.Refs[j].CreatedAt)
	})
	if len(context.Refs) > limit {
		context.Refs = context.Refs[:limit]
	}

	return context, nil
}

func GetChannelKnowledgeSummary(database *gorm.DB, channelID string, limit int, windowDays int) (ChannelKnowledgeSummary, error) {
	channelID = strings.TrimSpace(channelID)
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	if windowDays <= 0 {
		windowDays = 7
	}
	if windowDays > 30 {
		windowDays = 30
	}

	summary := ChannelKnowledgeSummary{
		ChannelID:   channelID,
		WindowDays:  windowDays,
		TopEntities: []ChannelKnowledgeSummaryEntity{},
	}
	if channelID == "" {
		return summary, errors.New("channel_id is required")
	}

	refs, err := listChannelKnowledgeEntityRefs(database, channelID)
	if err != nil {
		return summary, err
	}
	summary.TotalRefs = len(refs)
	if len(refs) == 0 {
		return summary, nil
	}

	windowStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(windowDays - 1))
	type aggregate struct {
		EntityID        string
		RefCount        int
		MessageRefCount int
		FileRefCount    int
		LastRefAt       time.Time
		TrendCounts     map[string]int
	}

	aggregates := make(map[string]*aggregate, len(refs))
	entityIDs := make([]string, 0)
	for _, ref := range refs {
		item, exists := aggregates[ref.EntityID]
		if !exists {
			item = &aggregate{
				EntityID:    ref.EntityID,
				TrendCounts: map[string]int{},
			}
			aggregates[ref.EntityID] = item
			entityIDs = append(entityIDs, ref.EntityID)
		}
		item.RefCount++
		switch ref.RefKind {
		case "file":
			item.FileRefCount++
		default:
			item.MessageRefCount++
		}
		if ref.CreatedAt.After(item.LastRefAt) {
			item.LastRefAt = ref.CreatedAt
		}
		if !ref.CreatedAt.Before(windowStart) {
			summary.RecentRefCount++
			item.TrendCounts[ref.CreatedAt.UTC().Format("2006-01-02")]++
		}
	}

	entityMap, err := loadKnowledgeEntityMap(database, entityIDs)
	if err != nil {
		return summary, err
	}

	for _, entityID := range entityIDs {
		aggregate := aggregates[entityID]
		entity := entityMap[entityID]
		topEntity := ChannelKnowledgeSummaryEntity{
			EntityID:        entityID,
			EntityTitle:     entity.Title,
			EntityKind:      entity.Kind,
			RefCount:        aggregate.RefCount,
			MessageRefCount: aggregate.MessageRefCount,
			FileRefCount:    aggregate.FileRefCount,
			LastRefAt:       aggregate.LastRefAt,
			Trend:           make([]ChannelKnowledgeTrendPoint, 0, windowDays),
		}
		if strings.TrimSpace(topEntity.EntityTitle) == "" {
			topEntity.EntityTitle = entityID
		}
		for day := 0; day < windowDays; day++ {
			date := windowStart.AddDate(0, 0, day).Format("2006-01-02")
			topEntity.Trend = append(topEntity.Trend, ChannelKnowledgeTrendPoint{
				Date:  date,
				Count: aggregate.TrendCounts[date],
			})
		}
		summary.TopEntities = append(summary.TopEntities, topEntity)
	}

	sort.SliceStable(summary.TopEntities, func(i, j int) bool {
		if summary.TopEntities[i].RefCount == summary.TopEntities[j].RefCount {
			if summary.TopEntities[i].LastRefAt.Equal(summary.TopEntities[j].LastRefAt) {
				return summary.TopEntities[i].EntityTitle < summary.TopEntities[j].EntityTitle
			}
			return summary.TopEntities[i].LastRefAt.After(summary.TopEntities[j].LastRefAt)
		}
		return summary.TopEntities[i].RefCount > summary.TopEntities[j].RefCount
	})
	if len(summary.TopEntities) > limit {
		summary.TopEntities = summary.TopEntities[:limit]
	}

	return summary, nil
}

func SuggestEntities(database *gorm.DB, params SuggestEntitiesParams) ([]KnowledgeEntitySuggestion, error) {
	query := strings.TrimSpace(params.Query)
	if query == "" {
		return []KnowledgeEntitySuggestion{}, nil
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}

	workspaceID := strings.TrimSpace(params.WorkspaceID)
	channelID := strings.TrimSpace(params.ChannelID)
	if channelID != "" && workspaceID == "" {
		var channel domain.Channel
		if err := database.Select("id, workspace_id").First(&channel, "id = ?", channelID).Error; err != nil {
			return nil, err
		}
		workspaceID = channel.WorkspaceID
	}

	needle := "%" + strings.ToLower(query) + "%"
	entityQuery := database.Model(&domain.KnowledgeEntity{})
	if workspaceID != "" {
		entityQuery = entityQuery.Where("workspace_id = ?", workspaceID)
	}

	var entities []domain.KnowledgeEntity
	if err := entityQuery.
		Where("LOWER(title) LIKE ? OR LOWER(summary) LIKE ?", needle, needle).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return []KnowledgeEntitySuggestion{}, nil
	}

	entityIDs := make([]string, 0, len(entities))
	for _, entity := range entities {
		entityIDs = append(entityIDs, entity.ID)
	}

	refCounts, err := countKnowledgeEntityRefs(database, entityIDs)
	if err != nil {
		return nil, err
	}

	channelRefCounts := map[string]int{}
	if channelID != "" {
		channelRefs, err := listChannelKnowledgeEntityRefs(database, channelID)
		if err != nil {
			return nil, err
		}
		for _, ref := range channelRefs {
			channelRefCounts[ref.EntityID]++
		}
	}

	type rankedSuggestion struct {
		KnowledgeEntitySuggestion
		matchScore int
	}

	ranked := make([]rankedSuggestion, 0, len(entities))
	lowerQuery := strings.ToLower(query)
	for _, entity := range entities {
		matchScore := knowledgeEntityMatchScore(entity, lowerQuery)
		if matchScore == 0 {
			continue
		}
		ranked = append(ranked, rankedSuggestion{
			KnowledgeEntitySuggestion: KnowledgeEntitySuggestion{
				ID:              entity.ID,
				Title:           entity.Title,
				Kind:            entity.Kind,
				Summary:         entity.Summary,
				SourceKind:      entity.SourceKind,
				RefCount:        refCounts[entity.ID],
				ChannelRefCount: channelRefCounts[entity.ID],
			},
			matchScore: matchScore,
		})
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].matchScore == ranked[j].matchScore {
			if ranked[i].ChannelRefCount == ranked[j].ChannelRefCount {
				if ranked[i].RefCount == ranked[j].RefCount {
					return ranked[i].Title < ranked[j].Title
				}
				return ranked[i].RefCount > ranked[j].RefCount
			}
			return ranked[i].ChannelRefCount > ranked[j].ChannelRefCount
		}
		return ranked[i].matchScore > ranked[j].matchScore
	})

	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	suggestions := make([]KnowledgeEntitySuggestion, 0, len(ranked))
	for _, item := range ranked {
		suggestions = append(suggestions, item.KnowledgeEntitySuggestion)
	}
	return suggestions, nil
}

func listChannelKnowledgeEntityRefs(database *gorm.DB, channelID string) ([]domain.KnowledgeEntityRef, error) {
	messageRefs, err := listChannelMessageKnowledgeEntityRefs(database, channelID)
	if err != nil {
		return nil, err
	}

	fileRefs, err := listChannelFileKnowledgeEntityRefs(database, channelID)
	if err != nil {
		return nil, err
	}

	refs := append(messageRefs, fileRefs...)
	sort.SliceStable(refs, func(i, j int) bool {
		if refs[i].CreatedAt.Equal(refs[j].CreatedAt) {
			return refs[i].ID < refs[j].ID
		}
		return refs[i].CreatedAt.After(refs[j].CreatedAt)
	})
	return refs, nil
}

func listChannelMessageKnowledgeEntityRefs(database *gorm.DB, channelID string) ([]domain.KnowledgeEntityRef, error) {
	var refs []domain.KnowledgeEntityRef
	if err := database.Model(&domain.KnowledgeEntityRef{}).
		Select("knowledge_entity_refs.*").
		Joins("JOIN messages ON messages.id = knowledge_entity_refs.ref_id").
		Where("knowledge_entity_refs.ref_kind = ? AND messages.channel_id = ?", "message", channelID).
		Order("knowledge_entity_refs.created_at desc").
		Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

func listChannelFileKnowledgeEntityRefs(database *gorm.DB, channelID string) ([]domain.KnowledgeEntityRef, error) {
	var refs []domain.KnowledgeEntityRef
	if err := database.Model(&domain.KnowledgeEntityRef{}).
		Select("knowledge_entity_refs.*").
		Joins("JOIN file_assets ON file_assets.id = knowledge_entity_refs.ref_id").
		Where("knowledge_entity_refs.ref_kind = ? AND file_assets.channel_id = ?", "file", channelID).
		Order("knowledge_entity_refs.created_at desc").
		Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

func listChannelMessageKnowledgeRefs(database *gorm.DB, channelID string) ([]ChannelKnowledgeRef, error) {
	refs, err := listChannelMessageKnowledgeEntityRefs(database, channelID)
	if err != nil {
		return nil, err
	}

	results := make([]ChannelKnowledgeRef, 0, len(refs))
	for _, ref := range refs {
		item, err := hydrateChannelKnowledgeRef(database, ref)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func listChannelFileKnowledgeRefs(database *gorm.DB, channelID string) ([]ChannelKnowledgeRef, error) {
	refs, err := listChannelFileKnowledgeEntityRefs(database, channelID)
	if err != nil {
		return nil, err
	}

	results := make([]ChannelKnowledgeRef, 0, len(refs))
	for _, ref := range refs {
		item, err := hydrateChannelKnowledgeRef(database, ref)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

func hydrateChannelKnowledgeRef(database *gorm.DB, ref domain.KnowledgeEntityRef) (ChannelKnowledgeRef, error) {
	item := ChannelKnowledgeRef{
		ID:        ref.ID,
		EntityID:  ref.EntityID,
		RefKind:   ref.RefKind,
		RefID:     ref.RefID,
		Role:      ref.Role,
		CreatedAt: ref.CreatedAt,
	}

	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", ref.EntityID).Error; err == nil {
		item.EntityTitle = entity.Title
		item.EntityKind = entity.Kind
	} else if err != gorm.ErrRecordNotFound {
		return item, err
	}

	switch ref.RefKind {
	case "message":
		var message domain.Message
		if err := database.First(&message, "id = ?", ref.RefID).Error; err == nil {
			item.SourceTitle = firstLine(message.Content)
			item.SourceSnippet = bestSnippet(message.Content, item.EntityTitle)
		} else if err != gorm.ErrRecordNotFound {
			return item, err
		}
	case "file":
		var file domain.FileAsset
		if err := database.First(&file, "id = ?", ref.RefID).Error; err == nil {
			item.SourceTitle = file.Name
			item.SourceSnippet = file.Summary
		} else if err != gorm.ErrRecordNotFound {
			return item, err
		}
	}

	return item, nil
}

func countKnowledgeEntityRefs(database *gorm.DB, entityIDs []string) (map[string]int, error) {
	counts := map[string]int{}
	if len(entityIDs) == 0 {
		return counts, nil
	}

	type refCountRow struct {
		EntityID string
		Count    int
	}

	var rows []refCountRow
	if err := database.Model(&domain.KnowledgeEntityRef{}).
		Select("entity_id, COUNT(*) as count").
		Where("entity_id IN ?", entityIDs).
		Group("entity_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.EntityID] = row.Count
	}
	return counts, nil
}

func loadKnowledgeEntityMap(database *gorm.DB, entityIDs []string) (map[string]domain.KnowledgeEntity, error) {
	result := map[string]domain.KnowledgeEntity{}
	if len(entityIDs) == 0 {
		return result, nil
	}

	var entities []domain.KnowledgeEntity
	if err := database.Where("id IN ?", entityIDs).Find(&entities).Error; err != nil {
		return nil, err
	}
	for _, entity := range entities {
		result[entity.ID] = entity
	}
	return result, nil
}

func knowledgeEntityMatchScore(entity domain.KnowledgeEntity, lowerQuery string) int {
	title := strings.ToLower(strings.TrimSpace(entity.Title))
	summary := strings.ToLower(strings.TrimSpace(entity.Summary))

	switch {
	case title == lowerQuery:
		return 4
	case strings.HasPrefix(title, lowerQuery):
		return 3
	case strings.Contains(title, lowerQuery):
		return 2
	case strings.Contains(summary, lowerQuery):
		return 1
	default:
		return 0
	}
}

func BuildEntityGraph(database *gorm.DB, entityID string) (EntityGraph, error) {
	entity, err := GetEntity(database, entityID)
	if err != nil {
		return EntityGraph{}, err
	}
	graph := EntityGraph{
		Nodes: []GraphNode{{ID: entity.ID, Kind: entity.Kind, Title: entity.Title, SourceKind: entity.SourceKind}},
		Edges: []GraphEdge{},
	}
	seen := map[string]bool{entity.ID: true}

	refs, err := ListEntityRefs(database, entityID)
	if err != nil {
		return graph, err
	}
	for _, ref := range refs {
		nodeID := ref.RefKind + ":" + ref.RefID
		if !seen[nodeID] {
			graph.Nodes = append(graph.Nodes, GraphNode{
				ID:         nodeID,
				Kind:       ref.RefKind,
				Title:      ref.RefID,
				SourceKind: "ref",
				RefKind:    ref.RefKind,
				RefID:      ref.RefID,
				Role:       ref.Role,
			})
			seen[nodeID] = true
		}
		graph.Edges = append(graph.Edges, GraphEdge{ID: ref.ID, From: entity.ID, To: nodeID, Relation: ref.Role, Weight: 1, Direction: "out", Role: ref.Role})
	}

	links, err := ListEntityLinks(database, entityID)
	if err != nil {
		return graph, err
	}
	for _, link := range links {
		otherID := link.ToEntityID
		if otherID == entityID {
			otherID = link.FromEntityID
		}
		if !seen[otherID] {
			if other, err := GetEntity(database, otherID); err == nil {
				graph.Nodes = append(graph.Nodes, GraphNode{ID: other.ID, Kind: other.Kind, Title: other.Title, SourceKind: other.SourceKind})
			} else {
				graph.Nodes = append(graph.Nodes, GraphNode{ID: otherID, Kind: "entity", Title: otherID})
			}
			seen[otherID] = true
		}
		direction := "out"
		if link.ToEntityID == entityID {
			direction = "in"
		}
		graph.Edges = append(graph.Edges, GraphEdge{ID: link.ID, From: link.FromEntityID, To: link.ToEntityID, Relation: link.Relation, Weight: link.Weight, Direction: direction, Role: link.Relation})
	}

	return graph, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func newKnowledgeID(prefix string) string {
	return ids.NewPrefixedUUID(prefix)
}

func normalizeKnowledgeText(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))
	lastWasSpace := true
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			builder.WriteByte(' ')
			lastWasSpace = true
		}
	}
	return strings.Join(strings.Fields(builder.String()), " ")
}

func includeKind(params LookupParams, kind string) bool {
	if len(params.IncludeKinds) == 0 {
		return true
	}
	return params.IncludeKinds[kind]
}

func scoreText(text, query string) (float64, bool) {
	content := strings.ToLower(strings.TrimSpace(text))
	needle := strings.ToLower(strings.TrimSpace(query))
	if content == "" || needle == "" {
		return 0, false
	}

	switch {
	case strings.Contains(content, needle):
		if strings.HasPrefix(content, needle) {
			return 3, true
		}
		return 2, true
	case strings.Contains(strings.Join(strings.Fields(content), " "), needle):
		return 1, true
	default:
		return 0, false
	}
}

func bestSnippet(text, query string) string {
	text = strings.TrimSpace(text)
	if len(text) <= 220 {
		return text
	}

	lower := strings.ToLower(text)
	needle := strings.ToLower(strings.TrimSpace(query))
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return text[:220]
	}

	start := idx - 60
	if start < 0 {
		start = 0
	}
	end := idx + len(needle) + 120
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

func buildLocator(locatorType, locatorValue string) string {
	locatorType = strings.TrimSpace(locatorType)
	locatorValue = strings.TrimSpace(locatorValue)
	if locatorType == "" {
		return locatorValue
	}
	if locatorValue == "" {
		return locatorType
	}
	return locatorType + ":" + locatorValue
}

func splitArtifactSections(content string) []string {
	parts := strings.Split(content, "\n\n")
	sections := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			sections = append(sections, part)
		}
	}
	if len(sections) == 0 && strings.TrimSpace(content) != "" {
		return []string{strings.TrimSpace(content)}
	}
	return sections
}

func firstLine(content string) string {
	line := strings.TrimSpace(strings.Split(content, "\n")[0])
	if line == "" {
		return ""
	}
	if len(line) > 80 {
		return line[:80]
	}
	return line
}
