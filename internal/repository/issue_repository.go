package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IssueFilter struct {
	Status   string
	Priority string
	Type     string
	Assignee *uuid.UUID
	Sort     string
}

type IssueRepository interface {
	Create(tx *gorm.DB, issue *domain.Issue) error
	FindByID(id uuid.UUID) (*domain.Issue, error)
	List(projectID uuid.UUID, filter IssueFilter, offset, limit int) ([]domain.Issue, int64, error)
	Update(issue *domain.Issue) error
	Delete(id uuid.UUID) error
	NextNumber(tx *gorm.DB, projectID uuid.UUID) (int, error)
	GetLabels(issueID uuid.UUID) ([]domain.Label, error)
	SetLabels(tx *gorm.DB, issueID uuid.UUID, labelIDs []uuid.UUID) error
}

type issueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) IssueRepository {
	return &issueRepository{db: db}
}

func (r *issueRepository) Create(tx *gorm.DB, issue *domain.Issue) error {
	return tx.Create(issue).Error
}

func (r *issueRepository) FindByID(id uuid.UUID) (*domain.Issue, error) {
	var issue domain.Issue
	err := r.db.
		Preload("Project").
		Preload("Reporter").
		Preload("Assignee").
		Where("issues.id = ? AND issues.deleted_at IS NULL", id).
		First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

func (r *issueRepository) List(projectID uuid.UUID, filter IssueFilter, offset, limit int) ([]domain.Issue, int64, error) {
	q := r.db.Model(&domain.Issue{}).
		Preload("Reporter").
		Preload("Assignee").
		Where("project_id = ? AND deleted_at IS NULL", projectID)

	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		q = q.Where("priority = ?", filter.Priority)
	}
	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.Assignee != nil {
		q = q.Where("assignee_id = ?", *filter.Assignee)
	}

	sort := "created_at DESC"
	if filter.Sort != "" {
		if filter.Sort[0] == '-' {
			sort = filter.Sort[1:] + " DESC"
		} else {
			sort = filter.Sort + " ASC"
		}
	}
	q = q.Order(sort)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var issues []domain.Issue
	if err := q.Offset(offset).Limit(limit).Find(&issues).Error; err != nil {
		return nil, 0, err
	}
	return issues, total, nil
}

func (r *issueRepository) Update(issue *domain.Issue) error {
	return r.db.Save(issue).Error
}

func (r *issueRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Issue{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}

// NextNumber acquires a lock on the project row and returns the next issue number.
// Must be called inside a transaction.
func (r *issueRepository) NextNumber(tx *gorm.DB, projectID uuid.UUID) (int, error) {
	// Lock the project row to serialize concurrent issue creation
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&domain.Project{}).
		Where("id = ?", projectID).
		First(&domain.Project{}).Error; err != nil {
		return 0, err
	}

	var maxNum int
	if err := tx.Model(&domain.Issue{}).
		Where("project_id = ? AND deleted_at IS NULL", projectID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&maxNum).Error; err != nil {
		return 0, err
	}
	return maxNum + 1, nil
}

func (r *issueRepository) GetLabels(issueID uuid.UUID) ([]domain.Label, error) {
	var labels []domain.Label
	err := r.db.
		Joins("JOIN issue_labels ON issue_labels.label_id = labels.id").
		Where("issue_labels.issue_id = ? AND labels.deleted_at IS NULL", issueID).
		Find(&labels).Error
	return labels, err
}

func (r *issueRepository) SetLabels(tx *gorm.DB, issueID uuid.UUID, labelIDs []uuid.UUID) error {
	if err := tx.Where("issue_id = ?", issueID).Delete(&domain.IssueLabel{}).Error; err != nil {
		return err
	}
	for _, labelID := range labelIDs {
		il := domain.IssueLabel{IssueID: issueID, LabelID: labelID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&il).Error; err != nil {
			return err
		}
	}
	return nil
}
