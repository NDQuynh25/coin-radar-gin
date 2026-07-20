package signal

import (
	"net/http"

	"coin-radar-gin/internal/config"
	signaldomain "coin-radar-gin/internal/domain/signal"
	"coin-radar-gin/internal/interfaces/http/dto"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for signal-related routes.
type Handler struct {
	cfg *config.Config
}

// New creates a new signal Handler.
func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

// Register mounts the signal routes onto the given router group.
func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/signals", h.getSignals)
}

func (h *Handler) getSignals(c *gin.Context) {
	c.JSON(http.StatusOK, dto.Success(dto.SignalsResponse{
		Signals: []signaldomain.Signal{},
	}))
}
