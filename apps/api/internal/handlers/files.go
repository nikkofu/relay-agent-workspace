package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/domain"
)

type fileAssetResponse struct {
	domain.FileAsset
	Uploader *domain.User `json:"uploader,omitempty"`
	URL      string       `json:"url"`
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

	fileID := "file-" + time.Now().UTC().Format("20060102150405.000000")
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
	}
	if err := db.DB.Create(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file asset"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func ListFiles(c *gin.Context) {
	var files []domain.FileAsset

	query := db.DB.Order("created_at desc")
	if channelID := c.Query("channel_id"); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
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

	if err := db.DB.Save(&asset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update file archive state"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"file": hydrateFileAssetResponse(asset)})
}

func GetArchivedFiles(c *gin.Context) {
	var files []domain.FileAsset
	query := db.DB.Where("is_archived = ?", true).Order("archived_at desc, created_at desc")
	if channelID := c.Query("channel_id"); channelID != "" {
		query = query.Where("channel_id = ?", channelID)
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

func hydrateFileAssetResponse(asset domain.FileAsset) fileAssetResponse {
	response := fileAssetResponse{
		FileAsset: asset,
		URL:       "/api/v1/files/" + asset.ID + "/content",
	}

	var uploader domain.User
	if err := db.DB.First(&uploader, "id = ?", asset.UploaderID).Error; err == nil {
		enriched := enrichUser(uploader)
		response.Uploader = &enriched
	}

	if strings.TrimSpace(response.ContentType) == "" {
		response.ContentType = "application/octet-stream"
	}

	return response
}
