package user

import (
	"errors"
	"net/http"

	domainuser "coin-radar/internal/domain/user"
	"coin-radar/internal/interfaces/http/dto"
	"coin-radar/internal/interfaces/http/middleware"
	authsvc "coin-radar/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// Handler exposes user routes. It depends on the auth service both to load
// users and to verify access tokens for its protected routes.
type Handler struct {
	svc *authsvc.Service
}

// New creates a new user Handler.
func New(svc *authsvc.Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts the user routes. Everything under /users requires a valid
// access token.
func (h *Handler) Register(rg *gin.RouterGroup) {
	g := rg.Group("/users")
	g.Use(middleware.Auth(h.svc))
	g.GET("/me", h.me)
}

func (h *Handler) me(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Error("unauthorized", "not authenticated"))
		return
	}

	u, err := h.svc.GetUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domainuser.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.Error("not_found", "user not found"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Error("internal_error", "something went wrong"))
		return
	}
	c.JSON(http.StatusOK, dto.Success(dto.UserResponse{User: u}))
}
