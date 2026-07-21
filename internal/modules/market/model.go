package market

import "time"

type Trade struct {
	TradeID  string    `json:"trade_id"`
	Time     time.Time `json:"time"`
	Exchange string    `json:"exchange"`
	Symbol   string    `json:"symbol"`
	Price    float64   `json:"price"`
	Qty      float64   `json:"qty"`
	Side     string    `json:"side"` // buy, sell
}

type Kline struct {
	Time       time.Time `json:"time"`
	Exchange   string    `json:"exchange"`
	Symbol     string    `json:"symbol"`
	Interval   string    `json:"interval"`
	Open       float64   `json:"open"`
	High       float64   `json:"high"`
	Low        float64   `json:"low"`
	Close      float64   `json:"close"`
	Volume     float64   `json:"volume"`
	TradeCount int       `json:"trade_count"`
}
