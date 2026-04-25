package knowledge

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
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

		var asset domain.FileAsset
		if err := database.First(&asset, "id = ?", chunk.FileID).Error; err == nil {
			citation.MimeType = asset.ContentType
			citation.Size = asset.SizeBytes
			citation.PreviewURL = "/api/v1/files/" + asset.ID + "/preview"
			if citation.Title == "" {
				citation.Title = asset.Name
			}
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
	workspaceID := strings.TrimSpace(input.WorkspaceID)
	if workspaceID == "" {
		var workspace domain.Workspace
		if err := database.Order("id asc").First(&workspace).Error; err == nil {
			workspaceID = workspace.ID
		} else if err != gorm.ErrRecordNotFound {
			return domain.KnowledgeEntity{}, err
		}
	}
	metadata := strings.TrimSpace(input.Metadata)
	if metadata == "" && len(input.Tags) > 0 {
		tags := make([]string, 0, len(input.Tags))
		for _, tag := range input.Tags {
			if trimmed := strings.TrimSpace(tag); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
		if len(tags) > 0 {
			if payload, err := json.Marshal(map[string][]string{"tags": tags}); err == nil {
				metadata = string(payload)
			}
		}
	}
	entity := domain.KnowledgeEntity{
		ID:          newKnowledgeID("entity"),
		WorkspaceID: workspaceID,
		Kind:        defaultString(strings.TrimSpace(input.Kind), "custom"),
		Title:       strings.TrimSpace(input.Title),
		Summary:     strings.TrimSpace(input.Summary),
		Status:      defaultString(strings.TrimSpace(input.Status), "active"),
		OwnerUserID: strings.TrimSpace(input.OwnerUserID),
		SourceKind:  defaultString(strings.TrimSpace(input.SourceKind), "manual"),
		SourceRef:   strings.TrimSpace(input.SourceRef),
		Metadata:    metadata,
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
		summary.Velocity = buildChannelKnowledgeVelocity(minInt(3, windowDays), 0, 0)
		return summary, nil
	}

	windowStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(windowDays - 1))
	recentWindowDays := minInt(3, windowDays)
	recentWindowStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(recentWindowDays - 1))
	previousWindowStart := recentWindowStart.AddDate(0, 0, -recentWindowDays)
	recentVelocityCount := 0
	previousVelocityCount := 0
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
		if !ref.CreatedAt.Before(recentWindowStart) {
			recentVelocityCount++
		} else if !ref.CreatedAt.Before(previousWindowStart) {
			previousVelocityCount++
		}
		if !ref.CreatedAt.Before(windowStart) {
			summary.RecentRefCount++
			item.TrendCounts[ref.CreatedAt.UTC().Format("2006-01-02")]++
		}
	}
	summary.Velocity = buildChannelKnowledgeVelocity(recentWindowDays, recentVelocityCount, previousVelocityCount)

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

func FollowEntity(database *gorm.DB, entityID, userID string) (domain.KnowledgeEntityFollow, error) {
	entityID = strings.TrimSpace(entityID)
	userID = strings.TrimSpace(userID)
	if entityID == "" || userID == "" {
		return domain.KnowledgeEntityFollow{}, errors.New("entity_id and user_id are required")
	}

	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", entityID).Error; err != nil {
		return domain.KnowledgeEntityFollow{}, err
	}
	if err := database.First(&domain.User{}, "id = ?", userID).Error; err != nil {
		return domain.KnowledgeEntityFollow{}, err
	}

	var existing domain.KnowledgeEntityFollow
	err := database.Where("entity_id = ? AND user_id = ?", entity.ID, userID).First(&existing).Error
	if err == nil {
		if strings.TrimSpace(existing.NotificationLevel) == "" {
			existing.NotificationLevel = "all"
			_ = database.Model(&existing).Update("notification_level", existing.NotificationLevel).Error
		}
		return existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return domain.KnowledgeEntityFollow{}, err
	}

	follow := domain.KnowledgeEntityFollow{
		ID:                ids.NewPrefixedUUID("kfollow"),
		WorkspaceID:       entity.WorkspaceID,
		EntityID:          entity.ID,
		UserID:            userID,
		NotificationLevel: "all",
		CreatedAt:         time.Now().UTC(),
	}
	if err := database.Create(&follow).Error; err != nil {
		return domain.KnowledgeEntityFollow{}, err
	}
	return follow, nil
}

func UpdateFollowNotificationLevel(database *gorm.DB, entityID, userID, notificationLevel string) (domain.KnowledgeEntityFollow, error) {
	entityID = strings.TrimSpace(entityID)
	userID = strings.TrimSpace(userID)
	notificationLevel = normalizeFollowNotificationLevel(notificationLevel)
	if entityID == "" || userID == "" {
		return domain.KnowledgeEntityFollow{}, errors.New("entity_id and user_id are required")
	}
	if notificationLevel == "" {
		return domain.KnowledgeEntityFollow{}, errors.New("notification_level must be all, digest_only, or silent")
	}

	var follow domain.KnowledgeEntityFollow
	if err := database.First(&follow, "entity_id = ? AND user_id = ?", entityID, userID).Error; err != nil {
		return domain.KnowledgeEntityFollow{}, err
	}
	follow.NotificationLevel = notificationLevel
	if err := database.Save(&follow).Error; err != nil {
		return domain.KnowledgeEntityFollow{}, err
	}
	return follow, nil
}

func BulkUpdateFollowNotificationLevels(database *gorm.DB, userID string, entityIDs []string, notificationLevel string) ([]domain.KnowledgeEntityFollow, error) {
	userID = strings.TrimSpace(userID)
	notificationLevel = normalizeFollowNotificationLevel(notificationLevel)
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if notificationLevel == "" {
		return nil, errors.New("notification_level must be all, digest_only, or silent")
	}

	normalizedIDs := make([]string, 0, len(entityIDs))
	seen := map[string]bool{}
	for _, entityID := range entityIDs {
		entityID = strings.TrimSpace(entityID)
		if entityID == "" || seen[entityID] {
			continue
		}
		seen[entityID] = true
		normalizedIDs = append(normalizedIDs, entityID)
	}
	if len(normalizedIDs) == 0 {
		return []domain.KnowledgeEntityFollow{}, nil
	}

	if err := database.Model(&domain.KnowledgeEntityFollow{}).
		Where("user_id = ? AND entity_id IN ?", userID, normalizedIDs).
		Update("notification_level", notificationLevel).Error; err != nil {
		return nil, err
	}

	var follows []domain.KnowledgeEntityFollow
	if err := database.Where("user_id = ? AND entity_id IN ?", userID, normalizedIDs).
		Order("created_at desc, id desc").
		Find(&follows).Error; err != nil {
		return nil, err
	}
	return follows, nil
}

func UnfollowEntity(database *gorm.DB, entityID, userID string) error {
	entityID = strings.TrimSpace(entityID)
	userID = strings.TrimSpace(userID)
	if entityID == "" || userID == "" {
		return errors.New("entity_id and user_id are required")
	}
	if err := database.First(&domain.KnowledgeEntity{}, "id = ?", entityID).Error; err != nil {
		return err
	}
	return database.Where("entity_id = ? AND user_id = ?", entityID, userID).Delete(&domain.KnowledgeEntityFollow{}).Error
}

func ListFollowedEntities(database *gorm.DB, userID string) ([]FollowedEntity, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return []FollowedEntity{}, nil
	}

	var follows []domain.KnowledgeEntityFollow
	if err := database.Where("user_id = ?", userID).Order("created_at desc, id desc").Find(&follows).Error; err != nil {
		return nil, err
	}
	items := make([]FollowedEntity, 0, len(follows))
	for _, follow := range follows {
		var entity domain.KnowledgeEntity
		if err := database.First(&entity, "id = ?", follow.EntityID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return nil, err
		}
		items = append(items, FollowedEntity{
			Follow:      follow,
			Entity:      entity,
			IsFollowing: true,
		})
	}
	return items, nil
}

func GetFollowedEntityStats(database *gorm.DB, userID string, now time.Time) (FollowedEntityStats, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return FollowedEntityStats{}, nil
	}

	var follows []domain.KnowledgeEntityFollow
	if err := database.Where("user_id = ?", userID).Find(&follows).Error; err != nil {
		return FollowedEntityStats{}, err
	}

	stats := FollowedEntityStats{
		ByKind: make([]FollowedEntityStatsKindCount, 0),
	}
	kindCounts := map[string]int{}

	for _, follow := range follows {
		var entity domain.KnowledgeEntity
		if err := database.First(&entity, "id = ?", follow.EntityID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return FollowedEntityStats{}, err
		}
		stats.TotalCount++
		kindCounts[entity.Kind]++
		if normalizeFollowNotificationLevel(follow.NotificationLevel) == "silent" {
			stats.MutedCount++
		}

		settings, err := getWorkspaceKnowledgeSettingsOrDefault(database, follow.WorkspaceID)
		if err != nil {
			return FollowedEntityStats{}, err
		}
		if follow.LastAlertedAt != nil && follow.LastAlertedAt.After(now.Add(-time.Duration(settings.SpikeCooldownMinutes)*time.Minute)) {
			stats.SpikingCount++
		}
	}

	for kind, count := range kindCounts {
		stats.ByKind = append(stats.ByKind, FollowedEntityStatsKindCount{
			Kind:  kind,
			Count: count,
		})
	}
	sort.SliceStable(stats.ByKind, func(i, j int) bool {
		if stats.ByKind[i].Count == stats.ByKind[j].Count {
			return stats.ByKind[i].Kind < stats.ByKind[j].Kind
		}
		return stats.ByKind[i].Count > stats.ByKind[j].Count
	})

	return stats, nil
}

func BuildSharedEntityLink(database *gorm.DB, entityID, baseURL string) (SharedEntityLink, error) {
	entityID = strings.TrimSpace(entityID)
	if entityID == "" {
		return SharedEntityLink{}, errors.New("entity_id is required")
	}

	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", entityID).Error; err != nil {
		return SharedEntityLink{}, err
	}

	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	relativePath := "/workspace/knowledge/" + entity.ID
	return SharedEntityLink{
		EntityID:     entity.ID,
		WorkspaceID:  entity.WorkspaceID,
		Title:        entity.Title,
		URL:          baseURL + relativePath,
		ShortURL:     baseURL + "/k/" + entity.ID,
		RelativePath: relativePath,
	}, nil
}

func BuildSharedWeeklyBriefLink(database *gorm.DB, summaryID, baseURL string) (SharedWeeklyBriefLink, error) {
	summaryID = strings.TrimSpace(summaryID)
	if summaryID == "" {
		return SharedWeeklyBriefLink{}, errors.New("summary_id is required")
	}

	var summary domain.AISummary
	if err := database.Where("id = ? AND scope_type = ?", summaryID, "knowledge_weekly").First(&summary).Error; err != nil {
		return SharedWeeklyBriefLink{}, err
	}

	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	userID := ""
	workspaceID := summary.ChannelID
	parts := strings.Split(summary.ScopeID, ":")
	if len(parts) >= 3 {
		userID = strings.TrimSpace(parts[0])
		if strings.TrimSpace(parts[1]) != "" {
			workspaceID = strings.TrimSpace(parts[1])
		}
	}

	relativePath := "/workspace/knowledge/following?brief=" + summaryID
	return SharedWeeklyBriefLink{
		ID:           summaryID,
		UserID:       userID,
		WorkspaceID:  workspaceID,
		Title:        "Weekly Knowledge Digest",
		URL:          baseURL + relativePath,
		ShortURL:     baseURL + "/kb/" + summaryID,
		RelativePath: relativePath,
	}, nil
}

func BuildEntityBriefPrompt(database *gorm.DB, entityID string) (domain.KnowledgeEntity, []ChannelKnowledgeRef, []domain.KnowledgeEvent, string, error) {
	entity, err := GetEntity(database, entityID)
	if err != nil {
		return domain.KnowledgeEntity{}, nil, nil, "", err
	}

	var refs []domain.KnowledgeEntityRef
	if err := database.Where("entity_id = ?", entity.ID).Order("created_at desc").Limit(20).Find(&refs).Error; err != nil {
		return entity, nil, nil, "", err
	}
	hydratedRefs := make([]ChannelKnowledgeRef, 0, len(refs))
	for _, ref := range refs {
		hydrated, err := hydrateChannelKnowledgeRef(database, ref)
		if err != nil {
			continue
		}
		hydratedRefs = append(hydratedRefs, hydrated)
	}

	var events []domain.KnowledgeEvent
	if err := database.Where("entity_id = ?", entity.ID).Order("occurred_at desc, created_at desc").Limit(10).Find(&events).Error; err != nil {
		return entity, hydratedRefs, nil, "", err
	}

	var prompt strings.Builder
	prompt.WriteString("Write an AI-native workspace knowledge brief for this entity. Return concise markdown with: Summary, Key Discussions, Risks, Next Actions.\n\n")
	prompt.WriteString("Entity: ")
	prompt.WriteString(entity.Title)
	prompt.WriteString("\nKind: ")
	prompt.WriteString(entity.Kind)
	if strings.TrimSpace(entity.Summary) != "" {
		prompt.WriteString("\nExisting summary: ")
		prompt.WriteString(entity.Summary)
	}
	prompt.WriteString("\n\nRecent references:\n")
	for _, ref := range hydratedRefs {
		prompt.WriteString("- ")
		prompt.WriteString(ref.RefKind)
		prompt.WriteString(" / ")
		prompt.WriteString(ref.Role)
		prompt.WriteString(": ")
		prompt.WriteString(ref.SourceSnippet)
		prompt.WriteString("\n")
	}
	prompt.WriteString("\nRecent timeline events:\n")
	for _, event := range events {
		prompt.WriteString("- ")
		prompt.WriteString(event.EventType)
		prompt.WriteString(": ")
		prompt.WriteString(event.Title)
		if strings.TrimSpace(event.Body) != "" {
			prompt.WriteString(" - ")
			prompt.WriteString(event.Body)
		}
		prompt.WriteString("\n")
	}
	return entity, hydratedRefs, events, prompt.String(), nil
}

func BuildEntityAskPrompt(database *gorm.DB, entityID, question string) (domain.KnowledgeEntity, []Citation, string, error) {
	entity, refs, events, prompt, err := BuildEntityBriefPrompt(database, entityID)
	if err != nil {
		return domain.KnowledgeEntity{}, nil, "", err
	}

	links, err := ListEntityLinks(database, entity.ID)
	if err != nil {
		return entity, nil, "", err
	}

	citations := make([]Citation, 0, len(refs))
	for idx, ref := range refs {
		if idx >= 5 {
			break
		}
		citation := Citation{
			ID:           ref.ID,
			EvidenceKind: ref.RefKind,
			SourceKind:   ref.RefKind,
			SourceRef:    ref.RefID,
			RefKind:      ref.RefKind,
			Snippet:      ref.SourceSnippet,
			Title:        ref.SourceTitle,
			Score:        float64(len(refs) - idx),
			EntityID:     ref.EntityID,
			EntityTitle:  ref.EntityTitle,
		}
		if citation.Snippet == "" {
			citation.Snippet = ref.SourceTitle
		}
		citations = append(citations, citation)
	}

	var builder strings.Builder
	builder.WriteString("Answer the user's question using only the grounded workspace knowledge below. If evidence is thin, say so clearly. Return concise markdown.\n\n")
	builder.WriteString("Question: ")
	builder.WriteString(strings.TrimSpace(question))
	builder.WriteString("\n\n")
	builder.WriteString(prompt)
	if len(links) > 0 {
		builder.WriteString("\nRelated linked entities:\n")
		for idx, link := range links {
			if idx >= 8 {
				break
			}
			builder.WriteString("- ")
			builder.WriteString(link.Relation)
			builder.WriteString(": ")
			if link.FromEntityID == entity.ID {
				builder.WriteString(link.ToEntityID)
			} else {
				builder.WriteString(link.FromEntityID)
			}
			builder.WriteString("\n")
		}
	}
	if len(events) == 0 && len(citations) == 0 {
		builder.WriteString("\nThere is limited evidence for this entity.\n")
	}

	return entity, citations, builder.String(), nil
}

func BuildWeeklyBriefPrompt(database *gorm.DB, userID, workspaceID string, now time.Time) (FollowedEntityStats, []FollowedEntity, []TrendingEntity, string, error) {
	userID = strings.TrimSpace(userID)
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		workspaceID = "ws-1"
	}

	stats, err := GetFollowedEntityStats(database, userID, now)
	if err != nil {
		return FollowedEntityStats{}, nil, nil, "", err
	}
	followed, err := ListFollowedEntities(database, userID)
	if err != nil {
		return stats, nil, nil, "", err
	}
	trending, err := GetTrendingEntities(database, TrendingEntitiesParams{WorkspaceID: workspaceID, Days: 7, Limit: 10, Now: now})
	if err != nil {
		return stats, followed, nil, "", err
	}

	var prompt strings.Builder
	prompt.WriteString("Write a weekly AI-native knowledge brief for a busy teammate. Focus on followed entities, emerging topics, risks, and next actions. Return concise markdown.\n\n")
	prompt.WriteString(fmt.Sprintf("Follow stats: total=%d spiking=%d muted=%d\n", stats.TotalCount, stats.SpikingCount, stats.MutedCount))
	prompt.WriteString("\nFollowed entities:\n")
	for _, item := range followed {
		if item.Entity.WorkspaceID != workspaceID {
			continue
		}
		prompt.WriteString("- ")
		prompt.WriteString(item.Entity.Title)
		prompt.WriteString(" (")
		prompt.WriteString(item.Entity.Kind)
		prompt.WriteString(", alerts=")
		prompt.WriteString(item.Follow.NotificationLevel)
		prompt.WriteString(")\n")
	}
	prompt.WriteString("\nTrending entities:\n")
	for _, item := range trending {
		prompt.WriteString(fmt.Sprintf("- %s: recent_refs=%d delta=%d\n", item.Entity.Title, item.RecentRefCount, item.VelocityDelta))
	}
	return stats, followed, trending, prompt.String(), nil
}

func GetEntityActivityBackfillStatus(database *gorm.DB, entityID string) (ActivityBackfillStatus, error) {
	return backfillEntityActivity(database, entityID, false)
}

func BackfillEntityActivity(database *gorm.DB, entityID string) (ActivityBackfillStatus, []domain.KnowledgeEntityRef, error) {
	var existingIDs []string
	if err := database.Model(&domain.KnowledgeEntityRef{}).Where("entity_id = ?", entityID).Pluck("id", &existingIDs).Error; err != nil {
		return ActivityBackfillStatus{}, nil, err
	}

	status, err := backfillEntityActivity(database, entityID, true)
	if err != nil {
		return status, nil, err
	}
	var refs []domain.KnowledgeEntityRef
	if status.CreatedRefCount > 0 {
		query := database.Where("entity_id = ?", entityID)
		if len(existingIDs) > 0 {
			query = query.Where("id NOT IN ?", existingIDs)
		}
		if err := query.Order("created_at desc").Limit(status.CreatedRefCount).Find(&refs).Error; err != nil {
			return status, nil, err
		}
	}
	return status, refs, nil
}

func backfillEntityActivity(database *gorm.DB, entityID string, mutate bool) (ActivityBackfillStatus, error) {
	entity, err := GetEntity(database, entityID)
	if err != nil {
		return ActivityBackfillStatus{}, err
	}
	status := ActivityBackfillStatus{
		EntityID:    entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Title:       entity.Title,
	}

	var existingRefs []domain.KnowledgeEntityRef
	if err := database.Where("entity_id = ?", entity.ID).Find(&existingRefs).Error; err != nil {
		return status, err
	}
	status.ExistingRefCount = len(existingRefs)
	existing := map[string]struct{}{}
	for _, ref := range existingRefs {
		existing[ref.RefKind+":"+ref.RefID] = struct{}{}
		if status.LastRefAt == nil || ref.CreatedAt.After(*status.LastRefAt) {
			ts := ref.CreatedAt
			status.LastRefAt = &ts
		}
	}

	var messages []domain.Message
	if err := database.Model(&domain.Message{}).
		Select("messages.*").
		Joins("JOIN channels ON channels.id = messages.channel_id").
		Where("channels.workspace_id = ? AND LOWER(messages.content) LIKE ?", entity.WorkspaceID, "%"+strings.ToLower(entity.Title)+"%").
		Order("messages.created_at asc").
		Find(&messages).Error; err != nil {
		return status, err
	}
	for _, message := range messages {
		status.MessageCandidateCount++
		key := "message:" + message.ID
		if _, ok := existing[key]; ok {
			continue
		}
		status.MissingRefCount++
		if mutate {
			ref := domain.KnowledgeEntityRef{
				ID:          newKnowledgeID("kref"),
				WorkspaceID: entity.WorkspaceID,
				EntityID:    entity.ID,
				RefKind:     "message",
				RefID:       message.ID,
				Role:        "backfilled",
				CreatedAt:   message.CreatedAt,
			}
			if err := database.Create(&ref).Error; err != nil {
				return status, err
			}
			existing[key] = struct{}{}
			status.CreatedRefCount++
			if status.LastRefAt == nil || ref.CreatedAt.After(*status.LastRefAt) {
				ts := ref.CreatedAt
				status.LastRefAt = &ts
			}
		}
	}

	var files []domain.FileAsset
	if err := database.Model(&domain.FileAsset{}).
		Select("file_assets.*").
		Joins("JOIN channels ON channels.id = file_assets.channel_id").
		Where("channels.workspace_id = ?", entity.WorkspaceID).
		Find(&files).Error; err != nil {
		return status, err
	}
	for _, file := range files {
		content := strings.Join([]string{file.Name, file.Description, file.Summary, file.ContentSummary}, " ")
		var extraction domain.FileExtraction
		if err := database.Where("file_id = ?", file.ID).First(&extraction).Error; err == nil {
			content += " " + extraction.ContentText + " " + extraction.ContentSummary
		}
		if !strings.Contains(strings.ToLower(content), strings.ToLower(entity.Title)) {
			continue
		}
		status.FileCandidateCount++
		key := "file:" + file.ID
		if _, ok := existing[key]; ok {
			continue
		}
		status.MissingRefCount++
		if mutate {
			ref := domain.KnowledgeEntityRef{
				ID:          newKnowledgeID("kref"),
				WorkspaceID: entity.WorkspaceID,
				EntityID:    entity.ID,
				RefKind:     "file",
				RefID:       file.ID,
				Role:        "backfilled",
				CreatedAt:   file.CreatedAt,
			}
			if err := database.Create(&ref).Error; err != nil {
				return status, err
			}
			existing[key] = struct{}{}
			status.CreatedRefCount++
			if status.LastRefAt == nil || ref.CreatedAt.After(*status.LastRefAt) {
				ts := ref.CreatedAt
				status.LastRefAt = &ts
			}
		}
	}
	if mutate {
		status.MissingRefCount = 0
	}
	status.IsBackfilled = status.MissingRefCount == 0
	return status, nil
}

func GetWorkspaceKnowledgeSettings(database *gorm.DB, workspaceID string) (WorkspaceKnowledgeSettings, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return WorkspaceKnowledgeSettings{}, errors.New("workspace_id is required")
	}

	var workspace domain.Workspace
	if err := database.First(&workspace, "id = ?", workspaceID).Error; err != nil {
		return WorkspaceKnowledgeSettings{}, err
	}
	return normalizeWorkspaceKnowledgeSettings(workspace), nil
}

func UpdateWorkspaceKnowledgeSettings(database *gorm.DB, workspaceID string, spikeThreshold, spikeCooldownMinutes int) (WorkspaceKnowledgeSettings, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return WorkspaceKnowledgeSettings{}, errors.New("workspace_id is required")
	}
	if spikeThreshold <= 0 {
		return WorkspaceKnowledgeSettings{}, errors.New("spike_threshold must be greater than 0")
	}
	if spikeCooldownMinutes <= 0 {
		return WorkspaceKnowledgeSettings{}, errors.New("spike_cooldown_minutes must be greater than 0")
	}

	var workspace domain.Workspace
	if err := database.First(&workspace, "id = ?", workspaceID).Error; err != nil {
		return WorkspaceKnowledgeSettings{}, err
	}
	workspace.KnowledgeSpikeThreshold = spikeThreshold
	workspace.KnowledgeSpikeCooldownMins = spikeCooldownMinutes
	if err := database.Save(&workspace).Error; err != nil {
		return WorkspaceKnowledgeSettings{}, err
	}
	return normalizeWorkspaceKnowledgeSettings(workspace), nil
}

func GetEntityActivity(database *gorm.DB, entityID string, days int, now time.Time) (EntityActivity, error) {
	entityID = strings.TrimSpace(entityID)
	if entityID == "" {
		return EntityActivity{}, errors.New("entity_id is required")
	}
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", entityID).Error; err != nil {
		return EntityActivity{}, err
	}

	end := truncateToUTCDate(now)
	start := end.AddDate(0, 0, -(days - 1))

	var refs []domain.KnowledgeEntityRef
	if err := database.Where("entity_id = ? AND created_at >= ? AND created_at < ?", entityID, start, end.AddDate(0, 0, 1)).
		Order("created_at asc").
		Find(&refs).Error; err != nil {
		return EntityActivity{}, err
	}

	counts := map[string]int{}
	for _, ref := range refs {
		key := truncateToUTCDate(ref.CreatedAt).Format("2006-01-02")
		counts[key]++
	}

	buckets := make([]EntityActivityBucket, 0, days)
	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		buckets = append(buckets, EntityActivityBucket{
			Date:     key,
			RefCount: counts[key],
		})
	}

	return EntityActivity{
		EntityID:    entity.ID,
		WorkspaceID: entity.WorkspaceID,
		Days:        days,
		Buckets:     buckets,
	}, nil
}

func GetTrendingEntities(database *gorm.DB, params TrendingEntitiesParams) ([]TrendingEntity, error) {
	workspaceID := strings.TrimSpace(params.WorkspaceID)
	if workspaceID == "" {
		return nil, errors.New("workspace_id is required")
	}
	days := params.Days
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}
	limit := params.Limit
	if limit <= 0 {
		limit = 5
	}
	now := params.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	recentStart := now.AddDate(0, 0, -days)
	previousStart := recentStart.AddDate(0, 0, -days)

	var entities []domain.KnowledgeEntity
	if err := database.Where("workspace_id = ? AND status <> ?", workspaceID, "archived").Find(&entities).Error; err != nil {
		return nil, err
	}

	items := make([]TrendingEntity, 0, len(entities))
	for _, entity := range entities {
		var recentCount int64
		if err := database.Model(&domain.KnowledgeEntityRef{}).
			Where("entity_id = ? AND created_at >= ? AND created_at <= ?", entity.ID, recentStart, now).
			Count(&recentCount).Error; err != nil {
			return nil, err
		}
		if recentCount == 0 {
			continue
		}

		var previousCount int64
		if err := database.Model(&domain.KnowledgeEntityRef{}).
			Where("entity_id = ? AND created_at >= ? AND created_at < ?", entity.ID, previousStart, recentStart).
			Count(&previousCount).Error; err != nil {
			return nil, err
		}

		relatedChannelIDs, err := listEntityRelatedChannelIDs(database, entity.ID)
		if err != nil {
			return nil, err
		}

		var lastRef domain.KnowledgeEntityRef
		var lastRefAt *time.Time
		if err := database.Where("entity_id = ?", entity.ID).Order("created_at desc").First(&lastRef).Error; err == nil {
			ts := lastRef.CreatedAt
			lastRefAt = &ts
		}

		items = append(items, TrendingEntity{
			Entity:            entity,
			RecentRefCount:    int(recentCount),
			PreviousRefCount:  int(previousCount),
			VelocityDelta:     int(recentCount - previousCount),
			RelatedChannelIDs: relatedChannelIDs,
			LastRefAt:         lastRefAt,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].VelocityDelta == items[j].VelocityDelta {
			if items[i].RecentRefCount == items[j].RecentRefCount {
				return items[i].Entity.Title < items[j].Entity.Title
			}
			return items[i].RecentRefCount > items[j].RecentRefCount
		}
		return items[i].VelocityDelta > items[j].VelocityDelta
	})

	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func DetectEntitySpikeAlerts(database *gorm.DB, entityID, channelID string, now time.Time) ([]EntitySpikeAlert, error) {
	entityID = strings.TrimSpace(entityID)
	channelID = strings.TrimSpace(channelID)
	if entityID == "" {
		return []EntitySpikeAlert{}, errors.New("entity_id is required")
	}

	var entity domain.KnowledgeEntity
	if err := database.First(&entity, "id = ?", entityID).Error; err != nil {
		return nil, err
	}

	windowStart := now.Add(-24 * time.Hour)
	previousWindowStart := now.Add(-48 * time.Hour)

	var recentRefCount int64
	if err := database.Model(&domain.KnowledgeEntityRef{}).
		Where("entity_id = ? AND created_at >= ? AND created_at <= ?", entityID, windowStart, now).
		Count(&recentRefCount).Error; err != nil {
		return nil, err
	}
	settings, err := getWorkspaceKnowledgeSettingsOrDefault(database, entity.WorkspaceID)
	if err != nil {
		return nil, err
	}

	if recentRefCount < int64(settings.SpikeThreshold) {
		return []EntitySpikeAlert{}, nil
	}

	var previousRefCount int64
	if err := database.Model(&domain.KnowledgeEntityRef{}).
		Where("entity_id = ? AND created_at >= ? AND created_at < ?", entityID, previousWindowStart, windowStart).
		Count(&previousRefCount).Error; err != nil {
		return nil, err
	}

	delta := int(recentRefCount - previousRefCount)
	if delta < 2 && !(previousRefCount == 0 && recentRefCount >= 3) {
		return []EntitySpikeAlert{}, nil
	}

	relatedChannelIDs, err := listEntityRelatedChannelIDs(database, entityID)
	if err != nil {
		return nil, err
	}

	var follows []domain.KnowledgeEntityFollow
	if err := database.Where("entity_id = ?", entityID).Find(&follows).Error; err != nil {
		return nil, err
	}

	notifiedUserIDs := make([]string, 0)
	for _, follow := range follows {
		if normalizeFollowNotificationLevel(follow.NotificationLevel) != "all" {
			continue
		}
		if follow.LastAlertedAt != nil && follow.LastAlertedAt.After(now.Add(-time.Duration(settings.SpikeCooldownMinutes)*time.Minute)) {
			continue
		}
		notifiedUserIDs = append(notifiedUserIDs, follow.UserID)
	}
	if len(notifiedUserIDs) == 0 {
		return []EntitySpikeAlert{}, nil
	}

	if err := database.Model(&domain.KnowledgeEntityFollow{}).
		Where("entity_id = ? AND user_id IN ?", entityID, notifiedUserIDs).
		Updates(map[string]any{"last_alerted_at": now.UTC()}).Error; err != nil {
		return nil, err
	}

	return []EntitySpikeAlert{{
		Entity:            entity,
		UserIDs:           notifiedUserIDs,
		ChannelID:         channelID,
		RecentRefCount:    int(recentRefCount),
		PreviousRefCount:  int(previousRefCount),
		Delta:             delta,
		RelatedChannelIDs: relatedChannelIDs,
		OccurredAt:        now.UTC(),
	}}, nil
}

func MatchEntitiesInText(database *gorm.DB, input MatchEntitiesInput) ([]EntityTextMatch, error) {
	text := strings.TrimSpace(input.Text)
	if text == "" {
		return []EntityTextMatch{}, nil
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var entities []domain.KnowledgeEntity
	query := database.Model(&domain.KnowledgeEntity{}).Where("status <> ?", "archived")
	if strings.TrimSpace(input.WorkspaceID) != "" {
		query = query.Where("workspace_id = ?", strings.TrimSpace(input.WorkspaceID))
	}
	if err := query.Order("LENGTH(title) DESC, title ASC").Find(&entities).Error; err != nil {
		return nil, err
	}

	lowerText := strings.ToLower(text)
	occupied := make([]bool, len(text))
	matches := make([]EntityTextMatch, 0)
	for _, entity := range entities {
		title := strings.TrimSpace(entity.Title)
		if title == "" {
			continue
		}
		lowerTitle := strings.ToLower(title)
		searchFrom := 0
		for searchFrom < len(lowerText) {
			idx := strings.Index(lowerText[searchFrom:], lowerTitle)
			if idx < 0 {
				break
			}
			start := searchFrom + idx
			end := start + len(lowerTitle)
			searchFrom = end
			if !isEntityTextBoundary(lowerText, start, end) || spanOccupied(occupied, start, end) {
				continue
			}
			for i := start; i < end && i < len(occupied); i++ {
				occupied[i] = true
			}
			matches = append(matches, EntityTextMatch{
				EntityID:    entity.ID,
				EntityTitle: entity.Title,
				EntityKind:  entity.Kind,
				SourceKind:  entity.SourceKind,
				MatchedText: text[start:end],
				Start:       start,
				End:         end,
			})
			if len(matches) >= limit {
				sort.SliceStable(matches, func(i, j int) bool {
					return matches[i].Start < matches[j].Start
				})
				return matches, nil
			}
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].Start < matches[j].Start
	})
	return matches, nil
}

func FindMentionedEntities(database *gorm.DB, workspaceID string, content string) ([]MentionedEntity, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return []MentionedEntity{}, nil
	}

	entities, err := ListEntities(database, workspaceID)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return []MentionedEntity{}, nil
	}

	plain := normalizeMentionContent(content)
	if plain == "" || !strings.Contains(plain, "@") {
		return []MentionedEntity{}, nil
	}
	lowerContent := strings.ToLower(plain)

	sort.SliceStable(entities, func(i, j int) bool {
		if len(entities[i].Title) == len(entities[j].Title) {
			return entities[i].Title < entities[j].Title
		}
		return len(entities[i].Title) > len(entities[j].Title)
	})

	mentions := make([]MentionedEntity, 0)
	seen := map[string]struct{}{}
	occupied := map[int]struct{}{}
	for _, entity := range entities {
		title := strings.TrimSpace(entity.Title)
		if title == "" {
			continue
		}
		mentionText := "@" + title
		start, end, ok := findExplicitMentionRange(lowerContent, strings.ToLower(mentionText), occupied)
		if !ok {
			continue
		}
		if _, exists := seen[entity.ID]; exists {
			continue
		}
		for idx := start; idx < end; idx++ {
			occupied[idx] = struct{}{}
		}
		seen[entity.ID] = struct{}{}
		mentions = append(mentions, MentionedEntity{
			EntityID:    entity.ID,
			EntityTitle: entity.Title,
			EntityKind:  entity.Kind,
			SourceKind:  entity.SourceKind,
			MentionText: mentionText,
		})
	}
	return mentions, nil
}

type EntityMessageSearchParams struct {
	EntityID  string
	ChannelID string
	Limit     int
}

type EntityMessageMatch struct {
	Message      domain.Message
	EntityID     string
	EntityTitle  string
	MatchSources []string
}

func FindMessagesByEntity(database *gorm.DB, params EntityMessageSearchParams) ([]EntityMessageMatch, error) {
	entityID := strings.TrimSpace(params.EntityID)
	if entityID == "" {
		return []EntityMessageMatch{}, errors.New("entity_id is required")
	}

	entity, err := GetEntity(database, entityID)
	if err != nil {
		return nil, err
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	channelID := strings.TrimSpace(params.ChannelID)
	lowerTitle := strings.ToLower(strings.TrimSpace(entity.Title))

	type matchAccumulator struct {
		Message      domain.Message
		MatchSources map[string]struct{}
	}

	matches := map[string]*matchAccumulator{}
	messageRefQuery := database.Model(&domain.Message{}).
		Select("messages.*").
		Joins("JOIN knowledge_entity_refs ON knowledge_entity_refs.ref_id = messages.id AND knowledge_entity_refs.ref_kind = ?", "message").
		Where("knowledge_entity_refs.entity_id = ?", entityID)
	if channelID != "" {
		messageRefQuery = messageRefQuery.Where("messages.channel_id = ?", channelID)
	}

	var refMessages []domain.Message
	if err := messageRefQuery.Order("messages.created_at desc").Find(&refMessages).Error; err != nil {
		return nil, err
	}
	for _, message := range refMessages {
		accumulator := matches[message.ID]
		if accumulator == nil {
			accumulator = &matchAccumulator{
				Message:      message,
				MatchSources: map[string]struct{}{},
			}
			matches[message.ID] = accumulator
		}
		accumulator.MatchSources["knowledge_ref"] = struct{}{}
	}

	if lowerTitle != "" {
		needle := "%" + lowerTitle + "%"
		contentQuery := database.Model(&domain.Message{}).
			Select("messages.*").
			Joins("JOIN channels ON channels.id = messages.channel_id").
			Where("channels.workspace_id = ? AND LOWER(messages.content) LIKE ?", entity.WorkspaceID, needle)
		if channelID != "" {
			contentQuery = contentQuery.Where("messages.channel_id = ?", channelID)
		}

		var titleMessages []domain.Message
		if err := contentQuery.Order("messages.created_at desc").Find(&titleMessages).Error; err != nil {
			return nil, err
		}
		for _, message := range titleMessages {
			accumulator := matches[message.ID]
			if accumulator == nil {
				accumulator = &matchAccumulator{
					Message:      message,
					MatchSources: map[string]struct{}{},
				}
				matches[message.ID] = accumulator
			}

			mentions, err := FindMentionedEntities(database, entity.WorkspaceID, message.Content)
			if err != nil {
				return nil, err
			}

			explicit := false
			for _, mention := range mentions {
				if mention.EntityID == entityID {
					accumulator.MatchSources["explicit_mention"] = struct{}{}
					explicit = true
					break
				}
			}
			if !explicit {
				accumulator.MatchSources["title_match"] = struct{}{}
			}
		}
	}

	results := make([]EntityMessageMatch, 0, len(matches))
	for _, accumulator := range matches {
		matchSources := make([]string, 0, len(accumulator.MatchSources))
		for matchSource := range accumulator.MatchSources {
			matchSources = append(matchSources, matchSource)
		}
		sort.Strings(matchSources)
		results = append(results, EntityMessageMatch{
			Message:      accumulator.Message,
			EntityID:     entity.ID,
			EntityTitle:  entity.Title,
			MatchSources: matchSources,
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Message.CreatedAt.Equal(results[j].Message.CreatedAt) {
			return results[i].Message.ID < results[j].Message.ID
		}
		return results[i].Message.CreatedAt.After(results[j].Message.CreatedAt)
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func GetEntityHoverSummary(database *gorm.DB, entityID string, channelID string, days int) (EntityHoverSummary, error) {
	entityID = strings.TrimSpace(entityID)
	channelID = strings.TrimSpace(channelID)
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	entity, err := GetEntity(database, entityID)
	if err != nil {
		return EntityHoverSummary{}, err
	}

	var refs []domain.KnowledgeEntityRef
	if err := database.Where("entity_id = ?", entityID).Order("created_at desc").Find(&refs).Error; err != nil {
		return EntityHoverSummary{}, err
	}

	summary := EntityHoverSummary{
		EntityID:         entity.ID,
		RecentWindowDays: days,
		RelatedChannels:  []EntityChannelSummary{},
	}
	if len(refs) == 0 {
		return summary, nil
	}

	recentWindowStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(days - 1))
	messageChannelMap, err := loadMessageChannelMap(database, refs)
	if err != nil {
		return summary, err
	}
	fileChannelMap, err := loadFileChannelMap(database, refs)
	if err != nil {
		return summary, err
	}
	channelNames, err := loadChannelNames(database, refs, messageChannelMap, fileChannelMap)
	if err != nil {
		return summary, err
	}

	type channelAggregate struct {
		ChannelID      string
		Name           string
		RefCount       int
		LastActivityAt time.Time
	}
	channelAggregates := map[string]*channelAggregate{}

	for _, ref := range refs {
		summary.RefCount++
		switch ref.RefKind {
		case "file":
			summary.FileRefCount++
		default:
			summary.MessageRefCount++
		}
		if summary.LastActivityAt == nil || ref.CreatedAt.After(*summary.LastActivityAt) {
			lastActivity := ref.CreatedAt
			summary.LastActivityAt = &lastActivity
		}
		if !ref.CreatedAt.Before(recentWindowStart) {
			summary.RecentRefCount++
		}

		resolvedChannelID := resolveKnowledgeRefChannelID(ref, messageChannelMap, fileChannelMap)
		if resolvedChannelID == "" {
			continue
		}
		aggregate := channelAggregates[resolvedChannelID]
		if aggregate == nil {
			aggregate = &channelAggregate{
				ChannelID: resolvedChannelID,
				Name:      channelNames[resolvedChannelID],
			}
			channelAggregates[resolvedChannelID] = aggregate
		}
		aggregate.RefCount++
		if ref.CreatedAt.After(aggregate.LastActivityAt) {
			aggregate.LastActivityAt = ref.CreatedAt
		}
		if resolvedChannelID == channelID {
			summary.ChannelRefCount++
		}
	}

	for _, aggregate := range channelAggregates {
		lastActivity := aggregate.LastActivityAt
		summary.RelatedChannels = append(summary.RelatedChannels, EntityChannelSummary{
			ChannelID:      aggregate.ChannelID,
			Name:           aggregate.Name,
			RefCount:       aggregate.RefCount,
			LastActivityAt: &lastActivity,
		})
	}
	sort.SliceStable(summary.RelatedChannels, func(i, j int) bool {
		if summary.RelatedChannels[i].RefCount == summary.RelatedChannels[j].RefCount {
			if summary.RelatedChannels[i].LastActivityAt != nil && summary.RelatedChannels[j].LastActivityAt != nil &&
				summary.RelatedChannels[i].LastActivityAt.Equal(*summary.RelatedChannels[j].LastActivityAt) {
				return summary.RelatedChannels[i].ChannelID < summary.RelatedChannels[j].ChannelID
			}
			if summary.RelatedChannels[i].LastActivityAt == nil {
				return false
			}
			if summary.RelatedChannels[j].LastActivityAt == nil {
				return true
			}
			return summary.RelatedChannels[i].LastActivityAt.After(*summary.RelatedChannels[j].LastActivityAt)
		}
		return summary.RelatedChannels[i].RefCount > summary.RelatedChannels[j].RefCount
	})

	return summary, nil
}

func BuildChannelKnowledgeDigest(database *gorm.DB, channelID string, window string, limit int) (ChannelKnowledgeDigest, error) {
	channelID = strings.TrimSpace(channelID)
	window, windowDays := normalizeDigestWindow(window)
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	digest := ChannelKnowledgeDigest{
		ChannelID:    channelID,
		Window:       window,
		WindowDays:   windowDays,
		GeneratedAt:  time.Now().UTC(),
		TopMovements: []ChannelKnowledgeDigestMovement{},
	}
	if channelID == "" {
		return digest, errors.New("channel_id is required")
	}

	var channel domain.Channel
	if err := database.Select("id, name").First(&channel, "id = ?", channelID).Error; err != nil {
		return digest, err
	}

	refs, err := listChannelKnowledgeEntityRefs(database, channelID)
	if err != nil {
		return digest, err
	}
	digest.TotalRefs = len(refs)
	if len(refs) == 0 {
		digest.Headline = fmt.Sprintf("No knowledge movement detected in #%s %s.", channel.Name, digestWindowPhrase(window))
		digest.Summary = "No entity references were created in the selected digest window."
		return digest, nil
	}

	recentWindowStart := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(windowDays - 1))
	previousWindowStart := recentWindowStart.AddDate(0, 0, -windowDays)
	type aggregate struct {
		EntityID         string
		RefCount         int
		RecentRefCount   int
		PreviousRefCount int
		LastActivityAt   time.Time
	}
	aggregates := map[string]*aggregate{}
	entityIDs := make([]string, 0)
	for _, ref := range refs {
		item := aggregates[ref.EntityID]
		if item == nil {
			item = &aggregate{EntityID: ref.EntityID}
			aggregates[ref.EntityID] = item
			entityIDs = append(entityIDs, ref.EntityID)
		}
		item.RefCount++
		if ref.CreatedAt.After(item.LastActivityAt) {
			item.LastActivityAt = ref.CreatedAt
		}
		if !ref.CreatedAt.Before(recentWindowStart) {
			item.RecentRefCount++
			digest.RecentRefCount++
		} else if !ref.CreatedAt.Before(previousWindowStart) {
			item.PreviousRefCount++
		}
	}

	entityMap, err := loadKnowledgeEntityMap(database, entityIDs)
	if err != nil {
		return digest, err
	}

	for _, entityID := range entityIDs {
		aggregate := aggregates[entityID]
		entity := entityMap[entityID]
		lastActivity := aggregate.LastActivityAt
		title := entity.Title
		if strings.TrimSpace(title) == "" {
			title = entityID
		}
		digest.TopMovements = append(digest.TopMovements, ChannelKnowledgeDigestMovement{
			EntityID:         entityID,
			EntityTitle:      title,
			EntityKind:       entity.Kind,
			RefCount:         aggregate.RefCount,
			RecentRefCount:   aggregate.RecentRefCount,
			PreviousRefCount: aggregate.PreviousRefCount,
			Delta:            aggregate.RecentRefCount - aggregate.PreviousRefCount,
			LastActivityAt:   &lastActivity,
		})
	}

	sort.SliceStable(digest.TopMovements, func(i, j int) bool {
		if digest.TopMovements[i].RecentRefCount == digest.TopMovements[j].RecentRefCount {
			if digest.TopMovements[i].Delta == digest.TopMovements[j].Delta {
				if digest.TopMovements[i].RefCount == digest.TopMovements[j].RefCount {
					if digest.TopMovements[i].LastActivityAt != nil && digest.TopMovements[j].LastActivityAt != nil &&
						digest.TopMovements[i].LastActivityAt.Equal(*digest.TopMovements[j].LastActivityAt) {
						return digest.TopMovements[i].EntityTitle < digest.TopMovements[j].EntityTitle
					}
					if digest.TopMovements[i].LastActivityAt == nil {
						return false
					}
					if digest.TopMovements[j].LastActivityAt == nil {
						return true
					}
					return digest.TopMovements[i].LastActivityAt.After(*digest.TopMovements[j].LastActivityAt)
				}
				return digest.TopMovements[i].RefCount > digest.TopMovements[j].RefCount
			}
			return digest.TopMovements[i].Delta > digest.TopMovements[j].Delta
		}
		return digest.TopMovements[i].RecentRefCount > digest.TopMovements[j].RecentRefCount
	})
	if len(digest.TopMovements) > limit {
		digest.TopMovements = digest.TopMovements[:limit]
	}

	top := digest.TopMovements[0]
	digest.Headline = fmt.Sprintf("%s leads knowledge movement in #%s %s.", top.EntityTitle, channel.Name, digestWindowPhrase(window))
	digest.Summary = fmt.Sprintf("%d entity references were created in #%s across %d highlighted movements.", digest.RecentRefCount, channel.Name, len(digest.TopMovements))
	return digest, nil
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

func loadMessageChannelMap(database *gorm.DB, refs []domain.KnowledgeEntityRef) (map[string]string, error) {
	messageIDs := make([]string, 0)
	for _, ref := range refs {
		if ref.RefKind == "message" {
			messageIDs = append(messageIDs, ref.RefID)
		}
	}

	messageChannelMap := map[string]string{}
	if len(messageIDs) == 0 {
		return messageChannelMap, nil
	}

	var messages []domain.Message
	if err := database.Select("id, channel_id").Where("id IN ?", messageIDs).Find(&messages).Error; err != nil {
		return nil, err
	}
	for _, message := range messages {
		messageChannelMap[message.ID] = message.ChannelID
	}
	return messageChannelMap, nil
}

func loadFileChannelMap(database *gorm.DB, refs []domain.KnowledgeEntityRef) (map[string]string, error) {
	fileIDs := make([]string, 0)
	for _, ref := range refs {
		if ref.RefKind == "file" {
			fileIDs = append(fileIDs, ref.RefID)
		}
	}

	fileChannelMap := map[string]string{}
	if len(fileIDs) == 0 {
		return fileChannelMap, nil
	}

	var files []domain.FileAsset
	if err := database.Select("id, channel_id").Where("id IN ?", fileIDs).Find(&files).Error; err != nil {
		return nil, err
	}
	for _, file := range files {
		fileChannelMap[file.ID] = file.ChannelID
	}
	return fileChannelMap, nil
}

func loadChannelNames(database *gorm.DB, refs []domain.KnowledgeEntityRef, messageChannelMap map[string]string, fileChannelMap map[string]string) (map[string]string, error) {
	channelIDs := make([]string, 0)
	seen := map[string]struct{}{}
	for _, ref := range refs {
		channelID := resolveKnowledgeRefChannelID(ref, messageChannelMap, fileChannelMap)
		if channelID == "" {
			continue
		}
		if _, exists := seen[channelID]; exists {
			continue
		}
		seen[channelID] = struct{}{}
		channelIDs = append(channelIDs, channelID)
	}

	channelNames := map[string]string{}
	if len(channelIDs) == 0 {
		return channelNames, nil
	}

	var channels []domain.Channel
	if err := database.Select("id, name").Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
		return nil, err
	}
	for _, channel := range channels {
		channelNames[channel.ID] = channel.Name
	}
	return channelNames, nil
}

func resolveKnowledgeRefChannelID(ref domain.KnowledgeEntityRef, messageChannelMap map[string]string, fileChannelMap map[string]string) string {
	switch ref.RefKind {
	case "file":
		return fileChannelMap[ref.RefID]
	default:
		return messageChannelMap[ref.RefID]
	}
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

func normalizeDigestWindow(window string) (string, int) {
	switch strings.ToLower(strings.TrimSpace(window)) {
	case "daily", "day", "1d":
		return "daily", 1
	case "monthly", "month", "30d":
		return "monthly", 30
	default:
		return "weekly", 7
	}
}

func digestWindowPhrase(window string) string {
	switch window {
	case "daily":
		return "today"
	case "monthly":
		return "this month"
	default:
		return "this week"
	}
}

func buildChannelKnowledgeVelocity(recentWindowDays, recentCount, previousCount int) ChannelKnowledgeVelocity {
	velocity := ChannelKnowledgeVelocity{
		RecentWindowDays: recentWindowDays,
		PreviousRefCount: previousCount,
		RecentRefCount:   recentCount,
		Delta:            recentCount - previousCount,
	}
	switch {
	case recentCount >= 2 && previousCount == 0:
		velocity.IsSpiking = true
	case previousCount > 0 && recentCount >= previousCount*2:
		velocity.IsSpiking = true
	}
	return velocity
}

func normalizeMentionContent(content string) string {
	plain := strings.TrimSpace(content)
	if plain == "" {
		return ""
	}
	tagPattern := regexp.MustCompile(`<[^>]+>`)
	plain = tagPattern.ReplaceAllString(plain, " ")
	plain = html.UnescapeString(plain)
	return strings.Join(strings.Fields(plain), " ")
}

func findExplicitMentionRange(lowerContent, lowerMention string, occupied map[int]struct{}) (int, int, bool) {
	contentRunes := []rune(lowerContent)
	mentionRunes := []rune(lowerMention)
	if len(mentionRunes) == 0 || len(contentRunes) < len(mentionRunes) {
		return 0, 0, false
	}

	for idx := 0; idx <= len(contentRunes)-len(mentionRunes); idx++ {
		if string(contentRunes[idx:idx+len(mentionRunes)]) != lowerMention {
			continue
		}
		if hasOccupiedRange(idx, idx+len(mentionRunes), occupied) {
			continue
		}
		if idx > 0 {
			prev := contentRunes[idx-1]
			if unicode.IsLetter(prev) || unicode.IsDigit(prev) {
				continue
			}
		}
		end := idx + len(mentionRunes)
		if end < len(contentRunes) {
			next := contentRunes[end]
			if unicode.IsLetter(next) || unicode.IsDigit(next) {
				continue
			}
		}
		return idx, end, true
	}

	return 0, 0, false
}

func hasOccupiedRange(start, end int, occupied map[int]struct{}) bool {
	for idx := start; idx < end; idx++ {
		if _, exists := occupied[idx]; exists {
			return true
		}
	}
	return false
}

func spanOccupied(occupied []bool, start, end int) bool {
	if start < 0 || end > len(occupied) || start >= end {
		return true
	}
	for idx := start; idx < end; idx++ {
		if occupied[idx] {
			return true
		}
	}
	return false
}

func isEntityTextBoundary(text string, start, end int) bool {
	if start > 0 && isEntityBoundaryRune(rune(text[start-1])) {
		return false
	}
	if end < len(text) && isEntityBoundaryRune(rune(text[end])) {
		return false
	}
	return true
}

func isEntityBoundaryRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func normalizeFollowNotificationLevel(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "all":
		return "all"
	case "digest_only":
		return "digest_only"
	case "silent":
		return "silent"
	default:
		return ""
	}
}

func normalizeWorkspaceKnowledgeSettings(workspace domain.Workspace) WorkspaceKnowledgeSettings {
	threshold := workspace.KnowledgeSpikeThreshold
	if threshold <= 0 {
		threshold = 3
	}
	cooldown := workspace.KnowledgeSpikeCooldownMins
	if cooldown <= 0 {
		cooldown = 360
	}
	return WorkspaceKnowledgeSettings{
		WorkspaceID:          workspace.ID,
		SpikeThreshold:       threshold,
		SpikeCooldownMinutes: cooldown,
	}
}

func getWorkspaceKnowledgeSettingsOrDefault(database *gorm.DB, workspaceID string) (WorkspaceKnowledgeSettings, error) {
	workspaceID = strings.TrimSpace(workspaceID)
	if workspaceID == "" {
		return WorkspaceKnowledgeSettings{
			SpikeThreshold:       3,
			SpikeCooldownMinutes: 360,
		}, nil
	}

	var workspace domain.Workspace
	if err := database.First(&workspace, "id = ?", workspaceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return WorkspaceKnowledgeSettings{
				WorkspaceID:          workspaceID,
				SpikeThreshold:       3,
				SpikeCooldownMinutes: 360,
			}, nil
		}
		return WorkspaceKnowledgeSettings{}, err
	}
	return normalizeWorkspaceKnowledgeSettings(workspace), nil
}

func truncateToUTCDate(ts time.Time) time.Time {
	ts = ts.UTC()
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 0, 0, 0, 0, time.UTC)
}

func listEntityRelatedChannelIDs(database *gorm.DB, entityID string) ([]string, error) {
	var messageChannelIDs []string
	if err := database.Model(&domain.Message{}).
		Select("DISTINCT messages.channel_id").
		Joins("JOIN knowledge_entity_refs ON knowledge_entity_refs.ref_id = messages.id AND knowledge_entity_refs.ref_kind = ?", "message").
		Where("knowledge_entity_refs.entity_id = ?", entityID).
		Order("messages.channel_id asc").
		Pluck("messages.channel_id", &messageChannelIDs).Error; err != nil {
		return nil, err
	}

	var fileChannelIDs []string
	if err := database.Model(&domain.FileAsset{}).
		Select("DISTINCT file_assets.channel_id").
		Joins("JOIN knowledge_entity_refs ON knowledge_entity_refs.ref_id = file_assets.id AND knowledge_entity_refs.ref_kind = ?", "file").
		Where("knowledge_entity_refs.entity_id = ?", entityID).
		Order("file_assets.channel_id asc").
		Pluck("file_assets.channel_id", &fileChannelIDs).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(messageChannelIDs)+len(fileChannelIDs))
	related := make([]string, 0, len(messageChannelIDs)+len(fileChannelIDs))
	for _, channelID := range append(messageChannelIDs, fileChannelIDs...) {
		channelID = strings.TrimSpace(channelID)
		if channelID == "" {
			continue
		}
		if _, exists := seen[channelID]; exists {
			continue
		}
		seen[channelID] = struct{}{}
		related = append(related, channelID)
	}
	sort.Strings(related)
	return related, nil
}
