package models

import "time"

// The models below map TimescaleDB hypertables for read queries through GORM.
// Hypertable creation, continuous aggregates, retention and ingestion remain
// SQL-migration/pgx responsibilities.

type Trade struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	Time      time.Time  `gorm:"column:time"`
	Exchange  string     `gorm:"column:exchange"`
	TradeID   string     `gorm:"column:trade_id"`
	Symbol    string     `gorm:"column:symbol"`
	Price     float64    `gorm:"column:price"`
	Qty       float64    `gorm:"column:qty"`
	Side      string     `gorm:"column:side"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (Trade) TableName() string { return "trades" }

type Kline struct {
	Time       time.Time `gorm:"column:time;primaryKey"`
	Exchange   string    `gorm:"column:exchange;primaryKey"`
	Symbol     string    `gorm:"column:symbol;primaryKey"`
	Interval   string    `gorm:"column:interval;primaryKey"`
	Open       float64   `gorm:"column:open"`
	High       float64   `gorm:"column:high"`
	Low        float64   `gorm:"column:low"`
	Close      float64   `gorm:"column:close"`
	Volume     float64   `gorm:"column:volume"`
	TradeCount *int      `gorm:"column:trade_count"`
}

func (Kline) TableName() string { return "klines" }

type Derivative struct {
	Time            time.Time  `gorm:"column:time"`
	Exchange        string     `gorm:"column:exchange"`
	Symbol          string     `gorm:"column:symbol"`
	FundingRate     *float64   `gorm:"column:funding_rate"`
	NextFundingTime *time.Time `gorm:"column:next_funding_time"`
	OpenInterest    *float64   `gorm:"column:open_interest"`
	MarkPrice       *float64   `gorm:"column:mark_price"`
	IndexPrice      *float64   `gorm:"column:index_price"`
}

func (Derivative) TableName() string { return "derivatives" }

type Liquidation struct {
	Time     time.Time `gorm:"column:time"`
	Exchange string    `gorm:"column:exchange"`
	Symbol   string    `gorm:"column:symbol"`
	Side     string    `gorm:"column:side"`
	Price    float64   `gorm:"column:price"`
	Qty      float64   `gorm:"column:qty"`
	ValueUSD *float64  `gorm:"column:value_usd"`
}

func (Liquidation) TableName() string { return "liquidations" }

type OrderbookSnapshot struct {
	Time     time.Time `gorm:"column:time"`
	Exchange string    `gorm:"column:exchange"`
	Symbol   string    `gorm:"column:symbol"`
	Bids     []byte    `gorm:"column:bids;type:jsonb"`
	Asks     []byte    `gorm:"column:asks;type:jsonb"`
}

func (OrderbookSnapshot) TableName() string { return "orderbook_snapshots" }

type WhaleTransfer struct {
	Time         time.Time `gorm:"column:time"`
	Chain        string    `gorm:"column:chain"`
	TxHash       string    `gorm:"column:tx_hash"`
	TokenAddress *string   `gorm:"column:token_address"`
	TokenSymbol  *string   `gorm:"column:token_symbol"`
	FromAddress  string    `gorm:"column:from_address"`
	ToAddress    string    `gorm:"column:to_address"`
	Amount       *string   `gorm:"column:amount;type:numeric"`
	ValueUSD     *float64  `gorm:"column:value_usd"`
	Direction    *string   `gorm:"column:direction"`
}

func (WhaleTransfer) TableName() string { return "whale_transfers" }

type Signal struct {
	Time      time.Time `gorm:"column:time;primaryKey"`
	ID        int64     `gorm:"column:id;primaryKey"`
	Type      string    `gorm:"column:type"`
	Exchange  *string   `gorm:"column:exchange"`
	Symbol    string    `gorm:"column:symbol"`
	Severity  string    `gorm:"column:severity"`
	Value     *float64  `gorm:"column:value"`
	Threshold *float64  `gorm:"column:threshold"`
	Payload   []byte    `gorm:"column:payload;type:jsonb"`
}

func (Signal) TableName() string { return "signals" }
