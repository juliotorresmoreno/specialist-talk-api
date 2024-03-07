package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/juliotorresmoreno/specialist-talk-api/db"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
	"github.com/juliotorresmoreno/specialist-talk-api/server"
	"github.com/juliotorresmoreno/specialist-talk-api/server/events"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	time.Local = time.UTC // default to UTC for all time values
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	logger.SetupLogrus()
	db.Setup()
	events.Setup()

	r := gin.Default()
	server.SetupAPIRoutes(r.Group("/api"))
	events.SetupAPIRoutes(r.Group("/events"))

	r.Run(os.Getenv("ADDR"))
}
