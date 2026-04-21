package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nikkofu/relay-agent-workspace/api/internal/db"
	"github.com/nikkofu/relay-agent-workspace/api/internal/knowledge"
)

func ListKnowledgeEntities(c *gin.Context) {
	entities, err := knowledge.ListEntities(db.DB, c.Query("workspace_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge entities"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"entities": entities})
}

func CreateKnowledgeEntity(c *gin.Context) {
	var input knowledge.CreateEntityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := knowledge.CreateEntity(db.DB, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"entity": entity})
}

func GetKnowledgeEntity(c *gin.Context) {
	entity, err := knowledge.GetEntity(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func UpdateKnowledgeEntity(c *gin.Context) {
	var input knowledge.UpdateEntityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entity, err := knowledge.UpdateEntity(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"entity": entity})
}

func ListKnowledgeEntityRefs(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	refs, err := knowledge.ListEntityRefs(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge refs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"refs": refs})
}

func AddKnowledgeEntityRef(c *gin.Context) {
	var input knowledge.AddEntityRefInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ref, err := knowledge.AddEntityRef(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge ref")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"ref": ref})
}

func ListKnowledgeEntityTimeline(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	events, err := knowledge.ListEntityTimeline(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge timeline"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func AddKnowledgeEntityEvent(c *gin.Context) {
	var input knowledge.AddEntityEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if currentUser, err := getCurrentUser(); err == nil && strings.TrimSpace(input.ActorUserID) == "" {
		input.ActorUserID = currentUser.ID
	}
	event, err := knowledge.AppendEntityEvent(db.DB, c.Param("id"), input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge event")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func ListKnowledgeEntityLinks(c *gin.Context) {
	if _, err := knowledge.GetEntity(db.DB, c.Param("id")); err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	links, err := knowledge.ListEntityLinks(db.DB, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load knowledge links"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"links": links})
}

func AddKnowledgeEntityLink(c *gin.Context) {
	var input knowledge.AddEntityLinkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	link, err := knowledge.AddEntityLink(db.DB, input)
	if err != nil {
		handleKnowledgeError(c, err, "failed to add knowledge link")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"link": link})
}

func GetKnowledgeEntityGraph(c *gin.Context) {
	graph, err := knowledge.BuildEntityGraph(db.DB, c.Param("id"))
	if err != nil {
		handleKnowledgeNotFound(c, err, "knowledge entity not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"graph": graph})
}

func handleKnowledgeError(c *gin.Context, err error, fallback string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "knowledge entity not found"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": fallback})
}

func handleKnowledgeNotFound(c *gin.Context, err error, message string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}
