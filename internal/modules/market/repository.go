package market

import "context"

type Repository interface {
	Save(ctx context.Context, trade *Trade) error
	SaveBatch(ctx context.Context, trades []*Trade) error
	GetLatestTrades(ctx context.Context, symbol string, limit int) ([]*Trade, error)
}
