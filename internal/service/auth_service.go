package service

import (
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/config"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/hash"
	pkgjwt "github.com/htooaunglynn/issue-tracker-backend/internal/pkg/jwt"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(req domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(req domain.LoginRequest) (*domain.AuthResponse, error)
	RefreshTokens(refreshToken string) (*domain.AuthResponse, error)
	Logout(refreshToken string) error
	GetMe(userID uuid.UUID) (*domain.UserResponse, error)
	ForgotPassword(email string) error
	ResetPassword(req domain.ResetPasswordRequest) error
	UpdateProfile(userID uuid.UUID, req domain.UpdateProfileRequest) (*domain.UserResponse, error)
	ChangePassword(userID uuid.UUID, req domain.ChangePasswordRequest) error
}

type authService struct {
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	passwordResetRepo repository.PasswordResetRepository
	jwtCfg           config.JWTConfig
	auditDB          *gorm.DB
}

func NewAuthService(
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	passwordResetRepo repository.PasswordResetRepository,
	jwtCfg config.JWTConfig,
	db *gorm.DB,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordResetRepo: passwordResetRepo,
		jwtCfg:           jwtCfg,
		auditDB:          db,
	}
}

func (s *authService) Register(req domain.RegisterRequest) (*domain.AuthResponse, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := hash.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		BaseEntity:   domain.BaseEntity{ID: uuid.New()},
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		GlobalRole:   domain.RoleDeveloper,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	accessToken, err := pkgjwt.GenerateAccessToken(user.ID, user.GlobalRole, s.jwtCfg.Secret, s.jwtCfg.AccessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := pkgjwt.GenerateRefreshToken(user.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshExpiry)
	if err != nil {
		return nil, err
	}

	tokenHash := pkgjwt.HashToken(refreshToken)
	rt := &domain.RefreshToken{
		BaseEntity: domain.BaseEntity{ID: uuid.New()},
		UserID:     user.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().Add(s.jwtCfg.RefreshExpiry),
		Revoked:    false,
	}

	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return nil, err
	}

	s.logAuditEvent("register", "user", user.ID.String(), user.Email)

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         domain.UserToResponse(user),
	}, nil
}

func (s *authService) Login(req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := hash.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := pkgjwt.GenerateAccessToken(user.ID, user.GlobalRole, s.jwtCfg.Secret, s.jwtCfg.AccessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := pkgjwt.GenerateRefreshToken(user.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshExpiry)
	if err != nil {
		return nil, err
	}

	tokenHash := pkgjwt.HashToken(refreshToken)
	rt := &domain.RefreshToken{
		BaseEntity: domain.BaseEntity{ID: uuid.New()},
		UserID:     user.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().Add(s.jwtCfg.RefreshExpiry),
		Revoked:    false,
	}

	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return nil, err
	}

	s.logAuditEvent("login", "user", user.ID.String(), user.Email)

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         domain.UserToResponse(user),
	}, nil
}

func (s *authService) RefreshTokens(refreshToken string) (*domain.AuthResponse, error) {
	claims, err := pkgjwt.ValidateToken(refreshToken, s.jwtCfg.Secret)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	tokenHash := pkgjwt.HashToken(refreshToken)
	if _, err := s.refreshTokenRepo.FindByTokenHash(tokenHash); err != nil {
		return nil, ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := s.refreshTokenRepo.RevokeByTokenHash(tokenHash); err != nil {
		return nil, err
	}

	accessToken, err := pkgjwt.GenerateAccessToken(user.ID, user.GlobalRole, s.jwtCfg.Secret, s.jwtCfg.AccessExpiry)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := pkgjwt.GenerateRefreshToken(user.ID, s.jwtCfg.Secret, s.jwtCfg.RefreshExpiry)
	if err != nil {
		return nil, err
	}

	newTokenHash := pkgjwt.HashToken(newRefreshToken)
	newRT := &domain.RefreshToken{
		BaseEntity: domain.BaseEntity{ID: uuid.New()},
		UserID:     user.ID,
		TokenHash:  newTokenHash,
		ExpiresAt:  time.Now().Add(s.jwtCfg.RefreshExpiry),
		Revoked:    false,
	}

	if err := s.refreshTokenRepo.Create(newRT); err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         domain.UserToResponse(user),
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	tokenHash := pkgjwt.HashToken(refreshToken)
	if err := s.refreshTokenRepo.RevokeByTokenHash(tokenHash); err != nil {
		return err
	}

	claims, _ := pkgjwt.ValidateToken(refreshToken, s.jwtCfg.Secret)
	if claims != nil {
		s.logAuditEvent("logout", "user", claims.UserID.String(), "")
	}

	return nil
}

func (s *authService) GetMe(userID uuid.UUID) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	resp := domain.UserToResponse(user)
	return &resp, nil
}

func (s *authService) ForgotPassword(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil
	}

	token := generateRandomToken(32)
	tokenHash := pkgjwt.HashToken(token)

	reset := &domain.PasswordReset{
		BaseEntity: domain.BaseEntity{ID: uuid.New()},
		UserID:     user.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		Used:       false,
	}

	if err := s.passwordResetRepo.Create(reset); err != nil {
		return nil
	}

	log.Printf("password reset token for %s: %s", email, token)

	return nil
}

func (s *authService) ResetPassword(req domain.ResetPasswordRequest) error {
	tokenHash := pkgjwt.HashToken(req.Token)
	reset, err := s.passwordResetRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return ErrTokenInvalid
	}

	user, err := s.userRepo.FindByID(reset.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	newHash, err := hash.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	if err := s.passwordResetRepo.MarkUsed(reset.ID); err != nil {
		return err
	}

	if err := s.refreshTokenRepo.RevokeAllForUser(user.ID); err != nil {
		return err
	}

	s.logAuditEvent("password_reset", "user", user.ID.String(), "")

	return nil
}

func (s *authService) UpdateProfile(userID uuid.UUID, req domain.UpdateProfileRequest) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
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

func (s *authService) ChangePassword(userID uuid.UUID, req domain.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	if err := hash.Compare(user.PasswordHash, req.CurrentPassword); err != nil {
		return ErrInvalidPassword
	}

	newHash, err := hash.Hash(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	if err := s.refreshTokenRepo.RevokeAllForUser(user.ID); err != nil {
		return err
	}

	s.logAuditEvent("password_changed", "user", user.ID.String(), "")

	return nil
}

func (s *authService) logAuditEvent(action, resourceType, resourceID, metadata string) {
	auditLog := domain.AuditLog{
		BaseEntity:   domain.BaseEntity{ID: uuid.New()},
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
	}
	s.auditDB.Create(&auditLog)
}

func generateRandomToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return fmt.Sprintf("%x", b)
}
