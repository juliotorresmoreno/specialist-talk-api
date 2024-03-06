package chats

import "github.com/gin-gonic/gin"

type ChatsRouter struct{}

func SetupAPIRoutes(g *gin.RouterGroup) {
	h := &ChatsRouter{}

	g.GET("", h.find)
}

func (h *ChatsRouter) find(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "chats find",
	})
}
