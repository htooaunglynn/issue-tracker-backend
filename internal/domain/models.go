package domain

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseEntity contains common fields for all entities
type BaseEntity struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

// BeforeCreate sets the UUID if not already set
func (b *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// User entity
type User struct {
	BaseEntity
	Name         string `gorm:"index"`
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string `gorm:"type:text"`
	GlobalRole   GlobalRole
	AvatarURL    *string
	IsActive     bool `gorm:"default:true"`
}

type GlobalRole string

const (
	RoleAdmin     GlobalRole = "admin"
	RoleManager   GlobalRole = "manager"
	RoleDeveloper GlobalRole = "developer"
	RoleReporter  GlobalRole = "reporter"
)

// Scan implements sql.Scanner interface
func (r *GlobalRole) Scan(value interface{}) error {
	*r = GlobalRole(value.(string))
	return nil
}

// Value implements driver.Valuer interface
func (r GlobalRole) Value() (driver.Value, error) {
	return string(r), nil
}

// Project entity
type Project struct {
	BaseEntity
	Key         string `gorm:"uniqueIndex;size:10"`
	Name        string `gorm:"index"`
	Description string
	Status      ProjectStatus
	OwnerID     uuid.UUID
	Owner       User `gorm:"foreignKey:OwnerID"`
}

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusArchived ProjectStatus = "archived"
)

// ProjectMember entity
type ProjectMember struct {
	BaseEntity
	ProjectID uuid.UUID
	Project   Project `gorm:"foreignKey:ProjectID"`
	UserID    uuid.UUID
	User      User `gorm:"foreignKey:UserID"`
	Role      ProjectRole
}

type ProjectRole string

const (
	ProjectRoleAdmin     ProjectRole = "admin"
	ProjectRoleManager   ProjectRole = "manager"
	ProjectRoleDeveloper ProjectRole = "developer"
	ProjectRoleReporter  ProjectRole = "reporter"
)

// Issue entity
type Issue struct {
	BaseEntity
	ProjectID   uuid.UUID
	Project     Project `gorm:"foreignKey:ProjectID;index"`
	Number      int
	Title       string `gorm:"index"`
	Description string
	Type        IssueType
	Priority    IssuePriority `gorm:"index"`
	Status      IssueStatus   `gorm:"index"`
	ReporterID  uuid.UUID
	Reporter    User `gorm:"foreignKey:ReporterID"`
	AssigneeID  *uuid.UUID
	Assignee    *User `gorm:"foreignKey:AssigneeID"`
	DueDate     *time.Time
}

type IssueType string

const (
	IssueTypeBug         IssueType = "bug"
	IssueTypeFeature     IssueType = "feature"
	IssueTypeTask        IssueType = "task"
	IssueTypeImprovement IssueType = "improvement"
)

type IssuePriority string

const (
	IssuePriorityLow      IssuePriority = "low"
	IssuePriorityMedium   IssuePriority = "medium"
	IssuePriorityHigh     IssuePriority = "high"
	IssuePriorityCritical IssuePriority = "critical"
)

type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusResolved   IssueStatus = "resolved"
	IssueStatusClosed     IssueStatus = "closed"
	IssueStatusReopened   IssueStatus = "reopened"
)

// Label entity
type Label struct {
	BaseEntity
	ProjectID uuid.UUID
	Project   Project `gorm:"foreignKey:ProjectID"`
	Name      string
	Color     string
}

// IssueLabel join table
type IssueLabel struct {
	IssueID uuid.UUID `gorm:"primaryKey;foreignKey:IssueID"`
	LabelID uuid.UUID `gorm:"primaryKey;foreignKey:LabelID"`
}

// Comment entity
type Comment struct {
	BaseEntity
	IssueID  uuid.UUID
	Issue    Issue `gorm:"foreignKey:IssueID"`
	AuthorID uuid.UUID
	Author   User   `gorm:"foreignKey:AuthorID"`
	Body     string `gorm:"type:text"`
}

// CommentMention tracks @mentions within a comment
type CommentMention struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	CommentID uuid.UUID `gorm:"index"`
	UserID    uuid.UUID
	User      User      `gorm:"foreignKey:UserID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// Attachment entity
type Attachment struct {
	BaseEntity
	IssueID    uuid.UUID
	Issue      Issue `gorm:"foreignKey:IssueID"`
	UploaderID uuid.UUID
	Uploader   User `gorm:"foreignKey:UploaderID"`
	FileName   string
	FilePath   string
	MimeType   string
	SizeBytes  int64
}

// IssueActivity entity (audit/history log)
type IssueActivity struct {
	BaseEntity
	IssueID  uuid.UUID
	Issue    Issue `gorm:"foreignKey:IssueID;index"`
	ActorID  uuid.UUID
	Actor    User `gorm:"foreignKey:ActorID"`
	Action   ActivityAction
	Field    *string
	OldValue *string
	NewValue *string
}

type ActivityAction string

const (
	ActivityCreated       ActivityAction = "created"
	ActivityUpdated       ActivityAction = "updated"
	ActivityStatusChanged ActivityAction = "status_changed"
	ActivityAssigned      ActivityAction = "assigned"
	ActivityCommented     ActivityAction = "commented"
)

// Notification entity
type Notification struct {
	BaseEntity
	UserID     uuid.UUID
	User       User `gorm:"foreignKey:UserID;index"`
	Type       NotificationType
	Title      string
	Message    string
	EntityType string
	EntityID   uuid.UUID
	IsRead     bool `gorm:"index"`
}

type NotificationType string

const (
	NotificationAssignment   NotificationType = "assignment"
	NotificationComment      NotificationType = "comment"
	NotificationStatusChange NotificationType = "status_change"
	NotificationMention      NotificationType = "mention"
)

// AuditLog entity
type AuditLog struct {
	BaseEntity
	ActorID      *uuid.UUID
	Actor        *User `gorm:"foreignKey:ActorID"`
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     string `gorm:"type:jsonb"`
	IP           string
	UserAgent    string
}

// RefreshToken entity
type RefreshToken struct {
	BaseEntity
	UserID    uuid.UUID
	User      User   `gorm:"foreignKey:UserID;index"`
	TokenHash string `gorm:"index"`
	ExpiresAt time.Time
	Revoked   bool
}

// PasswordReset entity
type PasswordReset struct {
	BaseEntity
	UserID    uuid.UUID
	User      User   `gorm:"foreignKey:UserID"`
	TokenHash string `gorm:"index"`
	ExpiresAt time.Time
	Used      bool
}
