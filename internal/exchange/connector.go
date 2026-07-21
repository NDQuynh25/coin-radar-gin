package exchange

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/modules/market"
)

// Connector streams normalized public trades until ctx is cancelled.
type Connector interface {
	Name() string
	Run(ctx context.Context, trades chan<- *market.Trade) error
}

type streamFunc func(context.Context) error

func reconnect(ctx context.Context, name string, stream streamFunc) error {
	backoff := time.Second
	for ctx.Err() == nil {
		if err := stream(ctx); err != nil && ctx.Err() == nil {
			log.Printf("[%s] stream disconnected: %v", name, err)
		}
		delay := backoff + time.Duration(rand.IntN(500))*time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
	return ctx.Err()
}

func New(cfg config.ExchangeConfig) (Connector, error) {
	switch cfg.Name {
	case "binance":
		return newBinance(cfg), nil
	case "bybit":
		return newBybit(cfg), nil
	case "okx":
		return newOKX(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported exchange %q", cfg.Name)
	}
}

func publish(ctx context.Context, out chan<- *market.Trade, trade *market.Trade) error {
	select {
	case out <- trade:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
