package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"gorm.io/gorm"
)

type IssueService interface {
	ListIssues(callerID uuid.UUID, projectID uuid.UUID, filter repository.IssueFilter, page, size int) ([]domain.IssueResponse, int64, error)
	CreateIssue(callerID uuid.UUID, projectID uuid.UUID, req domain.CreateIssueRequest) (*domain.IssueResponse, error)
	GetIssue(callerID uuid.UUID, issueID uuid.UUID) (*domain.IssueResponse, error)
	UpdateIssue(callerID uuid.UUID, issueID uuid.UUID, req domain.UpdateIssueRequest) (*domain.IssueResponse, error)
	DeleteIssue(callerID uuid.UUID, callerRole domain.GlobalRole, issueID uuid.UUID) error
	AssignIssue(callerID uuid.UUID, issueID uuid.UUID, req domain.AssignIssueRequest) (*domain.IssueResponse, error)
	ChangeStatus(callerID uuid.UUID, issueID uuid.UUID, req domain.ChangeStatusRequest) (*domain.IssueResponse, error)
	ListActivities(callerID uuid.UUID, issueID uuid.UUID, page, size int) ([]domain.IssueActivityResponse, int64, error)
}

type issueService struct {
	db                *gorm.DB
	issueRepo         repository.IssueRepository
	activityRepo      repository.ActivityRepository
	projectRepo       repository.ProjectRepository
	projectMemberRepo repository.ProjectMemberRepository
	labelRepo         repository.LabelRepository
}

func NewIssueService(
	db *gorm.DB,
	issueRepo repository.IssueRepository,
	activityRepo repository.ActivityRepository,
	projectRepo repository.ProjectRepository,
	projectMemberRepo repository.ProjectMemberRepository,
	labelRepo repository.LabelRepository,
) IssueService {
	return &issueService{
		db:                db,
		issueRepo:         issueRepo,
		activityRepo:      activityRepo,
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		labelRepo:         labelRepo,
	}
}

func (s *issueService) ListIssues(callerID uuid.UUID, projectID uuid.UUID, filter repository.IssueFilter, page, size int) ([]domain.IssueResponse, int64, error) {
	if _, err := s.projectRepo.FindByID(projectID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrProjectNotFound
		}
		return nil, 0, err
	}

	if !s.isProjectMember(callerID, projectID) {
		return nil, 0, ErrNotProjectMember
	}

	offset := (page - 1) * size
	issues, total, err := s.issueRepo.List(projectID, filter, offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.IssueResponse, len(issues))
	for i, issue := range issues {
		labels, _ := s.issueRepo.GetLabels(issue.ID)
		responses[i] = domain.IssueToResponse(&issue, labels)
	}
	return responses, total, nil
}

func (s *issueService) CreateIssue(callerID uuid.UUID, projectID uuid.UUID, req domain.CreateIssueRequest) (*domain.IssueResponse, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if !s.isProjectMember(callerID, projectID) {
		return nil, ErrNotProjectMember
	}

	var issue domain.Issue
	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		number, err := s.issueRepo.NextNumber(tx, projectID)
		if err != nil {
			return err
		}

		issue = domain.Issue{
			ProjectID:   projectID,
			Number:      number,
			Title:       req.Title,
			Description: req.Description,
			Type:        req.Type,
			Priority:    req.Priority,
			Status:      domain.IssueStatusOpen,
			ReporterID:  callerID,
			AssigneeID:  req.AssigneeID,
			DueDate:     req.DueDate,
		}

		if err := s.issueRepo.Create(tx, &issue); err != nil {
			return err
		}

		if len(req.LabelIDs) > 0 {
			if err := s.issueRepo.SetLabels(tx, issue.ID, req.LabelIDs); err != nil {
				return err
			}
		}

		activity := &domain.IssueActivity{
			IssueID: issue.ID,
			ActorID: callerID,
			Action:  domain.ActivityCreated,
		}
		return s.activityRepo.Create(tx, activity)
	})
	if txErr != nil {
		return nil, txErr
	}

	// Reload with associations
	full, err := s.issueRepo.FindByID(issue.ID)
	if err != nil {
		return nil, err
	}
	full.Project = *project
	labels, _ := s.issueRepo.GetLabels(full.ID)
	resp := domain.IssueToResponse(full, labels)
	return &resp, nil
}

func (s *issueService) GetIssue(callerID uuid.UUID, issueID uuid.UUID) (*domain.IssueResponse, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrIssueNotFound
		}
		return nil, err
	}

	if !s.isProjectMember(callerID, issue.ProjectID) {
		return nil, ErrNotProjectMember
	}

	labels, _ := s.issueRepo.GetLabels(issue.ID)
	resp := domain.IssueToResponse(issue, labels)
	return &resp, nil
}

func (s *issueService) UpdateIssue(callerID uuid.UUID, issueID uuid.UUID, req domain.UpdateIssueRequest) (*domain.IssueResponse, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrIssueNotFound
		}
		return nil, err
	}

	if !s.canEditIssue(callerID, issue) {
		return nil, ErrForbidden
	}

	var changedFields []string

	if req.Title != "" && req.Title != issue.Title {
		changedFields = append(changedFields, fmt.Sprintf("title: %q → %q", issue.Title, req.Title))
		issue.Title = req.Title
	}
	if req.Description != nil {
		issue.Description = *req.Description
	}
	if req.Type != "" {
		issue.Type = req.Type
	}
	if req.Priority != "" {
		issue.Priority = req.Priority
	}
	if req.DueDate != nil {
		issue.DueDate = req.DueDate
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.issueRepo.Update(issue); err != nil {
			return err
		}

		if req.LabelIDs != nil {
			if err := s.issueRepo.SetLabels(tx, issue.ID, *req.LabelIDs); err != nil {
				return err
			}
		}

		if len(changedFields) > 0 {
			activity := &domain.IssueActivity{
				IssueID: issue.ID,
				ActorID: callerID,
				Action:  domain.ActivityUpdated,
			}
			return s.activityRepo.Create(tx, activity)
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	labels, _ := s.issueRepo.GetLabels(issue.ID)
	resp := domain.IssueToResponse(issue, labels)
	return &resp, nil
}

func (s *issueService) DeleteIssue(callerID uuid.UUID, callerRole domain.GlobalRole, issueID uuid.UUID) error {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrIssueNotFound
		}
		return err
	}

	effective := s.effectiveRole(callerID, callerRole, issue.ProjectID)
	if effective != domain.ProjectRoleAdmin && effective != domain.ProjectRoleManager {
		return ErrForbidden
	}

	return s.issueRepo.Delete(issueID)
}

func (s *issueService) AssignIssue(callerID uuid.UUID, issueID uuid.UUID, req domain.AssignIssueRequest) (*domain.IssueResponse, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrIssueNotFound
		}
		return nil, err
	}

	effective := s.effectiveRole(callerID, domain.GlobalRole(""), issue.ProjectID)
	if effective == domain.ProjectRoleReporter {
		return nil, ErrForbidden
	}

	oldAssignee := ""
	if issue.AssigneeID != nil {
		oldAssignee = issue.AssigneeID.String()
	}
	newAssignee := ""
	if req.AssigneeID != nil {
		newAssignee = req.AssigneeID.String()
	}

	issue.AssigneeID = req.AssigneeID

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.issueRepo.Update(issue); err != nil {
			return err
		}
		field := "assignee_id"
		activity := &domain.IssueActivity{
			IssueID:  issue.ID,
			ActorID:  callerID,
			Action:   domain.ActivityAssigned,
			Field:    &field,
			OldValue: &oldAssignee,
			NewValue: &newAssignee,
		}
		return s.activityRepo.Create(tx, activity)
	})
	if txErr != nil {
		return nil, txErr
	}

	labels, _ := s.issueRepo.GetLabels(issue.ID)
	resp := domain.IssueToResponse(issue, labels)
	return &resp, nil
}

func (s *issueService) ChangeStatus(callerID uuid.UUID, issueID uuid.UUID, req domain.ChangeStatusRequest) (*domain.IssueResponse, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrIssueNotFound
		}
		return nil, err
	}

	if !s.canChangeStatus(callerID, issue, req.Status) {
		return nil, ErrForbidden
	}

	oldStatus := string(issue.Status)
	newStatus := string(req.Status)
	issue.Status = req.Status

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.issueRepo.Update(issue); err != nil {
			return err
		}
		field := "status"
		activity := &domain.IssueActivity{
			IssueID:  issue.ID,
			ActorID:  callerID,
			Action:   domain.ActivityStatusChanged,
			Field:    &field,
			OldValue: &oldStatus,
			NewValue: &newStatus,
		}
		return s.activityRepo.Create(tx, activity)
	})
	if txErr != nil {
		return nil, txErr
	}

	labels, _ := s.issueRepo.GetLabels(issue.ID)
	resp := domain.IssueToResponse(issue, labels)
	return &resp, nil
}

func (s *issueService) ListActivities(callerID uuid.UUID, issueID uuid.UUID, page, size int) ([]domain.IssueActivityResponse, int64, error) {
	issue, err := s.issueRepo.FindByID(issueID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrIssueNotFound
		}
		return nil, 0, err
	}

	if !s.isProjectMember(callerID, issue.ProjectID) {
		return nil, 0, ErrNotProjectMember
	}

	offset := (page - 1) * size
	activities, total, err := s.activityRepo.ListByIssue(issueID, offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.IssueActivityResponse, len(activities))
	for i, a := range activities {
		responses[i] = domain.IssueActivityToResponse(&a)
	}
	return responses, total, nil
}

// --- Permission helpers ---

func (s *issueService) isProjectMember(userID, projectID uuid.UUID) bool {
	_, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	return err == nil
}

// effectiveRole returns the caller's project-level effective role.
// Global admin always maps to project admin.
func (s *issueService) effectiveRole(userID uuid.UUID, globalRole domain.GlobalRole, projectID uuid.UUID) domain.ProjectRole {
	if globalRole == domain.RoleAdmin {
		return domain.ProjectRoleAdmin
	}
	member, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	if err != nil {
		return domain.ProjectRoleReporter
	}
	return member.Role
}

func (s *issueService) canEditIssue(callerID uuid.UUID, issue *domain.Issue) bool {
	member, err := s.projectMemberRepo.FindByProjectAndUser(issue.ProjectID, callerID)
	if err != nil {
		return false
	}
	switch member.Role {
	case domain.ProjectRoleAdmin, domain.ProjectRoleManager:
		return true
	case domain.ProjectRoleDeveloper:
		return issue.ReporterID == callerID || (issue.AssigneeID != nil && *issue.AssigneeID == callerID)
	case domain.ProjectRoleReporter:
		return issue.ReporterID == callerID
	}
	return false
}

func (s *issueService) canChangeStatus(callerID uuid.UUID, issue *domain.Issue, newStatus domain.IssueStatus) bool {
	member, err := s.projectMemberRepo.FindByProjectAndUser(issue.ProjectID, callerID)
	if err != nil {
		return false
	}
	switch member.Role {
	case domain.ProjectRoleAdmin, domain.ProjectRoleManager, domain.ProjectRoleDeveloper:
		return true
	case domain.ProjectRoleReporter:
		// Reporters can only close their own issues
		return issue.ReporterID == callerID && newStatus == domain.IssueStatusClosed
	}
	return false
}
