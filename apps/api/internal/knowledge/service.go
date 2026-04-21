package knowledge

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
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
			return nil
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
	return nil
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
