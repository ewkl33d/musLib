package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"musLib/db"
	_ "musLib/docs"
	"musLib/env"
	"musLib/handlers"
	"musLib/logger"
)

// @title Music Library API
// @version 0.0.1
// @description This is a simple API for managing music library.
// @host localhost:8080
// @BasePath /
func main() {
	logger := logger.GetLogger()

	r := gin.Default()

	env.EnvInit()

	database := initDB()

	setupHandlers(r, database)

	logger.Info("application starts")
	r.Run()
}

func initDB() *gorm.DB {
	logger := logger.GetLogger()

	dsn := connPgStr()
	logger.Debug("dsn = " + dsn)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Panic("failed to connect database")
	}
	logger.Info("connected to db")

	err = database.AutoMigrate(&db.Song{})
	if err != nil {
		logger.Panic("failed to migrate database: " + err.Error())
	}
	logger.Info("automigrate to db set succesifully")

	return database
}

func setupHandlers(r *gin.Engine, database *gorm.DB) {
	songHandler := &handlers.SongHandler{DB: database}
	logger := logger.GetLogger()
	r.GET("/songs", songHandler.GetSongs)
	logger.Debug("Registered GET /songs handler")

	r.GET("/songs/:id/text", songHandler.GetSongText)
	logger.Debug("Registered GET /songs/:id/text handler")

	r.DELETE("/songs/:id", songHandler.DeleteSong)
	logger.Debug("Registered DELETE /songs/:id handler")

	r.PUT("/songs/:id", songHandler.UpdateSong)
	logger.Debug("Registered PUT /songs/:id handler")

	r.POST("/songs/add", songHandler.AddSong)
	logger.Debug("Registered POST /songs/add handler")

	logger.Info("All handlers are registered")
}

func connPgStr() string {
	return "host=" + env.GetEnv("DB_PG_HOST", "") + " user=" + env.GetEnv("DB_PG_USER", "") + " password=" + env.GetEnv("DB_PG_PASSWORD", "") + " dbname=" + env.GetEnv("DB_PG_NAME", "") + " port=" + env.GetEnv("DB_PG_PORT", "")
}
