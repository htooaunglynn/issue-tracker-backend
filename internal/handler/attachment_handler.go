package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/middleware"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
)

type AttachmentHandler struct {
	attachmentSvc service.AttachmentService
	uploadDir     string
	maxSize       int64
}

func NewAttachmentHandler(attachmentSvc service.AttachmentService, uploadDir string, maxSize int64) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentSvc: attachmentSvc,
		uploadDir:     uploadDir,
		maxSize:       maxSize,
	}
}

// GET /api/v1/issues/:id/attachments
func (h *AttachmentHandler) ListAttachments(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	page, size := parsePagination(c)

	attachments, total, err := h.attachmentSvc.ListAttachments(userID, issueID, page, size)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	meta := &response.Meta{Page: page, PageSize: size, Total: total}
	c.JSON(http.StatusOK, response.SuccessWithMeta(attachments, "", meta))
}

// POST /api/v1/issues/:id/attachments  (multipart/form-data, field: "file")
func (h *AttachmentHandler) UploadAttachment(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	issueID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid issue ID"))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("BAD_REQUEST", "file field is required"))
		return
	}

	att, err := h.attachmentSvc.UploadAttachment(userID, issueID, fileHeader, h.uploadDir, h.maxSize)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(att, "Attachment uploaded successfully"))
}

// DELETE /api/v1/attachments/:attId
func (h *AttachmentHandler) DeleteAttachment(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	attID, err := uuid.Parse(c.Param("attId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid attachment ID"))
		return
	}

	if err := h.attachmentSvc.DeleteAttachment(userID, attID); err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GET /api/v1/attachments/:attId/download
func (h *AttachmentHandler) DownloadAttachment(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	attID, err := uuid.Parse(c.Param("attId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_ID", "Invalid attachment ID"))
		return
	}

	att, err := h.attachmentSvc.GetAttachmentForDownload(userID, attID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.FileAttachment(att.FilePath, att.FileName)
}
