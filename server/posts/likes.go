package posts

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
)

func (h *PostsRouter) LikePost(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	like := models.Like{
		PostId:   uint(id),
		AuthorId: uint(session.ID),
	}

	count := int64(0)
	db.DefaultClient.
		Model(&models.Like{}).
		Where(&models.Like{PostId: uint(id), AuthorId: session.ID}).
		Limit(1).
		Count(&count)
	if count > 0 {
		c.JSON(400, gin.H{
			"message": "Post already liked",
		})
		return
	}

	err = db.DefaultClient.
		Create(&like).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(201, gin.H{"message": "Post liked"})
}

func (h *PostsRouter) UnlikePost(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	err = db.DefaultClient.
		Where(models.Like{PostId: uint(id), AuthorId: session.ID}).
		Delete(&models.Like{}).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(200, gin.H{"message": "Post unliked"})
}
