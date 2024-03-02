package server

import (
	"github.com/gin-gonic/gin"
	"github.com/juliotorresmoreno/specialist-talk-api/server/auth"
	"github.com/juliotorresmoreno/specialist-talk-api/server/events"
	"github.com/juliotorresmoreno/specialist-talk-api/server/users"
)

func SetupAPIRoutes(r *gin.RouterGroup) {
	auth.SetupAPIRoutes(r)
	users.SetupAPIRoutes(r.Group("/users"))
	events.SetupAPIRoutes(r.Group("/events"))
}
