package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id uuid.UUID) (*domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	Update(user *domain.User) error
	List(offset, limit int) ([]domain.User, int64, error)
	Delete(id uuid.UUID) error
	FindMentionedUsers(projectID uuid.UUID, names []string) ([]domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) List(offset, limit int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	if err := r.db.Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}

// FindMentionedUsers finds project members whose name (lowercased, spaces stripped) matches any entry in names.
func (r *userRepository) FindMentionedUsers(projectID uuid.UUID, names []string) ([]domain.User, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var users []domain.User
	err := r.db.
		Joins("JOIN project_members pm ON pm.user_id = users.id").
		Where("pm.project_id = ? AND pm.deleted_at IS NULL", projectID).
		Where("users.deleted_at IS NULL").
		Where("LOWER(REPLACE(users.name, ' ', '')) IN ?", names).
		Find(&users).Error
	return users, err
}
