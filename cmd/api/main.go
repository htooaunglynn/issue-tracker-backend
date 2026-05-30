package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htooaunglynn/issue-tracker-backend/internal/config"
	"github.com/htooaunglynn/issue-tracker-backend/internal/db"
	"github.com/htooaunglynn/issue-tracker-backend/internal/db/seed"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/handler"
	"github.com/htooaunglynn/issue-tracker-backend/internal/middleware"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
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
		}
	}

	// --- Dependency wiring ---

	// Repositories
	userRepo := repository.NewUserRepository(gormDB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(gormDB)
	passwordResetRepo := repository.NewPasswordResetRepository(gormDB)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, cfg.JWT, gormDB)
	userSvc := service.NewUserService(userRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)

	// --- Routes ---

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

	v1 := engine.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login", authH.Login)
			auth.POST("/refresh", authH.Refresh)
			auth.POST("/logout", authH.Logout)
			auth.GET("/me", middleware.AuthRequired(cfg.JWT), authH.Me)
			auth.POST("/password/forgot", authH.ForgotPassword)
			auth.POST("/password/reset", authH.ResetPassword)
		}

		users := v1.Group("/users")
		{
			users.PATCH("/me", middleware.AuthRequired(cfg.JWT), authH.UpdateProfile)
			users.PATCH("/me/password", middleware.AuthRequired(cfg.JWT), authH.ChangePassword)
			users.GET("", middleware.AuthRequired(cfg.JWT), middleware.RequireRole(domain.RoleAdmin), userH.ListUsers)
			users.GET("/:id", middleware.AuthRequired(cfg.JWT), userH.GetUser)
			users.PATCH("/:id", middleware.AuthRequired(cfg.JWT), middleware.RequireRole(domain.RoleAdmin), userH.UpdateUser)
			users.DELETE("/:id", middleware.AuthRequired(cfg.JWT), middleware.RequireRole(domain.RoleAdmin), userH.DeleteUser)
		}
	}

	if err := engine.Run(cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
