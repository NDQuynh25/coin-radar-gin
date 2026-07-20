package market

import (
	"net/http"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for market-related routes.
type Handler struct {
	cfg *config.Config
}

// New creates a new market Handler.
func New(cfg *config.Config) *Handler {
	return &Handler{cfg: cfg}
}

// Register mounts the market routes onto the given router group.
func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/symbols", h.getSymbols)
}

func (h *Handler) getSymbols(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(SymbolsResponse{
		Symbols: []string{"BTCUSDT", "ETHUSDT"},
	}))
}

type SymbolsResponse struct {
	Symbols []string `json:"symbols"`
}
