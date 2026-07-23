package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	ossignal "os/signal"
	"syscall"
	"time"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/modules/auth"
	"coin-radar-gin/internal/modules/market"
	"coin-radar-gin/internal/modules/signal"
	"coin-radar-gin/internal/modules/user"
	"coin-radar-gin/internal/platform/database/orm"
	transport "coin-radar-gin/internal/platform/http"
)

func main() {
	log.Println("Starting API Server...")

	// 1. Load config
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Printf("Warning: Failed to load .env, using defaults: %v\n", err)
		cfg = config.Default()
	}

	// 2. Build application dependencies in the composition root.
	db, err := orm.Open(cfg)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database handle: %v", err)
	}
	defer sqlDB.Close()
	userRepo := orm.NewUserRepository(db)
	authService := auth.NewService(userRepo, auth.Config{
		JWTSecret:          cfg.Auth.JWTSecret,
		AccessTTL:          time.Duration(cfg.Auth.AccessTokenTTL) * time.Minute,
		RefreshTTL:         time.Duration(cfg.Auth.RefreshTokenTTL) * time.Hour,
		TelegramBotToken:   cfg.Telegram.Token,
		TelegramMaxAuthAge: 24 * time.Hour,
	})

	// 3. Setup Gin Router
	r := transport.NewRouter(cfg)
	v1 := r.Group("/api/v1")
	market.New(cfg).Register(v1)
	signal.New(cfg).Register(v1)
	auth.NewAuthHandler(authService).Register(v1)
	user.NewHandler(userRepo).Register(v1, authService)

	// 4. HTTP Server config
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 5. Graceful shutdown
	go func() {
		log.Printf("API Server is running on port %d\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	ossignal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("API Server stopped.")
}
