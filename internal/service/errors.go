package service

import "errors"

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("account is inactive")
	ErrTokenInvalid       = errors.New("token is invalid or expired")
	ErrInvalidPassword    = errors.New("current password is incorrect")
)
