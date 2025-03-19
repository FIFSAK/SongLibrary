package main

import (
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"os"

	"SongLibrary/internal/handlers"
	"SongLibrary/internal/models"
	"SongLibrary/pkg/logger"

	_ "SongLibrary/docs"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title           Song Library API
// @version         1.0
// @description     API for managing songs library with external API integration
// @host            localhost:8080
// @BasePath        /
func main() {
	if err := godotenv.Load(); err != nil {
		logger.Log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		logger.Log.Fatal("DATABASE_DSN is not set in .env file")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.WithError(err).Fatal("Failed to connect to database")
	}
	logger.Log.Info("Database connected")

	if err = db.AutoMigrate(&models.Song{}); err != nil {
		logger.Log.WithError(err).Fatal("Failed to migrate database")
	}
	logger.Log.Info("Database migrated")

	router := gin.New()
	router.Use(gin.LoggerWithWriter(logger.Log.Writer()), gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/songs", handlers.GetSongsHandler(db))
	router.GET("/songs/:id/verses", handlers.GetSongVersesHandler(db))
	router.POST("/songs", handlers.CreateSongHandler(db))
	router.PUT("/songs/:id", handlers.UpdateSongHandler(db))
	router.DELETE("/songs/:id", handlers.DeleteSongHandler(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Log.Infof("Server running on port %s", port)
	router.Run(":" + port)
}
