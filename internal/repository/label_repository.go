package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type LabelRepository interface {
	Create(l *domain.Label) error
	FindByID(id uuid.UUID) (*domain.Label, error)
	FindByNameAndProject(projectID uuid.UUID, name string) (*domain.Label, error)
	ListByProject(projectID uuid.UUID) ([]domain.Label, error)
	Update(l *domain.Label) error
	Delete(id uuid.UUID) error
}

type labelRepository struct {
	db *gorm.DB
}

func NewLabelRepository(db *gorm.DB) LabelRepository {
	return &labelRepository{db: db}
}

func (r *labelRepository) Create(l *domain.Label) error {
	return r.db.Create(l).Error
}

func (r *labelRepository) FindByID(id uuid.UUID) (*domain.Label, error) {
	var label domain.Label
	if err := r.db.Where("id = ?", id).First(&label).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &label, nil
}

func (r *labelRepository) FindByNameAndProject(projectID uuid.UUID, name string) (*domain.Label, error) {
	var label domain.Label
	if err := r.db.Where("project_id = ? AND name = ?", projectID, name).First(&label).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &label, nil
}

func (r *labelRepository) ListByProject(projectID uuid.UUID) ([]domain.Label, error) {
	var labels []domain.Label
	if err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&labels).Error; err != nil {
		return nil, err
	}
	return labels, nil
}

func (r *labelRepository) Update(l *domain.Label) error {
	return r.db.Save(l).Error
}

func (r *labelRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Label{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}
