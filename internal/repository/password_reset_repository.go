package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(reset *domain.PasswordReset) error
	FindByTokenHash(hash string) (*domain.PasswordReset, error)
	MarkUsed(id uuid.UUID) error
	DeleteExpired() error
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(reset *domain.PasswordReset) error {
	return r.db.Create(reset).Error
}

func (r *passwordResetRepository) FindByTokenHash(hash string) (*domain.PasswordReset, error) {
	var reset domain.PasswordReset
	if err := r.db.Where("token_hash = ? AND used = false AND expires_at > now()", hash).First(&reset).Error; err != nil {
		return nil, err
	}
	return &reset, nil
}

func (r *passwordResetRepository) MarkUsed(id uuid.UUID) error {
	return r.db.Model(&domain.PasswordReset{}).Where("id = ?", id).Update("used", true).Error
}

func (r *passwordResetRepository) DeleteExpired() error {
	return r.db.Model(&domain.PasswordReset{}).Where("expires_at < now()").Delete(&domain.PasswordReset{}).Error
}
