package auth

import "github.com/gin-gonic/gin"

func SetupAPIRoutes(r *gin.RouterGroup) {
	SetupAUTHRoutes(r.Group("auth"))
	SetupOAUTHRoutes(r.Group("oauth"))
}
