package http

import (
	"time"

	"coin-radar-gin/internal/config"
	"coin-radar-gin/internal/infrastructure/storage/memory"
	"coin-radar-gin/internal/interfaces/http/request"
	authsvc "coin-radar-gin/internal/services/auth"

	"coin-radar-gin/internal/interfaces/http/handlers/auth"
	"coin-radar-gin/internal/interfaces/http/handlers/health"
	"coin-radar-gin/internal/interfaces/http/handlers/market"
	"coin-radar-gin/internal/interfaces/http/handlers/signal"
	"coin-radar-gin/internal/interfaces/http/handlers/user"

	"github.com/gin-gonic/gin"
)

// NewRouter sets up the Gin engine and mounts each module's routes.
func NewRouter(cfg *config.Config) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Report validation errors using JSON field names.
	request.Init()

	// --- Dependency wiring ---
	// TODO: move construction to cmd/api/main.go and swap the in-memory repo
	// for a Postgres-backed implementation once storage is ready.
	userRepo := memory.NewUserRepository()
	authService := authsvc.NewService(userRepo, authsvc.Config{
		JWTSecret:          cfg.Auth.JWTSecret,
		AccessTTL:          time.Duration(cfg.Auth.AccessTokenTTL) * time.Minute,
		RefreshTTL:         time.Duration(cfg.Auth.RefreshTokenTTL) * time.Hour,
		TelegramBotToken:   cfg.Telegram.Token,
		TelegramMaxAuthAge: 24 * time.Hour,
	})

	// Root-level routes
	health.New(cfg).Register(r)

	// API v1 routes group
	v1 := r.Group("/api/v1")
	market.New(cfg).Register(v1)
	signal.New(cfg).Register(v1)
	auth.New(authService).Register(v1)
	user.New(authService).Register(v1)

	return r
}
