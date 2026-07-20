package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coin-radar-gin/internal/config"
	router "coin-radar-gin/internal/interfaces/http"
)

func main() {
	log.Println("Starting API Server...")

	// 1. Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config.yaml, using defaults: %v\n", err)
		cfg = &config.Config{
			Server: config.ServerConfig{Port: 9000},
		}
	}

	// 2. Setup Gin Router
	r := router.NewRouter(cfg)

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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down API Server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("API Server stopped.")
}
