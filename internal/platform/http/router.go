package http

import (
	"coin-radar-gin/config"
	"coin-radar-gin/internal/shared/validator"

	"github.com/gin-gonic/gin"
)

// NewRouter sets up the Gin engine and mounts each module's routes.
// Dependencies are created by the composition root in cmd/api.
func NewRouter(cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Report validation errors using JSON field names.
	validator.Init()

	// Root-level routes
	NewHealthHandler(cfg).Register(r)

	return r
}
