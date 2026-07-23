package timescale

import (
	"context"
	"fmt"

	"coin-radar-gin/internal/modules/market"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TradeRepository struct{ pool *pgxpool.Pool }

func NewTradeRepository(pool *pgxpool.Pool) *TradeRepository { return &TradeRepository{pool: pool} }

func (r *TradeRepository) Save(ctx context.Context, trade *market.Trade) error {
	return r.SaveBatch(ctx, []*market.Trade{trade})
}

func (r *TradeRepository) SaveBatch(ctx context.Context, trades []*market.Trade) error {
	if len(trades) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, trade := range trades {
		batch.Queue(`INSERT INTO trades (time, exchange, trade_id, symbol, price, qty, side)
VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (time, exchange, trade_id) DO NOTHING`,
			trade.Time, trade.Exchange, trade.TradeID, trade.Symbol, trade.Price, trade.Qty, trade.Side)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer results.Close()
	for range trades {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("insert trade batch: %w", err)
		}
	}
	return results.Close()
}

func (r *TradeRepository) GetLatestTrades(ctx context.Context, symbol string, limit int) ([]*market.Trade, error) {
	rows, err := r.pool.Query(ctx, `SELECT trade_id,time,exchange,symbol,price,qty,side FROM trades
WHERE symbol=$1 AND deleted_at IS NULL ORDER BY time DESC LIMIT $2`, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*market.Trade, 0, limit)
	for rows.Next() {
		trade := &market.Trade{}
		if err := rows.Scan(&trade.TradeID, &trade.Time, &trade.Exchange, &trade.Symbol, &trade.Price, &trade.Qty, &trade.Side); err != nil {
			return nil, err
		}
		result = append(result, trade)
	}
	return result, rows.Err()
}
