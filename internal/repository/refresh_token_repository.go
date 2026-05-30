package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *domain.RefreshToken) error
	FindByTokenHash(hash string) (*domain.RefreshToken, error)
	RevokeByTokenHash(hash string) error
	DeleteExpired() error
	RevokeAllForUser(userID uuid.UUID) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(hash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	if err := r.db.Where("token_hash = ? AND revoked = false AND expires_at > now()", hash).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) RevokeByTokenHash(hash string) error {
	return r.db.Model(&domain.RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Model(&domain.RefreshToken{}).Where("expires_at < now()").Delete(&domain.RefreshToken{}).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	return r.db.Model(&domain.RefreshToken{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
