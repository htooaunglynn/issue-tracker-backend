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
	projectRepo := repository.NewProjectRepository(gormDB)
	projectMemberRepo := repository.NewProjectMemberRepository(gormDB)
	labelRepo := repository.NewLabelRepository(gormDB)
	issueRepo := repository.NewIssueRepository(gormDB)
	activityRepo := repository.NewActivityRepository(gormDB)
	commentRepo := repository.NewCommentRepository(gormDB)
	attachmentRepo := repository.NewAttachmentRepository(gormDB)

	// Services
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, passwordResetRepo, cfg.JWT, gormDB)
	userSvc := service.NewUserService(userRepo)
	projectSvc := service.NewProjectService(projectRepo, projectMemberRepo, userRepo)
	labelSvc := service.NewLabelService(labelRepo, projectRepo, projectMemberRepo)
	issueSvc := service.NewIssueService(gormDB, issueRepo, activityRepo, projectRepo, projectMemberRepo, labelRepo)
	commentSvc := service.NewCommentService(gormDB, commentRepo, issueRepo, activityRepo, projectMemberRepo, userRepo)
	attachmentSvc := service.NewAttachmentService(attachmentRepo, issueRepo, projectMemberRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	labelH := handler.NewLabelHandler(labelSvc)
	issueH := handler.NewIssueHandler(issueSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	attachmentH := handler.NewAttachmentHandler(attachmentSvc, cfg.Upload.Dir, cfg.Upload.MaxSize)

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

		projects := v1.Group("/projects")
		{
			projects.GET("", middleware.AuthRequired(cfg.JWT), projectH.ListProjects)
			projects.POST("", middleware.AuthRequired(cfg.JWT), middleware.RequireRole(domain.RoleAdmin, domain.RoleManager), projectH.CreateProject)
			projects.GET("/:id", middleware.AuthRequired(cfg.JWT), projectH.GetProject)
			projects.PATCH("/:id", middleware.AuthRequired(cfg.JWT), projectH.UpdateProject)
			projects.DELETE("/:id", middleware.AuthRequired(cfg.JWT), projectH.DeleteProject)
			projects.POST("/:id/archive", middleware.AuthRequired(cfg.JWT), projectH.ArchiveProject)

			projects.GET("/:id/members", middleware.AuthRequired(cfg.JWT), projectH.ListMembers)
			projects.POST("/:id/members", middleware.AuthRequired(cfg.JWT), projectH.AddMember)
			projects.PATCH("/:id/members/:userId", middleware.AuthRequired(cfg.JWT), projectH.UpdateMemberRole)
			projects.DELETE("/:id/members/:userId", middleware.AuthRequired(cfg.JWT), projectH.RemoveMember)

			projects.GET("/:id/labels", middleware.AuthRequired(cfg.JWT), labelH.ListLabels)
			projects.POST("/:id/labels", middleware.AuthRequired(cfg.JWT), labelH.CreateLabel)
		}

		labels := v1.Group("/labels")
		{
			labels.PATCH("/:labelId", middleware.AuthRequired(cfg.JWT), labelH.UpdateLabel)
			labels.DELETE("/:labelId", middleware.AuthRequired(cfg.JWT), labelH.DeleteLabel)
		}

		projects.GET("/:id/issues", middleware.AuthRequired(cfg.JWT), issueH.ListIssues)
		projects.POST("/:id/issues", middleware.AuthRequired(cfg.JWT), issueH.CreateIssue)

		issues := v1.Group("/issues")
		{
			issues.GET("/:id", middleware.AuthRequired(cfg.JWT), issueH.GetIssue)
			issues.PATCH("/:id", middleware.AuthRequired(cfg.JWT), issueH.UpdateIssue)
			issues.DELETE("/:id", middleware.AuthRequired(cfg.JWT), issueH.DeleteIssue)
			issues.POST("/:id/assign", middleware.AuthRequired(cfg.JWT), issueH.AssignIssue)
			issues.POST("/:id/status", middleware.AuthRequired(cfg.JWT), issueH.ChangeStatus)
			issues.GET("/:id/activities", middleware.AuthRequired(cfg.JWT), issueH.ListActivities)
			issues.GET("/:id/comments", middleware.AuthRequired(cfg.JWT), commentH.ListComments)
			issues.POST("/:id/comments", middleware.AuthRequired(cfg.JWT), commentH.CreateComment)
			issues.GET("/:id/attachments", middleware.AuthRequired(cfg.JWT), attachmentH.ListAttachments)
			issues.POST("/:id/attachments", middleware.AuthRequired(cfg.JWT), attachmentH.UploadAttachment)
		}

		comments := v1.Group("/comments")
		{
			comments.PATCH("/:commentId", middleware.AuthRequired(cfg.JWT), commentH.UpdateComment)
			comments.DELETE("/:commentId", middleware.AuthRequired(cfg.JWT), commentH.DeleteComment)
		}

		attachments := v1.Group("/attachments")
		{
			attachments.DELETE("/:attId", middleware.AuthRequired(cfg.JWT), attachmentH.DeleteAttachment)
			attachments.GET("/:attId/download", middleware.AuthRequired(cfg.JWT), attachmentH.DownloadAttachment)
		}
	}

	if err := engine.Run(cfg.Server.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
