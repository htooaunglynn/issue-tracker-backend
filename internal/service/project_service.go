package service

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"gorm.io/gorm"
)

type ProjectService interface {
	ListProjects(callerID uuid.UUID, page, size int) ([]domain.ProjectResponse, int64, error)
	CreateProject(callerID uuid.UUID, req domain.CreateProjectRequest) (*domain.ProjectResponse, error)
	GetProject(callerID uuid.UUID, projectID uuid.UUID) (*domain.ProjectResponse, error)
	UpdateProject(callerID uuid.UUID, projectID uuid.UUID, req domain.UpdateProjectRequest) (*domain.ProjectResponse, error)
	DeleteProject(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID) error
	ArchiveProject(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID) error
	ListMembers(callerID uuid.UUID, projectID uuid.UUID, page, size int) ([]domain.ProjectMemberResponse, int64, error)
	AddMember(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, req domain.AddMemberRequest) (*domain.ProjectMemberResponse, error)
	UpdateMemberRole(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, targetUserID uuid.UUID, req domain.UpdateMemberRoleRequest) error
	RemoveMember(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, targetUserID uuid.UUID) error
}

type projectService struct {
	projectRepo       repository.ProjectRepository
	projectMemberRepo repository.ProjectMemberRepository
	userRepo          repository.UserRepository
}

func NewProjectService(
	projectRepo repository.ProjectRepository,
	projectMemberRepo repository.ProjectMemberRepository,
	userRepo repository.UserRepository,
) ProjectService {
	return &projectService{
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		userRepo:          userRepo,
	}
}

func (s *projectService) ListProjects(callerID uuid.UUID, page, size int) ([]domain.ProjectResponse, int64, error) {
	offset := (page - 1) * size
	projects, total, err := s.projectRepo.List(callerID, offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = domain.ProjectToResponse(&p)
	}
	return responses, total, nil
}

func (s *projectService) CreateProject(callerID uuid.UUID, req domain.CreateProjectRequest) (*domain.ProjectResponse, error) {
	if _, err := s.projectRepo.FindByKey(req.Key); err != gorm.ErrRecordNotFound {
		if err == nil {
			return nil, ErrProjectKeyTaken
		}
		return nil, err
	}

	project := &domain.Project{
		Name:        req.Name,
		Key:         req.Key,
		Description: req.Description,
		Status:      domain.ProjectStatusActive,
		OwnerID:     callerID,
	}

	if err := s.projectRepo.Create(project); err != nil {
		return nil, err
	}

	member := &domain.ProjectMember{
		ProjectID: project.ID,
		UserID:    callerID,
		Role:      domain.ProjectRoleAdmin,
	}
	if err := s.projectMemberRepo.Add(member); err != nil {
		return nil, err
	}

	resp := domain.ProjectToResponse(project)
	return &resp, nil
}

func (s *projectService) GetProject(callerID uuid.UUID, projectID uuid.UUID) (*domain.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if !s.isUserProjectMember(callerID, projectID) {
		return nil, ErrNotProjectMember
	}

	resp := domain.ProjectToResponse(project)
	return &resp, nil
}

func (s *projectService) UpdateProject(callerID uuid.UUID, projectID uuid.UUID, req domain.UpdateProjectRequest) (*domain.ProjectResponse, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if !s.canManageProject(callerID, project) {
		return nil, ErrForbidden
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := s.projectRepo.Update(project); err != nil {
		return nil, err
	}

	resp := domain.ProjectToResponse(project)
	return &resp, nil
}

func (s *projectService) DeleteProject(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProjectNotFound
		}
		return err
	}

	if !s.canManageProject(callerID, project) {
		return ErrForbidden
	}

	return s.projectRepo.Delete(projectID)
}

func (s *projectService) ArchiveProject(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProjectNotFound
		}
		return err
	}

	if !s.canManageProject(callerID, project) {
		return ErrForbidden
	}

	project.Status = domain.ProjectStatusArchived
	return s.projectRepo.Update(project)
}

func (s *projectService) ListMembers(callerID uuid.UUID, projectID uuid.UUID, page, size int) ([]domain.ProjectMemberResponse, int64, error) {
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, ErrProjectNotFound
		}
		return nil, 0, err
	}

	if !s.isUserProjectMember(callerID, projectID) {
		return nil, 0, ErrNotProjectMember
	}

	offset := (page - 1) * size
	members, total, err := s.projectMemberRepo.ListByProject(projectID, offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.ProjectMemberResponse, len(members))
	for i, m := range members {
		responses[i] = domain.ProjectMemberToResponse(&m, &m.User)
	}

	return responses, total, nil
}

func (s *projectService) AddMember(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, req domain.AddMemberRequest) (*domain.ProjectMemberResponse, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if !s.canManageProject(callerID, project) {
		return nil, ErrForbidden
	}

	if _, err := s.projectMemberRepo.FindByProjectAndUser(projectID, req.UserID); err != gorm.ErrRecordNotFound {
		if err == nil {
			return nil, ErrAlreadyMember
		}
		return nil, err
	}

	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	member := &domain.ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
	}

	if err := s.projectMemberRepo.Add(member); err != nil {
		return nil, err
	}

	resp := domain.ProjectMemberToResponse(member, user)
	return &resp, nil
}

func (s *projectService) UpdateMemberRole(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, targetUserID uuid.UUID, req domain.UpdateMemberRoleRequest) error {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProjectNotFound
		}
		return err
	}

	if !s.canManageProject(callerID, project) {
		return ErrForbidden
	}

	_, err = s.projectMemberRepo.FindByProjectAndUser(projectID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrMemberNotFound
		}
		return err
	}

	return s.projectMemberRepo.UpdateRole(projectID, targetUserID, req.Role)
}

func (s *projectService) RemoveMember(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, targetUserID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrProjectNotFound
		}
		return err
	}

	if !s.canManageProject(callerID, project) {
		return ErrForbidden
	}

	_, err = s.projectMemberRepo.FindByProjectAndUser(projectID, targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrMemberNotFound
		}
		return err
	}

	return s.projectMemberRepo.Remove(projectID, targetUserID)
}

func (s *projectService) isUserProjectMember(userID uuid.UUID, projectID uuid.UUID) bool {
	_, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	return err == nil
}

func (s *projectService) canManageProject(userID uuid.UUID, project *domain.Project) bool {
	if project.OwnerID == userID {
		return true
	}

	member, err := s.projectMemberRepo.FindByProjectAndUser(project.ID, userID)
	if err != nil {
		return false
	}

	return member.Role == domain.ProjectRoleAdmin || member.Role == domain.ProjectRoleManager
}
