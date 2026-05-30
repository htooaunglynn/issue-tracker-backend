package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/htooaunglynn/issue-tracker-backend/internal/domain"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
)

func RequireRole(roles ...domain.GlobalRole) gin.HandlerFunc {
	allowed := make(map[domain.GlobalRole]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden,
				response.ErrorSimple("FORBIDDEN", "access denied"))
			return
		}

		role := roleVal.(domain.GlobalRole)
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden,
				response.ErrorSimple("FORBIDDEN", "insufficient permissions"))
			return
		}

		c.Next()
	}
}
