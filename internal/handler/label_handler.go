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

type LabelHandler struct {
	labelSvc service.LabelService
	validate *validator.Validate
}

func NewLabelHandler(labelSvc service.LabelService) *LabelHandler {
	return &LabelHandler{
		labelSvc: labelSvc,
		validate: validator.New(),
	}
}

// GET /api/v1/projects/:id/labels
func (h *LabelHandler) ListLabels(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	labels, err := h.labelSvc.ListLabels(userID, projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(labels, ""))
}

// POST /api/v1/projects/:id/labels
func (h *LabelHandler) CreateLabel(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	var req domain.CreateLabelRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	label, err := h.labelSvc.CreateLabel(userID, role, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(label, "Label created successfully"))
}

// PATCH /api/v1/labels/:labelId
func (h *LabelHandler) UpdateLabel(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	labelID, err := uuid.Parse(c.Param("labelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid label ID"))
		return
	}

	var req domain.UpdateLabelRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	label, err := h.labelSvc.UpdateLabel(userID, role, labelID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(label, "Label updated successfully"))
}

// DELETE /api/v1/labels/:labelId
func (h *LabelHandler) DeleteLabel(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	labelID, err := uuid.Parse(c.Param("labelId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid label ID"))
		return
	}

	err = h.labelSvc.DeleteLabel(userID, role, labelID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
