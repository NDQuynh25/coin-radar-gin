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
	log.Println("Starting WebSocket Ingestor...")

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config.yaml, using defaults: %v\n", err)
		cfg = &config.Config{}
	}
	_ = cfg

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Mock websocket listen loop
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				log.Println("[Ingestor] Mock received stream trade: BTCUSDT @ 65000.00")
			}
		}
	}()

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Ingestor gracefully...")
	cancel()

	// Wait for cleanup
	time.Sleep(1 * time.Second)
	log.Println("Ingestor stopped.")
}
