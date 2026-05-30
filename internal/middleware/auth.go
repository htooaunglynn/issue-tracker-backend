package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/htooaunglynn/issue-tracker-backend/internal/config"
	pkgjwt "github.com/htooaunglynn/issue-tracker-backend/internal/pkg/jwt"
	"github.com/htooaunglynn/issue-tracker-backend/internal/pkg/response"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyRole   = "userRole"
)

func AuthRequired(cfg config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				response.ErrorSimple("UNAUTHORIZED", "missing or invalid authorization header"))
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := pkgjwt.ValidateToken(tokenStr, cfg.Secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				response.ErrorSimple("UNAUTHORIZED", "invalid or expired token"))
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}
