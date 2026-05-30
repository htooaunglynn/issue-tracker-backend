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
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
)

type ProjectHandler struct {
	projectSvc service.ProjectService
	validate   *validator.Validate
}

func NewProjectHandler(projectSvc service.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectSvc: projectSvc,
		validate:   validator.New(),
	}
}

// GET /api/v1/projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	page := 1
	size := 20
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

	projects, total, err := h.projectSvc.ListProjects(userID, page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{Page: page, PageSize: size, Total: total}
	c.JSON(http.StatusOK, response.SuccessWithMeta(projects, "", meta))
}

// POST /api/v1/projects
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	var req domain.CreateProjectRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	project, err := h.projectSvc.CreateProject(userID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(project, "Project created successfully"))
}

// GET /api/v1/projects/:id
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	project, err := h.projectSvc.GetProject(userID, projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(project, ""))
}

// PATCH /api/v1/projects/:id
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	var req domain.UpdateProjectRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	project, err := h.projectSvc.UpdateProject(userID, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(project, "Project updated successfully"))
}

// DELETE /api/v1/projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	err = h.projectSvc.DeleteProject(userID, role, projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// POST /api/v1/projects/:id/archive
func (h *ProjectHandler) ArchiveProject(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	err = h.projectSvc.ArchiveProject(userID, role, projectID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil, "Project archived successfully"))
}

// GET /api/v1/projects/:id/members
func (h *ProjectHandler) ListMembers(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	page := 1
	size := 20
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

	members, total, err := h.projectSvc.ListMembers(userID, projectID, page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{Page: page, PageSize: size, Total: total}
	c.JSON(http.StatusOK, response.SuccessWithMeta(members, "", meta))
}

// POST /api/v1/projects/:id/members
func (h *ProjectHandler) AddMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	var req domain.AddMemberRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	member, err := h.projectSvc.AddMember(userID, role, projectID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(member, "Member added successfully"))
}

// PATCH /api/v1/projects/:id/members/:userId
func (h *ProjectHandler) UpdateMemberRole(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid user ID"))
		return
	}

	var req domain.UpdateMemberRoleRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	err = h.projectSvc.UpdateMemberRole(userID, role, projectID, targetUserID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(nil, "Member role updated successfully"))
}

// DELETE /api/v1/projects/:id/members/:userId
func (h *ProjectHandler) RemoveMember(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(domain.GlobalRole)

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid project ID"))
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid user ID"))
		return
	}

	err = h.projectSvc.RemoveMember(userID, role, projectID, targetUserID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
