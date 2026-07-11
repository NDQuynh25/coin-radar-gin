package market

import (
	"context"
	"time"
)

type PriceCache interface {
	SetLatestPrice(ctx context.Context, symbol string, price float64, expiration time.Duration) error
	GetLatestPrice(ctx context.Context, symbol string) (float64, error)
}
