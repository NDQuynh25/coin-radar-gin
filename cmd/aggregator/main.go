package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"coin-radar-gin/config"
)

func main() {
	log.Println("Starting Background Aggregator...")

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Printf("Warning: Failed to load .env, using defaults: %v\n", err)
		cfg = &config.Config{}
	}
	_ = cfg

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mock aggregator ticker loop (e.g. rollup klines, clean old raw db records)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Println("[Aggregator] Mock continuous aggregate check: Success")
			}
		}
	}()

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Aggregator gracefully...")
	cancel()

	// Wait for cleanup
	time.Sleep(1 * time.Second)
	log.Println("Aggregator stopped.")
}
