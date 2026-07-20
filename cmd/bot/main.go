package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coin-radar-gin/internal/config"
)

func main() {
	log.Println("Starting Telegram Bot listener...")

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config.yaml, using defaults: %v\n", err)
		cfg = &config.Config{}
	}
	_ = cfg

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mock telegram listener loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(3 * time.Second):
				log.Println("[Telegram Bot] Waiting for command signals...")
			}
		}
	}()

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Telegram Bot gracefully...")
	cancel()

	// Wait for cleanup
	time.Sleep(1 * time.Second)
	log.Println("Telegram Bot stopped.")
}
