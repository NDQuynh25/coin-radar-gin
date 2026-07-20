package health

import (
	"net/http"
	"time"

	"coin-radar-gin/internal/config"
	"coin-radar-gin/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for health-check routes.
type Handler struct {
	cfg *config.Config
}

// New creates a new health Handler.
func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

// Register mounts the health routes onto the given router.
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/health", h.check)
}

func (h *Handler) check(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Success(gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
		"env":    h.cfg.App.Env,
	}))
}
