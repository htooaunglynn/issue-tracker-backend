package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type AttachmentRepository interface {
	Create(att *domain.Attachment) error
	FindByID(id uuid.UUID) (*domain.Attachment, error)
	ListByIssue(issueID uuid.UUID, offset, limit int) ([]domain.Attachment, int64, error)
	Delete(id uuid.UUID) error
}

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(att *domain.Attachment) error {
	return r.db.Create(att).Error
}

func (r *attachmentRepository) FindByID(id uuid.UUID) (*domain.Attachment, error) {
	var att domain.Attachment
	err := r.db.
		Preload("Uploader").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&att).Error
	if err != nil {
		return nil, err
	}
	return &att, nil
}

func (r *attachmentRepository) ListByIssue(issueID uuid.UUID, offset, limit int) ([]domain.Attachment, int64, error) {
	var attachments []domain.Attachment
	var total int64

	q := r.db.Model(&domain.Attachment{}).
		Preload("Uploader").
		Where("issue_id = ? AND deleted_at IS NULL", issueID).
		Order("created_at DESC")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Offset(offset).Limit(limit).Find(&attachments).Error; err != nil {
		return nil, 0, err
	}

	return attachments, total, nil
}

func (r *attachmentRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Attachment{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}
