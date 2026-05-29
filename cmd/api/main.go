package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htooaunglynn/issue-tracker-backend/internal/config"
	"github.com/htooaunglynn/issue-tracker-backend/internal/db"
	"github.com/htooaunglynn/issue-tracker-backend/internal/db/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	gormDB, err := db.New(cfg.Database)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	// Optional: seed only in development
	if cfg.Server.Env == "development" {
		if err := seed.Run(gormDB); err != nil {
			log.Printf("seed warning: %v", err)
			// non-fatal: seed failures should not kill the server
		}
	}

	engine := gin.Default()

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/healthz", func(c *gin.Context) {
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"db":     err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "db": "up"})
	})

	if err := engine.Run(cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
