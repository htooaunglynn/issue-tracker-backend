package service

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"gorm.io/gorm"
)

type AttachmentService interface {
	ListAttachments(callerID uuid.UUID, issueID uuid.UUID, page, size int) ([]domain.AttachmentResponse, int64, error)
	UploadAttachment(callerID uuid.UUID, issueID uuid.UUID, fileHeader *multipart.FileHeader, uploadDir string, maxSize int64) (*domain.AttachmentResponse, error)
	DeleteAttachment(callerID uuid.UUID, attID uuid.UUID) error
	GetAttachmentForDownload(callerID uuid.UUID, attID uuid.UUID) (*domain.Attachment, error)
}

type attachmentService struct {
	attachmentRepo    repository.AttachmentRepository
	issueRepo         repository.IssueRepository
	projectMemberRepo repository.ProjectMemberRepository
}

func NewAttachmentService(
	attachmentRepo repository.AttachmentRepository,
	issueRepo repository.IssueRepository,
	projectMemberRepo repository.ProjectMemberRepository,
) AttachmentService {
	return &attachmentService{
		attachmentRepo:    attachmentRepo,
		issueRepo:         issueRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

func (s *attachmentService) ListAttachments(callerID uuid.UUID, issueID uuid.UUID, page, size int) ([]domain.AttachmentResponse, int64, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrIssueNotFound
		}
		return nil, 0, err
	}

	if !s.isProjectMember(callerID, issue.ProjectID) {
		return nil, 0, ErrNotProjectMember
	}

	offset := (page - 1) * size
	attachments, total, err := s.attachmentRepo.ListByIssue(issueID, offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.AttachmentResponse, len(attachments))
	for i, a := range attachments {
		responses[i] = domain.AttachmentToResponse(&a)
	}
	return responses, total, nil
}

func (s *attachmentService) UploadAttachment(callerID uuid.UUID, issueID uuid.UUID, fileHeader *multipart.FileHeader, uploadDir string, maxSize int64) (*domain.AttachmentResponse, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrIssueNotFound
		}
		return nil, err
	}

	if !s.isProjectMember(callerID, issue.ProjectID) {
		return nil, ErrNotProjectMember
	}

	if fileHeader.Size > maxSize {
		return nil, ErrFileTooLarge
	}

	dir := filepath.Join(uploadDir, issueID.String())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	safeFilename := sanitizeFilename(fileHeader.Filename)
	storedName := uuid.New().String() + "_" + safeFilename
	filePath := filepath.Join(dir, storedName)

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	defer dst.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
				return nil, fmt.Errorf("failed to write file: %w", writeErr)
			}
		}
		if readErr != nil {
			break
		}
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	att := &domain.Attachment{
		IssueID:    issueID,
		UploaderID: callerID,
		FileName:   fileHeader.Filename,
		FilePath:   filePath,
		MimeType:   mimeType,
		SizeBytes:  fileHeader.Size,
	}

	if err := s.attachmentRepo.Create(att); err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	// Reload with Uploader preloaded
	full, err := s.attachmentRepo.FindByID(att.ID)
	if err != nil {
		return nil, err
	}
	resp := domain.AttachmentToResponse(full)
	return &resp, nil
}

func (s *attachmentService) DeleteAttachment(callerID uuid.UUID, attID uuid.UUID) error {
	att, err := s.attachmentRepo.FindByID(attID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrAttachmentNotFound
		}
		return err
	}

	if !s.canDeleteAttachment(callerID, att) {
		return ErrForbidden
	}

	// Remove from disk first; proceed with DB delete regardless
	_ = os.Remove(att.FilePath)

	return s.attachmentRepo.Delete(attID)
}

func (s *attachmentService) GetAttachmentForDownload(callerID uuid.UUID, attID uuid.UUID) (*domain.Attachment, error) {
	att, err := s.attachmentRepo.FindByID(attID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAttachmentNotFound
		}
		return nil, err
	}

	issue, err := s.issueRepo.FindByID(att.IssueID)
	if err != nil {
		return nil, err
	}

	if !s.isProjectMember(callerID, issue.ProjectID) {
		return nil, ErrNotProjectMember
	}

	return att, nil
}

func (s *attachmentService) isProjectMember(userID, projectID uuid.UUID) bool {
	_, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	return err == nil
}

func (s *attachmentService) canDeleteAttachment(callerID uuid.UUID, att *domain.Attachment) bool {
	if att.UploaderID == callerID {
		return true
	}
	issue, err := s.issueRepo.FindByID(att.IssueID)
	if err != nil {
		return false
	}
	member, err := s.projectMemberRepo.FindByProjectAndUser(issue.ProjectID, callerID)
	if err != nil {
		return false
	}
	return member.Role == domain.ProjectRoleAdmin || member.Role == domain.ProjectRoleManager
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	// Replace any non-alphanumeric chars (except dot, dash, underscore) with underscore
	var sb strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	result := sb.String()
	if result == "" || result == "." {
		return "file"
	}
	return result
}
