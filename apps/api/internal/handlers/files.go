package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
	"github.com/nikkofu/relay-agent-workspace/api/internal/fileindex"
	"github.com/nikkofu/relay-agent-workspace/api/internal/ids"
	"github.com/nikkofu/relay-agent-workspace/api/internal/realtime"
)

type fileAssetResponse struct {
	domain.FileAsset
	Uploader       *domain.User `json:"uploader,omitempty"`
	URL            string       `json:"url"`
	PreviewURL     string       `json:"preview_url,omitempty"`
	PreviewKind    string       `json:"preview_kind,omitempty"`
	Type           string       `json:"type"`
	Size           int64        `json:"size"`
	UserID         string       `json:"userId"`
	ChannelIDAlias string       `json:"channelId,omitempty"`
	CreatedAtAlias time.Time    `json:"createdAt"`
	CommentCount   int64        `json:"comment_count"`
	ShareCount     int64        `json:"share_count"`
	Starred        bool         `json:"starred"`
	Tags           []string     `json:"tags"`
	IsSearchable   bool         `json:"is_searchable"`
	IsCitable      bool         `json:"is_citable"`
}

type filePreviewResponse struct {
	FileID        string       `json:"file_id"`
	Name          string       `json:"name"`
	ContentType   string       `json:"content_type"`
	PreviewKind   string       `json:"preview_kind"`
	PreviewURL    string       `json:"preview_url,omitempty"`
	DownloadURL   string       `json:"download_url"`
	IsPreviewable bool         `json:"is_previewable"`
	Size          int64        `json:"size"`
	ChannelID     string       `json:"channel_id,omitempty"`
	Uploader      *domain.User `json:"uploader,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
}

type fileCommentResponse struct {
	domain.FileComment
	User *domain.User `json:"user,omitempty"`
}

type fileShareResponse struct {
	domain.FileShare
	Actor   *domain.User    `json:"actor,omitempty"`
	Message *domain.Message `json:"message,omitempty"`
}

func UploadFile(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer src.Close()

	fileID := ids.NewPrefixedUUID("file")
	ext := filepath.Ext(fileHeader.Filename)
	filename := fileID + ext
	storageDir := filepath.Join("uploads")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload storage"})
		return
	}

	dstPath := filepath.Join(storageDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload file"})
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist upload"})
		return
	}

	asset := domain.FileAsset{
		ID:          fileID,
		ChannelID:   c.PostForm("channel_id"),
		UploaderID:  currentUser.ID,
		Name:        fileHeader.Filename,
		StoragePath: filename,
		ContentType: fileHeader.Header.Get("Content-Type"),
		SizeBytes:   written,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := db.DB.Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file asset"})
		return
	}
	extraction, err := rebuildFileExtraction(asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build file extraction"})
		return
	}
	_ = db.DB.First(&asset, "id = ?", asset.ID).Error
	_ = broadcastFileExtractionEvent(asset, extraction)
	recordFileEvent(asset.ID, currentUser.ID, "uploaded", "Uploaded "+asset.Name)

	c.JSON(http.StatusCreated, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func RebuildFileExtraction(c *gin.Context) {
	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	extraction, err := rebuildFileExtraction(asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rebuild file extraction"})
		return
	}
	_ = db.DB.First(&asset, "id = ?", asset.ID).Error
	_ = broadcastFileExtractionEvent(asset, extraction)

	c.JSON(http.StatusOK, gin.H{
		"file":       hydrateFileAssetResponse(asset),
		"extraction": extraction,
	})
}

func ListFiles(c *gin.Context) {
	var files []domain.FileAsset

	query := db.DB.Order("created_at desc")
	if channelID := c.Query("channel_id"); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}
	if uploaderID := c.Query("uploader_id"); uploaderID != "" {
		query = query.Where("uploader_id = ?", uploaderID)
	}
	if contentType := strings.TrimSpace(c.Query("content_type")); contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}
	if archived := strings.TrimSpace(c.Query("is_archived")); archived != "" {
		query = query.Where("is_archived = ?", archived == "true")
	}
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load files"})
		return
	}

	items := make([]fileAssetResponse, 0, len(files))
	for _, file := range files {
		items = append(items, hydrateFileAssetResponse(file))
	}

	c.JSON(http.StatusOK, gin.H{"files": items})
}

func GetFile(c *gin.Context) {
	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func GetFileExtraction(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var extraction domain.FileExtraction
	if err := db.DB.First(&extraction, "file_id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file extraction not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"extraction": extraction})
}

func GetFileExtractedContent(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var extraction domain.FileExtraction
	if err := db.DB.First(&extraction, "file_id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file extraction not found"})
		return
	}

	var chunkCount int64
	db.DB.Model(&domain.FileExtractionChunk{}).Where("file_id = ?", c.Param("id")).Count(&chunkCount)

	c.JSON(http.StatusOK, gin.H{
		"file_id":         c.Param("id"),
		"status":          extraction.Status,
		"extractor":       extraction.Extractor,
		"content_text":    extraction.ContentText,
		"content_summary": extraction.ContentSummary,
		"chunks_count":    chunkCount,
	})
}

func GetFileExtractionChunks(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var chunks []domain.FileExtractionChunk
	if err := db.DB.Where("file_id = ?", c.Param("id")).Order("chunk_index asc, id asc").Find(&chunks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load extraction chunks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id": c.Param("id"),
		"chunks":  chunks,
	})
}

func SearchFiles(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q is required"})
		return
	}

	type searchResult struct {
		File         fileAssetResponse `json:"file"`
		Snippet      string            `json:"snippet"`
		LocatorType  string            `json:"locator_type"`
		LocatorValue string            `json:"locator_value"`
		Heading      string            `json:"heading"`
		MatchReason  string            `json:"match_reason"`
	}

	needle := strings.ToLower(q)
	var assets []domain.FileAsset
	if err := db.DB.Where("extraction_status = ?", "ready").Order("created_at desc").Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search files"})
		return
	}

	results := make([]searchResult, 0)
	for _, asset := range assets {
		var chunks []domain.FileExtractionChunk
		if err := db.DB.Where("file_id = ?", asset.ID).Order("chunk_index asc").Find(&chunks).Error; err != nil {
			continue
		}

		matchReason := ""
		snippet := ""
		locatorType := ""
		locatorValue := ""
		heading := ""

		if strings.Contains(strings.ToLower(asset.Name), needle) {
			matchReason = "file_name"
			snippet = asset.Name
		}
		for _, chunk := range chunks {
			if !strings.Contains(strings.ToLower(chunk.Text), needle) {
				continue
			}
			if matchReason == "" {
				matchReason = "content"
				snippet = chunk.Text
				locatorType = chunk.LocatorType
				locatorValue = chunk.LocatorValue
				heading = chunk.Heading
			}
			break
		}
		if matchReason == "" {
			continue
		}

		results = append(results, searchResult{
			File:         hydrateFileAssetResponse(asset),
			Snippet:      snippet,
			LocatorType:  locatorType,
			LocatorValue: locatorValue,
			Heading:      heading,
			MatchReason:  matchReason,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   q,
		"results": results,
	})
}

func GetFileCitations(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var chunks []domain.FileExtractionChunk
	if err := db.DB.Where("file_id = ?", c.Param("id")).Order("chunk_index asc, id asc").Find(&chunks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load file citations"})
		return
	}

	type citationResponse struct {
		ChunkID      uint   `json:"chunk_id"`
		Text         string `json:"text"`
		LocatorType  string `json:"locator_type"`
		LocatorValue string `json:"locator_value"`
		Heading      string `json:"heading"`
	}

	citations := make([]citationResponse, 0, len(chunks))
	for _, chunk := range chunks {
		citations = append(citations, citationResponse{
			ChunkID:      chunk.ID,
			Text:         chunk.Text,
			LocatorType:  chunk.LocatorType,
			LocatorValue: chunk.LocatorValue,
			Heading:      chunk.Heading,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"file_id":   c.Param("id"),
		"citations": citations,
	})
}

func GetFileComments(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var comments []domain.FileComment
	if err := db.DB.Where("file_id = ?", c.Param("id")).Order("created_at asc").Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load file comments"})
		return
	}

	items := make([]fileCommentResponse, 0, len(comments))
	for _, comment := range comments {
		item := fileCommentResponse{FileComment: comment}
		var user domain.User
		if err := db.DB.First(&user, "id = ?", comment.UserID).Error; err == nil {
			enriched := enrichUser(user)
			item.User = &enriched
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"comments": items})
}

func CreateFileComment(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	comment := domain.FileComment{
		ID:        ids.NewPrefixedUUID("fcomment"),
		FileID:    asset.ID,
		UserID:    currentUser.ID,
		Content:   strings.TrimSpace(input.Content),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if comment.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}
	if err := db.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file comment"})
		return
	}

	recordFileEvent(asset.ID, currentUser.ID, "commented", "Commented on "+asset.Name)
	c.JSON(http.StatusCreated, gin.H{"comment": fileCommentResponse{FileComment: comment, User: ptrUser(enrichUser(currentUser))}})
}

func UpdateFileKnowledge(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var input struct {
		KnowledgeState string   `json:"knowledge_state"`
		SourceKind     string   `json:"source_kind"`
		Summary        string   `json:"summary"`
		Tags           []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset.KnowledgeState = strings.TrimSpace(input.KnowledgeState)
	asset.SourceKind = strings.TrimSpace(input.SourceKind)
	asset.Summary = strings.TrimSpace(input.Summary)
	asset.Tags = encodeTags(input.Tags)
	asset.UpdatedAt = time.Now().UTC()
	if err := db.DB.Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file knowledge"})
		return
	}

	recordFileEvent(asset.ID, currentUser.ID, "knowledge.updated", "Updated knowledge metadata for "+asset.Name)
	c.JSON(http.StatusOK, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func ToggleFileStar(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var star domain.StarredFile
	err = db.DB.First(&star, "file_id = ? AND user_id = ?", asset.ID, currentUser.ID).Error
	if err == nil {
		if err := db.DB.Delete(&star).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unstar file"})
			return
		}
		recordFileEvent(asset.ID, currentUser.ID, "unstarred", "Removed star from "+asset.Name)
		c.JSON(http.StatusOK, gin.H{"starred": false, "file": hydrateFileAssetResponse(asset)})
		return
	}

	now := time.Now().UTC()
	star = domain.StarredFile{FileID: asset.ID, UserID: currentUser.ID, CreatedAt: now}
	if err := db.DB.Create(&star).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to star file"})
		return
	}
	recordFileEvent(asset.ID, currentUser.ID, "starred", "Starred "+asset.Name)
	c.JSON(http.StatusOK, gin.H{"starred": true, "file": hydrateFileAssetResponse(asset)})
}

func GetStarredFiles(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var stars []domain.StarredFile
	if err := db.DB.Where("user_id = ?", currentUser.ID).Order("created_at desc").Find(&stars).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load starred files"})
		return
	}

	items := make([]fileAssetResponse, 0, len(stars))
	for _, star := range stars {
		var asset domain.FileAsset
		if err := db.DB.First(&asset, "id = ?", star.FileID).Error; err == nil {
			items = append(items, hydrateFileAssetResponse(asset))
		}
	}

	c.JSON(http.StatusOK, gin.H{"files": items})
}

func ShareFile(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var input struct {
		ChannelID string `json:"channel_id" binding:"required"`
		ThreadID  string `json:"thread_id"`
		Comment   string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var channel domain.Channel
	if err := db.DB.First(&channel, "id = ?", input.ChannelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	now := time.Now().UTC()
	message := domain.Message{
		ID:        ids.NewPrefixedUUID("msg"),
		ChannelID: input.ChannelID,
		UserID:    currentUser.ID,
		Content:   strings.TrimSpace(input.Comment),
		ThreadID:  strings.TrimSpace(input.ThreadID),
		CreatedAt: now,
	}
	if err := db.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file share message"})
		return
	}

	attachment := domain.MessageFileAttachment{
		MessageID: message.ID,
		FileID:    asset.ID,
		CreatedAt: now,
	}
	if err := db.DB.Create(&attachment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to attach shared file"})
		return
	}

	refreshed, err := refreshMessageMetadata(message.ID)
	if err == nil && refreshed != nil {
		message = *refreshed
	}

	share := domain.FileShare{
		ID:        ids.NewPrefixedUUID("fshare"),
		FileID:    asset.ID,
		ChannelID: input.ChannelID,
		ThreadID:  strings.TrimSpace(input.ThreadID),
		MessageID: message.ID,
		SharedBy:  currentUser.ID,
		Comment:   message.Content,
		CreatedAt: now,
	}
	if err := db.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist file share"})
		return
	}

	recordFileEvent(asset.ID, currentUser.ID, "shared", "Shared "+asset.Name+" in #"+channel.Name)
	if RealtimeHub != nil {
		_ = broadcastRealtimeEvent("message.created", message, message)
	}

	c.JSON(http.StatusCreated, gin.H{"share": hydrateFileShareResponse(share, &message)})
}

func GetFileShares(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var shares []domain.FileShare
	if err := db.DB.Where("file_id = ?", c.Param("id")).Order("created_at desc").Find(&shares).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load file shares"})
		return
	}

	items := make([]fileShareResponse, 0, len(shares))
	for _, share := range shares {
		var message *domain.Message
		var msg domain.Message
		if share.MessageID != "" && db.DB.First(&msg, "id = ?", share.MessageID).Error == nil {
			message = &msg
		}
		items = append(items, hydrateFileShareResponse(share, message))
	}

	c.JSON(http.StatusOK, gin.H{"shares": items})
}

func GetFilePreview(c *gin.Context) {
	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preview": buildFilePreview(asset)})
}

func GetFileContent(c *gin.Context) {
	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	path := filepath.Join("uploads", asset.StoragePath)
	c.File(path)
}

func ToggleFileArchive(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var input struct {
		IsArchived bool `json:"is_archived"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	asset.IsArchived = input.IsArchived
	if input.IsArchived {
		now := time.Now().UTC()
		asset.ArchivedAt = &now
	} else {
		asset.ArchivedAt = nil
	}
	asset.UpdatedAt = time.Now().UTC()

	if err := db.DB.Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file archive state"})
		return
	}
	action := "restored"
	detail := "Restored " + asset.Name
	if asset.IsArchived {
		action = "archived"
		detail = "Archived " + asset.Name
	}
	recordFileEvent(asset.ID, currentUser.ID, action, detail)

	c.JSON(http.StatusOK, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func GetArchivedFiles(c *gin.Context) {
	var files []domain.FileAsset
	query := db.DB.Where("is_archived = ?", true).Order("archived_at desc, created_at desc")
	if channelID := c.Query("channel_id"); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
	}
	if uploaderID := c.Query("uploader_id"); uploaderID != "" {
		query = query.Where("uploader_id = ?", uploaderID)
	}
	if contentType := strings.TrimSpace(c.Query("content_type")); contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(q)+"%")
	}
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load archived files"})
		return
	}

	items := make([]fileAssetResponse, 0, len(files))
	for _, file := range files {
		items = append(items, hydrateFileAssetResponse(file))
	}

	c.JSON(http.StatusOK, gin.H{"files": items})
}

func DeleteFile(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	recordFileEvent(asset.ID, currentUser.ID, "deleted", "Deleted "+asset.Name)
	if err := db.DB.Delete(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "file_id": asset.ID})
}

func UpdateFileRetention(c *gin.Context) {
	currentUser, err := getCurrentUser()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var asset domain.FileAsset
	if err := db.DB.First(&asset, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var input struct {
		RetentionDays int `json:"retention_days"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.RetentionDays < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "retention_days must be >= 0"})
		return
	}

	asset.RetentionDays = input.RetentionDays
	if input.RetentionDays == 0 {
		asset.ExpiresAt = nil
	} else {
		expiresAt := time.Now().UTC().Add(time.Duration(input.RetentionDays) * 24 * time.Hour)
		asset.ExpiresAt = &expiresAt
	}
	asset.UpdatedAt = time.Now().UTC()
	if err := db.DB.Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file retention"})
		return
	}
	recordFileEvent(asset.ID, currentUser.ID, "retention.updated", "Retention set to "+strconv.Itoa(input.RetentionDays)+" days")

	c.JSON(http.StatusOK, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func GetFileAudit(c *gin.Context) {
	if err := db.DB.First(&domain.FileAsset{}, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	var events []domain.FileAssetEvent
	if err := db.DB.Where("file_id = ?", c.Param("id")).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load file audit"})
		return
	}

	type fileAuditEventResponse struct {
		ID         uint         `json:"id"`
		FileID     string       `json:"fileId"`
		UserID     string       `json:"userId"`
		Action     string       `json:"action"`
		OccurredAt time.Time    `json:"occurredAt"`
		Metadata   any          `json:"metadata,omitempty"`
		User       *domain.User `json:"user,omitempty"`
		Detail     string       `json:"detail"`
		CreatedAt  time.Time    `json:"created_at"`
		Actor      *domain.User `json:"actor,omitempty"`
	}

	items := make([]fileAuditEventResponse, 0, len(events))
	for _, event := range events {
		item := fileAuditEventResponse{
			ID:         event.ID,
			FileID:     event.FileID,
			UserID:     event.ActorID,
			Action:     normalizeAuditAction(event.Action),
			OccurredAt: event.CreatedAt,
			Metadata: gin.H{
				"detail": event.Detail,
			},
			Detail:    event.Detail,
			CreatedAt: event.CreatedAt,
		}
		if event.ActorID != "" {
			var actor domain.User
			if err := db.DB.First(&actor, "id = ?", event.ActorID).Error; err == nil {
				enriched := enrichUser(actor)
				item.Actor = &enriched
				item.User = &enriched
			}
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"events": items, "audit_history": items})
}

func hydrateFileAssetResponse(asset domain.FileAsset) fileAssetResponse {
	response := fileAssetResponse{
		FileAsset:      asset,
		URL:            "/api/v1/files/" + asset.ID + "/content",
		Type:           asset.ContentType,
		Size:           asset.SizeBytes,
		UserID:         asset.UploaderID,
		ChannelIDAlias: asset.ChannelID,
		CreatedAtAlias: asset.CreatedAt,
		Tags:           decodeTags(asset.Tags),
		IsSearchable:   asset.ExtractionStatus == "ready",
		IsCitable:      asset.ExtractionStatus == "ready",
	}

	var uploader domain.User
	if err := db.DB.First(&uploader, "id = ?", asset.UploaderID).Error; err == nil {
		enriched := enrichUser(uploader)
		response.Uploader = &enriched
	}

	if strings.TrimSpace(response.ContentType) == "" {
		response.ContentType = "application/octet-stream"
	}
	if strings.HasPrefix(strings.ToLower(response.ContentType), "image/") {
		response.PreviewKind = "image"
		response.PreviewURL = response.URL
	} else if response.ContentType == "application/pdf" {
		response.PreviewKind = "pdf"
		response.PreviewURL = response.URL
	} else {
		response.PreviewKind = "file"
	}

	db.DB.Model(&domain.FileComment{}).Where("file_id = ?", asset.ID).Count(&response.CommentCount)
	db.DB.Model(&domain.FileShare{}).Where("file_id = ?", asset.ID).Count(&response.ShareCount)
	if currentUser, err := getCurrentUser(); err == nil {
		var count int64
		db.DB.Model(&domain.StarredFile{}).Where("file_id = ? AND user_id = ?", asset.ID, currentUser.ID).Count(&count)
		response.Starred = count > 0
	}

	return response
}

func buildFilePreview(asset domain.FileAsset) filePreviewResponse {
	file := hydrateFileAssetResponse(asset)
	isPreviewable := file.PreviewKind == "image" || file.PreviewKind == "pdf"
	return filePreviewResponse{
		FileID:        asset.ID,
		Name:          asset.Name,
		ContentType:   file.ContentType,
		PreviewKind:   file.PreviewKind,
		PreviewURL:    file.PreviewURL,
		DownloadURL:   file.URL,
		IsPreviewable: isPreviewable,
		Size:          asset.SizeBytes,
		ChannelID:     asset.ChannelID,
		Uploader:      file.Uploader,
		CreatedAt:     asset.CreatedAt,
		ExpiresAt:     asset.ExpiresAt,
	}
}

func normalizeAuditAction(action string) string {
	switch action {
	case "uploaded":
		return "upload"
	case "archived":
		return "archive"
	case "restored":
		return "restore"
	case "deleted":
		return "delete"
	case "retention.updated":
		return "retention_update"
	default:
		return action
	}
}

func recordFileEvent(fileID, actorID, action, detail string) {
	if fileID == "" || action == "" {
		return
	}
	_ = db.DB.Create(&domain.FileAssetEvent{
		FileID:    fileID,
		ActorID:   actorID,
		Action:    action,
		Detail:    detail,
		CreatedAt: time.Now().UTC(),
	}).Error
}

func broadcastFileExtractionEvent(asset domain.FileAsset, extraction domain.FileExtraction) error {
	if RealtimeHub == nil {
		return nil
	}

	workspaceID := ""
	if asset.ChannelID != "" {
		var channel domain.Channel
		if err := db.DB.First(&channel, "id = ?", asset.ChannelID).Error; err == nil {
			workspaceID = channel.WorkspaceID
		}
	}

	return RealtimeHub.Broadcast(realtime.Event{
		ID:          "evt_" + time.Now().Format("20060102150405.000000"),
		Type:        "file.extraction.updated",
		WorkspaceID: workspaceID,
		ChannelID:   asset.ChannelID,
		EntityID:    asset.ID,
		TS:          time.Now().UTC().Format(time.RFC3339Nano),
		Payload: gin.H{
			"file_id":         asset.ID,
			"status":          extraction.Status,
			"content_summary": extraction.ContentSummary,
			"needs_ocr":       extraction.NeedsOCR,
			"ocr_provider":    extraction.OCRProvider,
			"ocr_is_mock":     extraction.OCRIsMock,
			"is_searchable":   extraction.Status == "ready",
			"is_citable":      extraction.Status == "ready",
		},
	})
}

func rebuildFileExtraction(asset domain.FileAsset) (domain.FileExtraction, error) {
	extraction, err := ensureFileExtraction(asset)
	if err != nil {
		return domain.FileExtraction{}, err
	}

	path := filepath.Join("uploads", asset.StoragePath)
	now := time.Now().UTC()
	extraction.Status = "processing"
	extraction.StartedAt = &now
	extraction.ErrorCode = ""
	extraction.ErrorMessage = ""
	extraction.UpdatedAt = now
	if extraction.CreatedAt.IsZero() {
		extraction.CreatedAt = now
	}

	if extraction.ID == "" {
		extraction.ID = ids.NewPrefixedUUID("fextract")
		if err := db.DB.Create(&extraction).Error; err != nil {
			return domain.FileExtraction{}, err
		}
	} else if err := db.DB.Save(&extraction).Error; err != nil {
		return domain.FileExtraction{}, err
	}

	result := fileindex.NewService(fileindex.MockOCRProvider{}).ExtractFile(path, asset)
	return persistExtractionResult(asset, extraction, result)
}

func ensureFileExtraction(asset domain.FileAsset) (domain.FileExtraction, error) {
	var extraction domain.FileExtraction
	err := db.DB.First(&extraction, "file_id = ?", asset.ID).Error
	switch {
	case err == nil:
		return extraction, nil
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		return domain.FileExtraction{}, err
	}

	now := time.Now().UTC()
	extraction = domain.FileExtraction{
		ID:        ids.NewPrefixedUUID("fextract"),
		FileID:    asset.ID,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.DB.Create(&extraction).Error; err != nil {
		return domain.FileExtraction{}, err
	}
	return extraction, nil
}

func persistExtractionResult(asset domain.FileAsset, extraction domain.FileExtraction, result fileindex.ExtractionResult) (domain.FileExtraction, error) {
	completedAt := time.Now().UTC()
	extraction.Status = result.Status
	extraction.Extractor = result.Extractor
	extraction.ContentText = result.ContentText
	extraction.ContentSummary = result.ContentSummary
	extraction.ErrorCode = result.ErrorCode
	extraction.ErrorMessage = result.ErrorMessage
	extraction.NeedsOCR = result.NeedsOCR
	extraction.OCRProvider = result.OCRProvider
	extraction.OCRIsMock = result.OCRIsMock
	extraction.CompletedAt = &completedAt
	extraction.UpdatedAt = completedAt

	if err := db.DB.Save(&extraction).Error; err != nil {
		return domain.FileExtraction{}, err
	}

	if err := db.DB.Where("file_id = ?", asset.ID).Delete(&domain.FileExtractionChunk{}).Error; err != nil {
		return domain.FileExtraction{}, err
	}
	for idx, chunk := range result.Chunks {
		row := domain.FileExtractionChunk{
			ExtractionID:  extraction.ID,
			FileID:        asset.ID,
			ChunkIndex:    idx,
			Text:          chunk.Text,
			TokenEstimate: chunk.TokenEstimate,
			LocatorType:   chunk.LocatorType,
			LocatorValue:  chunk.LocatorValue,
			Heading:       chunk.Heading,
			CreatedAt:     completedAt,
		}
		if err := db.DB.Create(&row).Error; err != nil {
			return domain.FileExtraction{}, err
		}
	}

	updates := map[string]any{
		"extraction_status": extraction.Status,
		"content_summary":   extraction.ContentSummary,
		"needs_ocr":         extraction.NeedsOCR,
		"ocr_provider":      extraction.OCRProvider,
		"ocr_is_mock":       extraction.OCRIsMock,
		"updated_at":        completedAt,
	}
	if extraction.Status == "ready" {
		updates["last_indexed_at"] = &completedAt
	} else {
		updates["last_indexed_at"] = nil
	}
	if err := db.DB.Model(&domain.FileAsset{}).Where("id = ?", asset.ID).Updates(updates).Error; err != nil {
		return domain.FileExtraction{}, err
	}

	return extraction, nil
}

func hydrateFileShareResponse(share domain.FileShare, message *domain.Message) fileShareResponse {
	response := fileShareResponse{FileShare: share, Message: message}
	var actor domain.User
	if err := db.DB.First(&actor, "id = ?", share.SharedBy).Error; err == nil {
		enriched := enrichUser(actor)
		response.Actor = &enriched
	}
	return response
}

func encodeTags(tags []string) string {
	items := make([]string, 0, len(tags))
	for _, tag := range tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	if len(items) == 0 {
		return ""
	}
	raw, _ := json.Marshal(items)
	return string(raw)
}

func decodeTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err == nil {
		return tags
	}
	items := strings.Split(raw, ",")
	tags = make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	return tags
}

func ptrUser(user domain.User) *domain.User {
	return &user
}
