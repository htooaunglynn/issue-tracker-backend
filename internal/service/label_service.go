package service

import (
	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"gorm.io/gorm"
)

type LabelService interface {
	ListLabels(callerID uuid.UUID, projectID uuid.UUID) ([]domain.LabelResponse, error)
	CreateLabel(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, req domain.CreateLabelRequest) (*domain.LabelResponse, error)
	UpdateLabel(callerID uuid.UUID, callerRole domain.GlobalRole, labelID uuid.UUID, req domain.UpdateLabelRequest) (*domain.LabelResponse, error)
	DeleteLabel(callerID uuid.UUID, callerRole domain.GlobalRole, labelID uuid.UUID) error
}

type labelService struct {
	labelRepo         repository.LabelRepository
	projectRepo       repository.ProjectRepository
	projectMemberRepo repository.ProjectMemberRepository
}

func NewLabelService(
	labelRepo repository.LabelRepository,
	projectRepo repository.ProjectRepository,
	projectMemberRepo repository.ProjectMemberRepository,
) LabelService {
	return &labelService{
		labelRepo:         labelRepo,
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
	}
}

func (s *labelService) ListLabels(callerID uuid.UUID, projectID uuid.UUID) ([]domain.LabelResponse, error) {
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if !s.isUserProjectMember(callerID, projectID) {
		return nil, ErrNotProjectMember
	}

	labels, err := s.labelRepo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.LabelResponse, len(labels))
	for i, l := range labels {
		responses[i] = domain.LabelToResponse(&l)
	}

	return responses, nil
}

func (s *labelService) CreateLabel(callerID uuid.UUID, callerRole domain.GlobalRole, projectID uuid.UUID, req domain.CreateLabelRequest) (*domain.LabelResponse, error) {
	_, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}

	if !s.canManageLabels(callerID, projectID) {
		return nil, ErrForbidden
	}

	if _, err := s.labelRepo.FindByNameAndProject(projectID, req.Name); err != gorm.ErrRecordNotFound {
		if err == nil {
			return nil, ErrLabelNameTaken
		}
		return nil, err
	}

	label := &domain.Label{
		ProjectID: projectID,
		Name:      req.Name,
		Color:     req.Color,
	}

	if err := s.labelRepo.Create(label); err != nil {
		return nil, err
	}

	resp := domain.LabelToResponse(label)
	return &resp, nil
}

func (s *labelService) UpdateLabel(callerID uuid.UUID, callerRole domain.GlobalRole, labelID uuid.UUID, req domain.UpdateLabelRequest) (*domain.LabelResponse, error) {
	label, err := s.labelRepo.FindByID(labelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrLabelNotFound
		}
		return nil, err
	}

	if !s.canManageLabels(callerID, label.ProjectID) {
		return nil, ErrForbidden
	}

	if req.Name != "" && req.Name != label.Name {
		if _, err := s.labelRepo.FindByNameAndProject(label.ProjectID, req.Name); err != gorm.ErrRecordNotFound {
			if err == nil {
				return nil, ErrLabelNameTaken
			}
			return nil, err
		}
		label.Name = req.Name
	}

	if req.Color != "" {
		label.Color = req.Color
	}

	if err := s.labelRepo.Update(label); err != nil {
		return nil, err
	}

	resp := domain.LabelToResponse(label)
	return &resp, nil
}

func (s *labelService) DeleteLabel(callerID uuid.UUID, callerRole domain.GlobalRole, labelID uuid.UUID) error {
	label, err := s.labelRepo.FindByID(labelID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrLabelNotFound
		}
		return err
	}

	if !s.canManageLabels(callerID, label.ProjectID) {
		return ErrForbidden
	}

	return s.labelRepo.Delete(labelID)
}

func (s *labelService) isUserProjectMember(userID uuid.UUID, projectID uuid.UUID) bool {
	_, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	return err == nil
}

func (s *labelService) canManageLabels(userID uuid.UUID, projectID uuid.UUID) bool {
	member, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	if err != nil {
		return false
	}

	return member.Role == domain.ProjectRoleAdmin || member.Role == domain.ProjectRoleManager
}
