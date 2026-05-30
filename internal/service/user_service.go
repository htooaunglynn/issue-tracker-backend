package service

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
)

type UserService interface {
	ListUsers(page, size int) ([]domain.UserResponse, int64, error)
	GetUser(id uuid.UUID) (*domain.UserResponse, error)
	UpdateUser(id uuid.UUID, req domain.UpdateProfileRequest) (*domain.UserResponse, error)
	DeleteUser(id uuid.UUID) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) ListUsers(page, size int) ([]domain.UserResponse, int64, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.List(offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.UserResponse, len(users))
	for i, u := range users {
		responses[i] = domain.UserToResponse(&u)
	}

	return responses, total, nil
}

func (s *userService) GetUser(id uuid.UUID) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	resp := domain.UserToResponse(user)
	return &resp, nil
}

func (s *userService) UpdateUser(id uuid.UUID, req domain.UpdateProfileRequest) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	resp := domain.UserToResponse(user)
	return &resp, nil
}

func (s *userService) DeleteUser(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}
