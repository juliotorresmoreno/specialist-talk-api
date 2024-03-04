package posts

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/models"
	"github.com/juliotorresmoreno/specialist-talk-api/utils"
)

type PostsRouter struct{}

func SetupApiRoutes(g *gin.RouterGroup) {
	h := &PostsRouter{}

	g.GET("", h.GetPosts)
	g.GET("/:id", h.GetPost)
	g.POST("", h.CreatePost)
	g.PATCH("/:id", h.UpdatePost)
	g.DELETE("/:id", h.DeletePost)

	g.POST("/:id/like", h.LikePost)
	g.DELETE("/:id/like", h.UnlikePost)

	g.POST("/:id/comment", h.CreateComment)
	g.GET("/:id/comments", h.GetComments)
	g.DELETE("/:id/comment/:commentId", h.DeleteComment)
}

type User struct {
	ID           uint       `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Username     string     `json:"username"`
	PhotoURL     string     `json:"photo_url"`
	Business     string     `json:"business"`
	PositionName string     `json:"position_name"`
	CreationAt   time.Time  `json:"creation_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type Post struct {
	ID         int        `json:"id"`
	Content    string     `json:"content"`
	AuthorId   int        `json:"author_id"`
	Author     User       `json:"author"`
	Likes      int64      `json:"likes" gorm:"-"`
	Liked      bool       `json:"liked" gorm:"-"`
	Comments   int64      `json:"comments" gorm:"-"`
	CreationAt time.Time  `json:"creation_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (h *PostsRouter) GetPost(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))

	post := &Post{}
	err = db.DefaultClient.
		Model(&models.Post{}).
		Preload("Author").
		Where(models.Post{ID: uint(id)}).
		First(post).Error
	if err != nil {
		c.JSON(404, gin.H{
			"message": "Post not found",
		})
		return
	}

	count := int64(0)
	err = db.DefaultClient.
		Model(&models.Like{}).
		Where(models.Like{PostId: uint(post.ID)}).
		Count(&count).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	post.Likes = count

	db.DefaultClient.
		Model(&models.Like{}).
		Where(&models.Like{PostId: uint(post.ID), AuthorId: session.ID}).
		Limit(1).
		Count(&count)
	post.Liked = count == 1

	err = db.DefaultClient.
		Model(&models.Comment{}).
		Where(models.Comment{PostId: uint(post.ID)}).
		Count(&count).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	post.Comments = count

	c.JSON(200, post)
}

func (h *PostsRouter) GetPosts(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	posts := &[]*Post{}
	err = db.DefaultClient.
		Model(&models.Post{}).
		Preload("Author").
		Order("creation_at DESC").
		Find(posts).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}
	for _, post := range *posts {
		count := int64(0)
		err = db.DefaultClient.
			Model(&models.Like{}).
			Where(models.Like{PostId: uint(post.ID)}).
			Count(&count).Error
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error",
			})
			return
		}
		post.Likes = count

		db.DefaultClient.
			Model(&models.Like{}).
			Where(&models.Like{PostId: uint(post.ID), AuthorId: session.ID}).
			Limit(1).
			Count(&count)
		post.Liked = count == 1

		err = db.DefaultClient.
			Model(&models.Comment{}).
			Where(models.Comment{PostId: uint(post.ID)}).
			Count(&count).Error
		if err != nil {
			c.JSON(500, gin.H{
				"message": "Internal server error",
			})
			return
		}
		post.Comments = count
	}

	c.JSON(200, posts)
}

func (h *PostsRouter) CreatePost(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	payload := &Post{}
	err = c.ShouldBind(payload)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid payload",
		})
		return
	}

	post := models.Post{
		Content:  payload.Content,
		AuthorId: uint(session.ID),
	}

	err = db.DefaultClient.
		Create(&post).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(201, gin.H{"message": "Post created"})
}

func (h *PostsRouter) UpdatePost(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	payload := &Post{}
	err = c.ShouldBind(payload)
	if err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid payload",
		})
		return
	}

	post := models.Post{Content: payload.Content}

	err = db.DefaultClient.
		Model(&models.Post{}).
		Where(models.Post{ID: uint(payload.ID), AuthorId: session.ID}).
		Updates(&post).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(200, gin.H{"message": "Post updated"})
}

func (h *PostsRouter) DeletePost(c *gin.Context) {
	session, err := utils.ValidateSession(c)
	if err != nil {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	id, _ := strconv.Atoi(c.Param("id"))
	err = db.DefaultClient.
		Where(models.Post{ID: uint(id), AuthorId: session.ID}).
		Delete(&models.Post{}).Error
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(200, gin.H{"message": "Post deleted"})
}
