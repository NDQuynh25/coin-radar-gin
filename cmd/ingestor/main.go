package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"
	"time"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/exchange"
	"coin-radar-gin/internal/modules/market"
	"coin-radar-gin/internal/platform/database/timescale"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := timescale.NewPgxPool(cfg)
	if err != nil {
		log.Fatalf("connect TimescaleDB: %v", err)
	}
	defer pool.Close()
	repository := timescale.NewTradeRepository(pool)
	if err := repository.EnsureSchema(ctx); err != nil {
		log.Fatalf("prepare storage: %v", err)
	}

	trades := make(chan *market.Trade, cfg.Ingestor.BufferSize)
	active := 0
	for _, exchangeCfg := range cfg.Exchanges {
		if !exchangeCfg.Enabled {
			continue
		}
		connector, err := exchange.New(exchangeCfg)
		if err != nil {
			log.Fatalf("configure exchange: %v", err)
		}
		active++
		go func() {
			log.Printf("[%s] starting trade stream", connector.Name())
			if err := connector.Run(ctx, trades); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("[%s] stopped: %v", connector.Name(), err)
			}
		}()
	}
	if active == 0 {
		log.Fatal("no exchange is enabled")
	}

	if err := persist(ctx, repository, trades, cfg.Ingestor.BatchSize, time.Duration(cfg.Ingestor.FlushIntervalMS)*time.Millisecond); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("persist trades: %v", err)
	}
	log.Println("ingestor stopped")
}

type tradeWriter interface {
	SaveBatch(context.Context, []*market.Trade) error
}

func persist(ctx context.Context, writer tradeWriter, input <-chan *market.Trade, batchSize int, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	batch := make([]*market.Trade, 0, batchSize)
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := writer.SaveBatch(ctx, batch); err != nil {
			return err
		}
		log.Printf("stored %d trades", len(batch))
		batch = batch[:0]
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if len(batch) > 0 {
				return writer.SaveBatch(flushCtx, batch)
			}
			return ctx.Err()
		case trade := <-input:
			batch = append(batch, trade)
			if len(batch) >= batchSize {
				if err := flush(); err != nil {
					return err
				}
			}
		case <-ticker.C:
			if err := flush(); err != nil {
				return err
			}
		}
	}
}
