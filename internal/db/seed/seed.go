package seed

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
)

func Run(db *gorm.DB) error {
	log.Println("running database seed...")

	// --- Users ---
	users := []struct {
		name  string
		email string
		pass  string
		role  domain.GlobalRole
	}{
		{"Admin User", "admin@issuetracker.local", "Admin@1234!", domain.RoleAdmin},
		{"Project Manager", "manager@issuetracker.local", "Manager@1234!", domain.RoleManager},
		{"Developer One", "developer@issuetracker.local", "Dev@1234!", domain.RoleDeveloper},
		{"Reporter One", "reporter@issuetracker.local", "Reporter@1234!", domain.RoleReporter},
	}

	createdUsers := make(map[string]*domain.User)
	for _, u := range users {
		var existing domain.User
		if err := db.Where("email = ?", u.email).First(&existing).Error; err == nil {
			createdUsers[u.email] = &existing
			log.Printf("user %s already exists, skipping", u.email)
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.pass), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", u.email, err)
		}

		user := domain.User{
			BaseEntity:   domain.BaseEntity{ID: uuid.New()},
			Name:         u.name,
			Email:        u.email,
			PasswordHash: string(hash),
			GlobalRole:   u.role,
			IsActive:     true,
		}
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("create user %s: %w", u.email, err)
		}
		createdUsers[u.email] = &user
		log.Printf("created user %s", u.email)
	}

	adminUser := createdUsers["admin@issuetracker.local"]

	// --- Project ---
	var project domain.Project
	if err := db.Where("key = ?", "DEMO").First(&project).Error; err != nil {
		project = domain.Project{
			BaseEntity:  domain.BaseEntity{ID: uuid.New()},
			Key:         "DEMO",
			Name:        "Demo Project",
			Description: "A sample project for demonstration purposes",
			Status:      domain.ProjectStatusActive,
			OwnerID:     adminUser.ID,
		}
		if err := db.Create(&project).Error; err != nil {
			return fmt.Errorf("create project: %w", err)
		}
		log.Println("created demo project")
	}

	// --- Project Members ---
	memberEmails := []string{
		"manager@issuetracker.local",
		"developer@issuetracker.local",
		"reporter@issuetracker.local",
	}
	memberRoles := map[string]domain.ProjectRole{
		"manager@issuetracker.local":   domain.ProjectRoleManager,
		"developer@issuetracker.local": domain.ProjectRoleDeveloper,
		"reporter@issuetracker.local":  domain.ProjectRoleReporter,
	}
	for _, email := range memberEmails {
		u := createdUsers[email]
		var pm domain.ProjectMember
		if err := db.Where("project_id = ? AND user_id = ?", project.ID, u.ID).
			First(&pm).Error; err != nil {
			pm = domain.ProjectMember{
				BaseEntity: domain.BaseEntity{ID: uuid.New()},
				ProjectID:  project.ID,
				UserID:     u.ID,
				Role:       memberRoles[email],
			}
			if err := db.Create(&pm).Error; err != nil {
				return fmt.Errorf("add member %s: %w", email, err)
			}
		}
	}

	// --- Labels ---
	labelData := []struct {
		name  string
		color string
	}{
		{"bug", "#EF4444"},
		{"feature", "#3B82F6"},
		{"enhancement", "#10B981"},
		{"documentation", "#8B5CF6"},
		{"urgent", "#F59E0B"},
	}
	createdLabels := make(map[string]*domain.Label)
	for _, ld := range labelData {
		var label domain.Label
		if err := db.Where("project_id = ? AND name = ?", project.ID, ld.name).
			First(&label).Error; err != nil {
			label = domain.Label{
				BaseEntity: domain.BaseEntity{ID: uuid.New()},
				ProjectID:  project.ID,
				Name:       ld.name,
				Color:      ld.color,
			}
			if err := db.Create(&label).Error; err != nil {
				return fmt.Errorf("create label %s: %w", ld.name, err)
			}
		}
		createdLabels[ld.name] = &label
	}

	// --- Issues ---
	developer := createdUsers["developer@issuetracker.local"]
	devID := developer.ID
	issueData := []struct {
		number      int
		title       string
		description string
		itype       domain.IssueType
		priority    domain.IssuePriority
		status      domain.IssueStatus
		labels      []string
	}{
		{1, "Initial project setup", "Set up project structure and CI/CD pipeline",
			domain.IssueTypeTask, domain.IssuePriorityMedium, domain.IssueStatusClosed,
			[]string{"feature"}},
		{2, "Login page crashes on mobile", "Safari iOS 16 — login form submits but returns 500",
			domain.IssueTypeBug, domain.IssuePriorityCritical, domain.IssueStatusOpen,
			[]string{"bug", "urgent"}},
		{3, "Add dark mode support", "Implement system-preference-based dark mode toggle",
			domain.IssueTypeFeature, domain.IssuePriorityLow, domain.IssueStatusOpen,
			[]string{"feature", "enhancement"}},
		{4, "Write API documentation", "Document all REST endpoints with examples",
			domain.IssueTypeTask, domain.IssuePriorityMedium, domain.IssueStatusInProgress,
			[]string{"documentation"}},
	}

	reporter := createdUsers["reporter@issuetracker.local"]
	for _, id := range issueData {
		var existing domain.Issue
		if err := db.Where("project_id = ? AND number = ?", project.ID, id.number).
			First(&existing).Error; err == nil {
			continue
		}

		dueDate := time.Now().AddDate(0, 0, 14)
		issue := domain.Issue{
			BaseEntity:  domain.BaseEntity{ID: uuid.New()},
			ProjectID:   project.ID,
			Number:      id.number,
			Title:       id.title,
			Description: id.description,
			Type:        id.itype,
			Priority:    id.priority,
			Status:      id.status,
			ReporterID:  reporter.ID,
			AssigneeID:  &devID,
			DueDate:     &dueDate,
		}
		if err := db.Create(&issue).Error; err != nil {
			return fmt.Errorf("create issue %d: %w", id.number, err)
		}

		for _, lname := range id.labels {
			if l, ok := createdLabels[lname]; ok {
				il := domain.IssueLabel{IssueID: issue.ID, LabelID: l.ID}
				db.FirstOrCreate(&il, il)
			}
		}
		log.Printf("created issue #%d: %s", id.number, id.title)
	}

	log.Println("seed complete")
	return nil
}
