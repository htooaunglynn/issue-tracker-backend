package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(p *domain.Project) error
	FindByID(id uuid.UUID) (*domain.Project, error)
	FindByKey(key string) (*domain.Project, error)
	List(userID uuid.UUID, offset, limit int) ([]domain.Project, int64, error)
	Update(p *domain.Project) error
	Delete(id uuid.UUID) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(p *domain.Project) error {
	return r.db.Create(p).Error
}

func (r *projectRepository) FindByID(id uuid.UUID) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.Where("id = ?", id).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindByKey(key string) (*domain.Project, error) {
	var project domain.Project
	if err := r.db.Where("key = ?", key).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) List(userID uuid.UUID, offset, limit int) ([]domain.Project, int64, error) {
	var projects []domain.Project
	var total int64

	query := r.db.
		Joins("LEFT JOIN project_members pm ON projects.id = pm.project_id").
		Where("projects.owner_id = ? OR pm.user_id = ?", userID, userID).
		Distinct("projects.id").
		Order("projects.created_at DESC")

	if err := query.Model(&domain.Project{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&projects).Error; err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *projectRepository) Update(p *domain.Project) error {
	return r.db.Save(p).Error
}

func (r *projectRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Project{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}
