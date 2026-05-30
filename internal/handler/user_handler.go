package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
)

type UserHandler struct {
	userSvc  service.UserService
	validate *validator.Validate
}

func NewUserHandler(userSvc service.UserService) *UserHandler {
	return &UserHandler{
		userSvc:  userSvc,
		validate: validator.New(),
	}
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page := 1
	if p := c.DefaultQuery("page", "1"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	size := 20
	if s := c.DefaultQuery("page_size", "20"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 100 {
			size = parsed
		}
	}

	users, total, err := h.userSvc.ListUsers(page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{
		Page:     page,
		PageSize: size,
		Total:    total,
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(users, "", meta))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("BAD_REQUEST", "invalid user id"))
		return
	}

	userResp, err := h.userSvc.GetUser(id)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(userResp, ""))
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("BAD_REQUEST", "invalid user id"))
		return
	}

	var req domain.UpdateProfileRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	userResp, err := h.userSvc.UpdateUser(id, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(userResp, "user updated successfully"))
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("BAD_REQUEST", "invalid user id"))
		return
	}

	if err := h.userSvc.DeleteUser(id); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
