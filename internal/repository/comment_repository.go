package repository

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"gorm.io/gorm"
)

type CommentRepository interface {
	Create(tx *gorm.DB, comment *domain.Comment) error
	FindByID(id uuid.UUID) (*domain.Comment, error)
	ListByIssue(issueID uuid.UUID, offset, limit int) ([]domain.Comment, int64, error)
	Update(comment *domain.Comment) error
	Delete(id uuid.UUID) error
	CreateMentions(tx *gorm.DB, commentID uuid.UUID, userIDs []uuid.UUID) error
	DeleteMentionsByComment(tx *gorm.DB, commentID uuid.UUID) error
	GetMentionedUsers(commentID uuid.UUID) ([]domain.User, error)
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(tx *gorm.DB, comment *domain.Comment) error {
	return tx.Create(comment).Error
}

func (r *commentRepository) FindByID(id uuid.UUID) (*domain.Comment, error) {
	var comment domain.Comment
	err := r.db.
		Preload("Author").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) ListByIssue(issueID uuid.UUID, offset, limit int) ([]domain.Comment, int64, error) {
	var comments []domain.Comment
	var total int64

	q := r.db.Model(&domain.Comment{}).
		Preload("Author").
		Where("issue_id = ? AND deleted_at IS NULL", issueID).
		Order("created_at ASC")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentRepository) Update(comment *domain.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uuid.UUID) error {
	return r.db.Model(&domain.Comment{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *commentRepository) CreateMentions(tx *gorm.DB, commentID uuid.UUID, userIDs []uuid.UUID) error {
	for _, userID := range userIDs {
		mention := domain.CommentMention{
			CommentID: commentID,
			UserID:    userID,
		}
		if err := tx.Create(&mention).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *commentRepository) DeleteMentionsByComment(tx *gorm.DB, commentID uuid.UUID) error {
	return tx.Where("comment_id = ?", commentID).Delete(&domain.CommentMention{}).Error
}

func (r *commentRepository) GetMentionedUsers(commentID uuid.UUID) ([]domain.User, error) {
	var users []domain.User
	err := r.db.
		Joins("JOIN comment_mentions cm ON cm.user_id = users.id").
		Where("cm.comment_id = ?", commentID).
		Where("users.deleted_at IS NULL").
		Find(&users).Error
	return users, err
}
