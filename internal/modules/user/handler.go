package user

import (
	"errors"
	"net/http"

	"coin-radar-gin/internal/middleware"
	"coin-radar-gin/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	users Repository
}

func NewHandler(users Repository) *Handler {
	return &Handler{users: users}
}

// Register mounts the user routes. Everything under /users requires a valid
// access token.
func (h *Handler) Register(rg *gin.RouterGroup, verifier middleware.TokenVerifier) {
	g := rg.Group("/users")
	g.Use(middleware.Auth(verifier))
	g.GET("/me", h.me)
}

func (h *Handler) me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Error("unauthorized", "not authenticated"))
		return
	}

	u, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, response.Error("not_found", "user not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error("internal_error", "something went wrong"))
		return
	}
	c.JSON(http.StatusOK, response.Success(UserResponse{User: u}))
}
