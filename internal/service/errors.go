package service

import "errors"

var (
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("account is inactive")
	ErrTokenInvalid       = errors.New("token is invalid or expired")
	ErrInvalidPassword    = errors.New("current password is incorrect")
	ErrProjectNotFound    = errors.New("project not found")
	ErrProjectKeyTaken    = errors.New("project key already taken")
	ErrAlreadyMember      = errors.New("user is already a member of this project")
	ErrMemberNotFound     = errors.New("member not found")
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelNameTaken     = errors.New("label name already taken in this project")
	ErrNotProjectMember   = errors.New("user is not a member of this project")
	ErrForbidden          = errors.New("forbidden")
	ErrIssueNotFound      = errors.New("issue not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrFileTooLarge       = errors.New("file exceeds maximum allowed size")
)
