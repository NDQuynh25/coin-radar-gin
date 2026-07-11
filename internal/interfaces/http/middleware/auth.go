package middleware

import (
	"net/http"
	"strings"

	"coin-radar/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// ctxUserIDKey is the gin.Context key under which the authenticated user id is stored.
const ctxUserIDKey = "auth_user_id"

// TokenVerifier is the minimal dependency the auth middleware needs.
// *auth.Service satisfies this.
type TokenVerifier interface {
	VerifyAccessToken(token string) (int64, error)
}

// Auth returns middleware that requires a valid Bearer access token.
// On success it stores the user id in the context for downstream handlers.
func Auth(v TokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := bearerToken(header)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				dto.Error("unauthorized", "missing or malformed Authorization header"))
			return
		}

		userID, err := v.VerifyAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				dto.Error("unauthorized", "invalid or expired token"))
			return
		}

		c.Set(ctxUserIDKey, userID)
		c.Next()
	}
}

// UserID returns the authenticated user id set by the Auth middleware.
func UserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ctxUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}
