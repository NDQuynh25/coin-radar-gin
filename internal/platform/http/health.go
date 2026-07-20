package http

import (
	"net/http"
	"time"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/shared/response"
	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for health-check routes.
type HealthHandler struct {
	cfg *config.Config
}

// New creates a new health Handler.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// Register mounts the health routes onto the given router.
func (h *HealthHandler) Register(r gin.IRouter) {
	r.GET("/health", h.check)
}

func (h *HealthHandler) check(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
		"env":    h.cfg.App.Env,
	}))
}
