package posts

import (
	"time"

	"github.com/gin-gonic/gin"
)

type PostsRouter struct{}

func SetupApiRoutes(g *gin.RouterGroup) {
	h := &PostsRouter{}

	g.GET("", h.GetPosts)
	g.GET("/:id", h.GetPost)
	g.POST("", h.CreatePost)
	g.PUT("/:id", h.UpdatePost)
}

type User struct {
	ID           uint      `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Username     string    `json:"username"`
	PhotoURL     string    `json:"photo_url"`
	Business     string    `json:"business"`
	PositionName string    `json:"position_name"`
	CreationAt   time.Time `json:"creation_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `json:"deleted_at"`
}

type Post struct {
	ID         int       `json:"id"`
	Content    string    `json:"content"`
	AuthorId   int       `json:"author_id"`
	Author     User      `json:"author"`
	CreationAt time.Time `json:"creation_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  time.Time `json:"deleted_at"`
}

func (h *PostsRouter) GetPost(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Get post",
	})
}

func (h *PostsRouter) GetPosts(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Get posts",
	})
}

func (h *PostsRouter) CreatePost(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Create post",
	})
}

func (h *PostsRouter) UpdatePost(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Update post",
	})
}
