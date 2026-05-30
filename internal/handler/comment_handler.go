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

type CommentHandler struct {
	commentSvc service.CommentService
	validate   *validator.Validate
}

func NewCommentHandler(commentSvc service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentSvc: commentSvc,
		validate:   validator.New(),
	}
}

// GET /api/v1/issues/:id/comments
func (h *CommentHandler) ListComments(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	page, size := parsePagination(c)

	comments, total, err := h.commentSvc.ListComments(userID, issueID, page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{Page: page, PageSize: size, Total: total}
	c.JSON(http.StatusOK, response.SuccessWithMeta(comments, "", meta))
}

// POST /api/v1/issues/:id/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	var req domain.CreateCommentRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	comment, err := h.commentSvc.CreateComment(userID, issueID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(comment, "Comment created successfully"))
}

// PATCH /api/v1/comments/:commentId
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid comment ID"))
		return
	}

	var req domain.UpdateCommentRequest
	if !bindAndValidate(c, h.validate, &req) {
		return
	}

	comment, err := h.commentSvc.UpdateComment(userID, commentID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.Success(comment, "Comment updated successfully"))
}

// DELETE /api/v1/comments/:commentId
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid comment ID"))
		return
	}

	if err := h.commentSvc.DeleteComment(userID, commentID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
