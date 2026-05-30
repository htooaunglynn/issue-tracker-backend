package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/middleware"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
)

type AuthHandler struct {
	authSvc  service.AuthService
	validate *validator.Validate
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{
		authSvc:  authSvc,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	authResp, err := h.authSvc.Register(req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(authResp, "user registered successfully"))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	authResp, err := h.authSvc.Login(req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(authResp, "logged in successfully"))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	authResp, err := h.authSvc.RefreshTokens(req.RefreshToken)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(authResp, "tokens refreshed successfully"))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	if err := h.authSvc.Logout(req.RefreshToken); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	userResp, err := h.authSvc.GetMe(userID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(userResp, ""))
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req domain.ForgotPasswordRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	if err := h.authSvc.ForgotPassword(req.Email); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil, "password reset instructions sent"))
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	if err := h.authSvc.ResetPassword(req); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil, "password reset successfully"))
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	var req domain.UpdateProfileRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	userResp, err := h.authSvc.UpdateProfile(userID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(userResp, "profile updated successfully"))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	var req domain.ChangePasswordRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	if err := h.authSvc.ChangePassword(userID, req); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil, "password changed successfully"))
}
