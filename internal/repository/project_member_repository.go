package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type ProjectMemberRepository interface {
	Add(m *domain.ProjectMember) error
	FindByProjectAndUser(projectID, userID uuid.UUID) (*domain.ProjectMember, error)
	ListByProject(projectID uuid.UUID, offset, limit int) ([]domain.ProjectMember, int64, error)
	UpdateRole(projectID, userID uuid.UUID, role domain.ProjectRole) error
	Remove(projectID, userID uuid.UUID) error
}

type projectMemberRepository struct {
	db *gorm.DB
}

func NewProjectMemberRepository(db *gorm.DB) ProjectMemberRepository {
	return &projectMemberRepository{db: db}
}

func (r *projectMemberRepository) Add(m *domain.ProjectMember) error {
	return r.db.Create(m).Error
}

func (r *projectMemberRepository) FindByProjectAndUser(projectID, userID uuid.UUID) (*domain.ProjectMember, error) {
	var member domain.ProjectMember
	if err := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *projectMemberRepository) ListByProject(projectID uuid.UUID, offset, limit int) ([]domain.ProjectMember, int64, error) {
	var members []domain.ProjectMember
	var total int64

	query := r.db.Where("project_id = ?", projectID).Preload("User")

	if err := query.Model(&domain.ProjectMember{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&members).Error; err != nil {
		return nil, 0, err
	}

	return members, total, nil
}

func (r *projectMemberRepository) UpdateRole(projectID, userID uuid.UUID, role domain.ProjectRole) error {
	return r.db.Model(&domain.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("role", role).Error
}

func (r *projectMemberRepository) Remove(projectID, userID uuid.UUID) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}
