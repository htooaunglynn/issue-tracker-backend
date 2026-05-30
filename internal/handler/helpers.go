package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
	"github.com/htooaunglynn/issue-tracker-backend/internal/service"
)

func bindAndValidate(c *gin.Context, v *validator.Validate, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorSimple("BAD_REQUEST", err.Error()))
		return false
	}

	if err := v.Struct(req); err != nil {
		var details []response.ErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			details = append(details, response.ErrorDetail{
				Field:   fe.Field(),
				Message: fe.Tag(),
			})
		}
		c.JSON(http.StatusUnprocessableEntity, response.Error("VALIDATION_ERROR", "validation failed", details))
		return false
	}

	return true
}

func handleServiceError(c *gin.Context, err error) {
	switch err {
	case service.ErrEmailTaken:
		c.JSON(http.StatusConflict, response.ErrorSimple("EMAIL_TAKEN", err.Error()))
	case service.ErrInvalidCredentials:
		c.JSON(http.StatusUnauthorized, response.ErrorSimple("INVALID_CREDENTIALS", err.Error()))
	case service.ErrUserNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrUserInactive:
		c.JSON(http.StatusForbidden, response.ErrorSimple("ACCOUNT_INACTIVE", err.Error()))
	case service.ErrTokenInvalid:
		c.JSON(http.StatusUnauthorized, response.ErrorSimple("TOKEN_INVALID", err.Error()))
	case service.ErrInvalidPassword:
		c.JSON(http.StatusBadRequest, response.ErrorSimple("INVALID_PASSWORD", err.Error()))
	case service.ErrProjectNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrProjectKeyTaken:
		c.JSON(http.StatusConflict, response.ErrorSimple("PROJECT_KEY_TAKEN", err.Error()))
	case service.ErrAlreadyMember:
		c.JSON(http.StatusConflict, response.ErrorSimple("ALREADY_MEMBER", err.Error()))
	case service.ErrMemberNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrLabelNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrLabelNameTaken:
		c.JSON(http.StatusConflict, response.ErrorSimple("LABEL_NAME_TAKEN", err.Error()))
	case service.ErrNotProjectMember:
		c.JSON(http.StatusForbidden, response.ErrorSimple("FORBIDDEN", err.Error()))
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, response.ErrorSimple("FORBIDDEN", err.Error()))
	case service.ErrIssueNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrCommentNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrAttachmentNotFound:
		c.JSON(http.StatusNotFound, response.ErrorSimple("NOT_FOUND", err.Error()))
	case service.ErrFileTooLarge:
		c.JSON(http.StatusRequestEntityTooLarge, response.ErrorSimple("FILE_TOO_LARGE", err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, response.ErrorSimple("INTERNAL_ERROR", "an unexpected error occurred"))
	}
}
