package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/middleware"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
)

type IssueHandler struct {
	issueSvc service.IssueService
	validate *validator.Validate
}

func NewIssueHandler(issueSvc service.IssueService) *IssueHandler {
	return &IssueHandler{
		issueSvc: issueSvc,
		validate: validator.New(),
	}
}

// GET /api/v1/projects/:id/issues
func (h *IssueHandler) ListIssues(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	page, size := parsePagination(c)
	filter := repository.IssueFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Type:     c.Query("type"),
		Sort:     c.Query("sort"),
	}
	if assigneeStr := c.Query("assignee"); assigneeStr != "" {
		if id, err := uuid.Parse(assigneeStr); err == nil {
			filter.Assignee = &id
		}
	}

	issues, total, err := h.issueSvc.ListIssues(userID, projectID, filter, page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{Page: page, PageSize: size, Total: total}
	c.JSON(http.StatusOK, response.SuccessWithMeta(issues, "", meta))
}

// POST /api/v1/projects/:id/issues
func (h *IssueHandler) CreateIssue(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	var req domain.CreateIssueRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	issue, err := h.issueSvc.CreateIssue(userID, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(issue, "Issue created successfully"))
}

// GET /api/v1/issues/:id
func (h *IssueHandler) GetIssue(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	issue, err := h.issueSvc.GetIssue(userID, issueID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(issue, ""))
}

// PATCH /api/v1/issues/:id
func (h *IssueHandler) UpdateIssue(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	var req domain.UpdateIssueRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	issue, err := h.issueSvc.UpdateIssue(userID, issueID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(issue, "Issue updated successfully"))
}

// DELETE /api/v1/issues/:id
func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	if err := h.issueSvc.DeleteIssue(userID, role, issueID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// POST /api/v1/issues/:id/assign
func (h *IssueHandler) AssignIssue(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	var req domain.AssignIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("BAD_REQUEST", err.Error()))
		return
	}

	issue, err := h.issueSvc.AssignIssue(userID, issueID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(issue, "Issue assigned successfully"))
}

// POST /api/v1/issues/:id/status
func (h *IssueHandler) ChangeStatus(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	var req domain.ChangeStatusRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	issue, err := h.issueSvc.ChangeStatus(userID, issueID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(issue, "Issue status updated"))
}

// GET /api/v1/issues/:id/activities
func (h *IssueHandler) ListActivities(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	page, size := parsePagination(c)

	activities, total, err := h.issueSvc.ListActivities(userID, issueID, page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{Page: page, PageSize: size, Total: total}
	c.JSON(http.StatusOK, response.SuccessWithMeta(activities, "", meta))
}

func parsePagination(c *gin.Context) (page, size int) {
	page = 1
	size = 20
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if s := c.Query("page_size"); s != "" {
		if parsed, err := strconv.Atoi(s); err == nil && parsed > 0 && parsed <= 100 {
			size = parsed
		}
	}
	return
}
