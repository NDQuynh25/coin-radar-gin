package orm

import (
	"context"

	"coin-radar-gin/internal/modules/market"
	"coin-radar-gin/internal/platform/database/orm/models"

	"gorm.io/gorm"
)

// TradeReader uses GORM for simple dashboard reads. High-volume writes remain
// in timescale.TradeRepository, which uses pgx batch inserts.
type TradeReader struct{ db *gorm.DB }

func NewTradeReader(db *gorm.DB) *TradeReader { return &TradeReader{db: db} }

func (r *TradeReader) GetLatestTrades(ctx context.Context, symbol string, limit int) ([]*market.Trade, error) {
	if limit <= 0 {
		return []*market.Trade{}, nil
	}
	var rows []models.Trade
	if err := r.db.WithContext(ctx).Where("symbol = ? AND deleted_at IS NULL", symbol).
		Order("time DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}
	trades := make([]*market.Trade, 0, len(rows))
	for _, row := range rows {
		trades = append(trades, &market.Trade{TradeID: row.TradeID, Time: row.Time, Exchange: row.Exchange,
			Symbol: row.Symbol, Price: row.Price, Qty: row.Qty, Side: row.Side})
	}
	return trades, nil
}
