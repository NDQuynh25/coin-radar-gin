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

func (r *TradeRepository) EnsureSchema(ctx context.Context) error {
	const schema = `
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE TABLE IF NOT EXISTS trades (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
    time TIMESTAMPTZ NOT NULL,
    exchange TEXT NOT NULL,
    trade_id TEXT NOT NULL,
    symbol TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    qty DOUBLE PRECISION NOT NULL,
    side TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    PRIMARY KEY (time, id),
    UNIQUE (time, exchange, trade_id)
);
SELECT create_hypertable('trades', by_range('time'), if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS idx_trades_symbol_time ON trades (symbol, time DESC);
CREATE INDEX IF NOT EXISTS idx_trades_exchange_time ON trades (exchange, time DESC);`
	if _, err := r.pool.Exec(ctx, schema); err != nil {
		return fmt.Errorf("initialize trades schema: %w", err)
	}
	return nil
}

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
