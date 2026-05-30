package domain

import (
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Request DTOs
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
	Name      string  `json:"name"       validate:"omitempty,min=2,max=100"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"        validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"        validate:"required,min=1,max=255"`
	Key         string `json:"key"         validate:"required,min=2,max=10,alphanum"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"         validate:"omitempty,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=1000"`
}

type AddMemberRequest struct {
	UserID uuid.UUID   `json:"user_id" validate:"required"`
	Role   ProjectRole `json:"role"     validate:"required"`
}

type UpdateMemberRoleRequest struct {
	Role ProjectRole `json:"role" validate:"required"`
}

type CreateLabelRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=255"`
	Color string `json:"color" validate:"required,hexcolor"`
}

type UpdateLabelRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=1,max=255"`
	Color string `json:"color" validate:"omitempty,hexcolor"`
}

// Response DTOs
type UserResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	GlobalRole GlobalRole `json:"global_role"`
	AvatarURL  *string    `json:"avatar_url"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
}

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type ProjectResponse struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Key         string        `json:"key"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status"`
	OwnerID     uuid.UUID     `json:"owner_id"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type ProjectMemberResponse struct {
	UserID    uuid.UUID   `json:"user_id"`
	Name      string      `json:"name"`
	Email     string      `json:"email"`
	AvatarURL *string     `json:"avatar_url"`
	Role      ProjectRole `json:"role"`
}

type LabelResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// Comment request DTOs
type CreateCommentRequest struct {
	Body string `json:"body" validate:"required,min=1,max=10000"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" validate:"required,min=1,max=10000"`
}

// Comment response DTO
type CommentResponse struct {
	ID        uuid.UUID    `json:"id"`
	IssueID   uuid.UUID    `json:"issue_id"`
	Author    UserResponse `json:"author"`
	Body      string       `json:"body"`
	Mentions  []UserResponse `json:"mentions"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Attachment response DTO
type AttachmentResponse struct {
	ID        uuid.UUID    `json:"id"`
	IssueID   uuid.UUID    `json:"issue_id"`
	Uploader  UserResponse `json:"uploader"`
	FileName  string       `json:"file_name"`
	MimeType  string       `json:"mime_type"`
	SizeBytes int64        `json:"size_bytes"`
	CreatedAt time.Time    `json:"created_at"`
}

func CommentToResponse(c *Comment, mentions []User) CommentResponse {
	mentionResponses := make([]UserResponse, len(mentions))
	for i, u := range mentions {
		mentionResponses[i] = UserToResponse(&u)
	}
	return CommentResponse{
		ID:        c.ID,
		IssueID:   c.IssueID,
		Author:    UserToResponse(&c.Author),
		Body:      c.Body,
		Mentions:  mentionResponses,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func AttachmentToResponse(a *Attachment) AttachmentResponse {
	return AttachmentResponse{
		ID:        a.ID,
		IssueID:   a.IssueID,
		Uploader:  UserToResponse(&a.Uploader),
		FileName:  a.FileName,
		MimeType:  a.MimeType,
		SizeBytes: a.SizeBytes,
		CreatedAt: a.CreatedAt,
	}
}

// Issue request DTOs
type CreateIssueRequest struct {
	Title       string        `json:"title"        validate:"required,min=1,max=255"`
	Description string        `json:"description"  validate:"omitempty"`
	Type        IssueType     `json:"type"         validate:"required,oneof=bug feature task improvement"`
	Priority    IssuePriority `json:"priority"     validate:"required,oneof=low medium high critical"`
	AssigneeID  *uuid.UUID    `json:"assignee_id"  validate:"omitempty"`
	DueDate     *time.Time    `json:"due_date"     validate:"omitempty"`
	LabelIDs    []uuid.UUID   `json:"label_ids"    validate:"omitempty"`
}

type UpdateIssueRequest struct {
	Title       string        `json:"title"        validate:"omitempty,min=1,max=255"`
	Description *string       `json:"description"  validate:"omitempty"`
	Type        IssueType     `json:"type"         validate:"omitempty,oneof=bug feature task improvement"`
	Priority    IssuePriority `json:"priority"     validate:"omitempty,oneof=low medium high critical"`
	DueDate     *time.Time    `json:"due_date"     validate:"omitempty"`
	LabelIDs    *[]uuid.UUID  `json:"label_ids"    validate:"omitempty"`
}

type AssignIssueRequest struct {
	AssigneeID *uuid.UUID `json:"assignee_id"`
}

type ChangeStatusRequest struct {
	Status IssueStatus `json:"status" validate:"required,oneof=open in_progress resolved closed reopened"`
}

// Issue response DTOs
type IssueResponse struct {
	ID          uuid.UUID      `json:"id"`
	ProjectID   uuid.UUID      `json:"project_id"`
	ProjectKey  string         `json:"project_key"`
	Number      int            `json:"number"`
	Ref         string         `json:"ref"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        IssueType      `json:"type"`
	Priority    IssuePriority  `json:"priority"`
	Status      IssueStatus    `json:"status"`
	Reporter    UserResponse   `json:"reporter"`
	Assignee    *UserResponse  `json:"assignee"`
	Labels      []LabelResponse `json:"labels"`
	DueDate     *time.Time     `json:"due_date"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type IssueActivityResponse struct {
	ID        uuid.UUID      `json:"id"`
	IssueID   uuid.UUID      `json:"issue_id"`
	Actor     UserResponse   `json:"actor"`
	Action    ActivityAction `json:"action"`
	Field     *string        `json:"field"`
	OldValue  *string        `json:"old_value"`
	NewValue  *string        `json:"new_value"`
	CreatedAt time.Time      `json:"created_at"`
}

func IssueToResponse(issue *Issue, labels []Label) IssueResponse {
	r := IssueResponse{
		ID:          issue.ID,
		ProjectID:   issue.ProjectID,
		ProjectKey:  issue.Project.Key,
		Number:      issue.Number,
		Ref:         issue.Project.Key + "-" + strconv.Itoa(issue.Number),
		Title:       issue.Title,
		Description: issue.Description,
		Type:        issue.Type,
		Priority:    issue.Priority,
		Status:      issue.Status,
		Reporter:    UserToResponse(&issue.Reporter),
		DueDate:     issue.DueDate,
		CreatedAt:   issue.CreatedAt,
		UpdatedAt:   issue.UpdatedAt,
	}
	if issue.Assignee != nil {
		u := UserToResponse(issue.Assignee)
		r.Assignee = &u
	}
	r.Labels = make([]LabelResponse, len(labels))
	for i, l := range labels {
		r.Labels[i] = LabelToResponse(&l)
	}
	return r
}

func IssueActivityToResponse(a *IssueActivity) IssueActivityResponse {
	return IssueActivityResponse{
		ID:        a.ID,
		IssueID:   a.IssueID,
		Actor:     UserToResponse(&a.Actor),
		Action:    a.Action,
		Field:     a.Field,
		OldValue:  a.OldValue,
		NewValue:  a.NewValue,
		CreatedAt: a.CreatedAt,
	}
}

// Mappers
func UserToResponse(u *User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		Name:       u.Name,
		Email:      u.Email,
		GlobalRole: u.GlobalRole,
		AvatarURL:  u.AvatarURL,
		IsActive:   u.IsActive,
		CreatedAt:  u.CreatedAt,
	}
}

func ProjectToResponse(p *Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
		Status:      p.Status,
		OwnerID:     p.OwnerID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func ProjectMemberToResponse(pm *ProjectMember, user *User) ProjectMemberResponse {
	return ProjectMemberResponse{
		UserID:    pm.UserID,
		Name:      user.Name,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
		Role:      pm.Role,
	}
}

func LabelToResponse(l *Label) LabelResponse {
	return LabelResponse{
		ID:        l.ID,
		ProjectID: l.ProjectID,
		Name:      l.Name,
		Color:     l.Color,
		CreatedAt: l.CreatedAt,
	}
}
