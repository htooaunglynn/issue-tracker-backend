package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type ActivityRepository interface {
	Create(tx *gorm.DB, activity *domain.IssueActivity) error
	ListByIssue(issueID uuid.UUID, offset, limit int) ([]domain.IssueActivity, int64, error)
}

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(tx *gorm.DB, activity *domain.IssueActivity) error {
	return tx.Create(activity).Error
}

func (r *activityRepository) ListByIssue(issueID uuid.UUID, offset, limit int) ([]domain.IssueActivity, int64, error) {
	var activities []domain.IssueActivity
	var total int64

	q := r.db.Model(&domain.IssueActivity{}).
		Preload("Actor").
		Where("issue_id = ?", issueID).
		Order("created_at ASC")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Offset(offset).Limit(limit).Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}
