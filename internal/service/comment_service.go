package service

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/repository"
	"gorm.io/gorm"
)

type CommentService interface {
	ListComments(callerID uuid.UUID, issueID uuid.UUID, page, size int) ([]domain.CommentResponse, int64, error)
	CreateComment(callerID uuid.UUID, issueID uuid.UUID, req domain.CreateCommentRequest) (*domain.CommentResponse, error)
	UpdateComment(callerID uuid.UUID, commentID uuid.UUID, req domain.UpdateCommentRequest) (*domain.CommentResponse, error)
	DeleteComment(callerID uuid.UUID, commentID uuid.UUID) error
}

type commentService struct {
	db                *gorm.DB
	commentRepo       repository.CommentRepository
	issueRepo         repository.IssueRepository
	activityRepo      repository.ActivityRepository
	projectMemberRepo repository.ProjectMemberRepository
	userRepo          repository.UserRepository
}

func NewCommentService(
	db *gorm.DB,
	commentRepo repository.CommentRepository,
	issueRepo repository.IssueRepository,
	activityRepo repository.ActivityRepository,
	projectMemberRepo repository.ProjectMemberRepository,
	userRepo repository.UserRepository,
) CommentService {
	return &commentService{
		db:                db,
		commentRepo:       commentRepo,
		issueRepo:         issueRepo,
		activityRepo:      activityRepo,
		projectMemberRepo: projectMemberRepo,
		userRepo:          userRepo,
	}
}

func (s *commentService) ListComments(callerID uuid.UUID, issueID uuid.UUID, page, size int) ([]domain.CommentResponse, int64, error) {
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
	comments, total, err := s.commentRepo.ListByIssue(issueID, offset, size)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]domain.CommentResponse, len(comments))
	for i, c := range comments {
		mentions, _ := s.commentRepo.GetMentionedUsers(c.ID)
		responses[i] = domain.CommentToResponse(&c, mentions)
	}
	return responses, total, nil
}

func (s *commentService) CreateComment(callerID uuid.UUID, issueID uuid.UUID, req domain.CreateCommentRequest) (*domain.CommentResponse, error) {
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

	mentionedUserIDs, err := s.resolveMentions(req.Body, issue.ProjectID)
	if err != nil {
		return nil, err
	}

	comment := domain.Comment{
		IssueID:  issueID,
		AuthorID: callerID,
		Body:     req.Body,
	}

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.commentRepo.Create(tx, &comment); err != nil {
			return err
		}
		if len(mentionedUserIDs) > 0 {
			if err := s.commentRepo.CreateMentions(tx, comment.ID, mentionedUserIDs); err != nil {
				return err
			}
		}
		activity := &domain.IssueActivity{
			IssueID: issueID,
			ActorID: callerID,
			Action:  domain.ActivityCommented,
		}
		return s.activityRepo.Create(tx, activity)
	})
	if txErr != nil {
		return nil, txErr
	}

	full, err := s.commentRepo.FindByID(comment.ID)
	if err != nil {
		return nil, err
	}
	mentions, _ := s.commentRepo.GetMentionedUsers(comment.ID)
	resp := domain.CommentToResponse(full, mentions)
	return &resp, nil
}

func (s *commentService) UpdateComment(callerID uuid.UUID, commentID uuid.UUID, req domain.UpdateCommentRequest) (*domain.CommentResponse, error) {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCommentNotFound
		}
		return nil, err
	}

	if !s.canModifyComment(callerID, comment) {
		return nil, ErrForbidden
	}

	issue, err := s.issueRepo.FindByID(comment.IssueID)
	if err != nil {
		return nil, err
	}

	mentionedUserIDs, err := s.resolveMentions(req.Body, issue.ProjectID)
	if err != nil {
		return nil, err
	}

	comment.Body = req.Body

	txErr := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.commentRepo.Update(comment); err != nil {
			return err
		}
		if err := s.commentRepo.DeleteMentionsByComment(tx, comment.ID); err != nil {
			return err
		}
		if len(mentionedUserIDs) > 0 {
			if err := s.commentRepo.CreateMentions(tx, comment.ID, mentionedUserIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}

	mentions, _ := s.commentRepo.GetMentionedUsers(comment.ID)
	resp := domain.CommentToResponse(comment, mentions)
	return &resp, nil
}

func (s *commentService) DeleteComment(callerID uuid.UUID, commentID uuid.UUID) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrCommentNotFound
		}
		return err
	}

	if !s.canModifyComment(callerID, comment) {
		return ErrForbidden
	}

	return s.commentRepo.Delete(commentID)
}

// --- helpers ---

var mentionRegex = regexp.MustCompile(`@(\w+)`)

func parseMentionNames(body string) []string {
	matches := mentionRegex.FindAllStringSubmatch(body, -1)
	seen := make(map[string]struct{})
	var names []string
	for _, m := range matches {
		lower := strings.ToLower(m[1])
		if _, exists := seen[lower]; !exists {
			seen[lower] = struct{}{}
			names = append(names, lower)
		}
	}
	return names
}

func (s *commentService) resolveMentions(body string, projectID uuid.UUID) ([]uuid.UUID, error) {
	names := parseMentionNames(body)
	if len(names) == 0 {
		return nil, nil
	}
	users, err := s.userRepo.FindMentionedUsers(projectID, names)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	return ids, nil
}

func (s *commentService) isProjectMember(userID, projectID uuid.UUID) bool {
	_, err := s.projectMemberRepo.FindByProjectAndUser(projectID, userID)
	return err == nil
}

func (s *commentService) canModifyComment(callerID uuid.UUID, comment *domain.Comment) bool {
	if comment.AuthorID == callerID {
		return true
	}
	issue, err := s.issueRepo.FindByID(comment.IssueID)
	if err != nil {
		return false
	}
	member, err := s.projectMemberRepo.FindByProjectAndUser(issue.ProjectID, callerID)
	if err != nil {
		return false
	}
	return member.Role == domain.ProjectRoleAdmin || member.Role == domain.ProjectRoleManager
}
