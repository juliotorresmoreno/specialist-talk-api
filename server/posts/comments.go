package posts

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
)

type Comment struct {
	ID         int        `json:"id"`
	Content    string     `json:"content"`
	AuthorID   uint       `json:"author_id"`
	Author     User       `json:"author"`
	CreationAt time.Time  `json:"creation_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (h *PostsRouter) createComment(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	payload := &Comment{}
	err = c.ShouldBind(payload)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid payload",
		})
		return
	}

	comment := &models.Comment{
		Content:  payload.Content,
		AuthorId: session.ID,
		PostId:   uint(id),
	}
	err = db.DefaultClient.Create(comment).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(201, gin.H{"message": "Comment created"})
}

func (h *PostsRouter) getComments(c *gin.Context) {
	_, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))

	comments := &[]Comment{}
	err = db.DefaultClient.
		Model(&models.Comment{}).
		Preload("Author").
		Where(&models.Comment{PostId: uint(id)}).
		Find(comments).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(200, comments)
}

func (h *PostsRouter) deleteComment(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	postID, _ := strconv.Atoi(c.Param("id"))
	commentID, _ := strconv.Atoi(c.Param("commentId"))

	err = db.DefaultClient.
		Where(&models.Comment{ID: uint(commentID), AuthorId: session.ID, PostId: uint(postID)}).
		Delete(&models.Comment{}).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(200, gin.H{"message": "Comment deleted"})
}
