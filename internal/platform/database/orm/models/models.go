// Package models contains GORM mappings for relational application tables.
// These structs never create or alter schema; migrations/ owns the database schema.
package models

import "time"

type AuditFields struct {
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedBy *string    `gorm:"column:created_by;type:uuid"`
	UpdatedBy *string    `gorm:"column:updated_by;type:uuid"`
}

type User struct {
	ID           string  `gorm:"column:id;type:uuid;primaryKey"`
	TelegramID   *int64  `gorm:"column:telegram_id"`
	Email        *string `gorm:"column:email"`
	Username     *string `gorm:"column:username"`
	PasswordHash *string `gorm:"column:password_hash"`
	AuditFields
}

func (User) TableName() string { return "users" }

type Subscription struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string     `gorm:"column:user_id;type:uuid"`
	Plan      string     `gorm:"column:plan"`
	Status    string     `gorm:"column:status"`
	StartedAt time.Time  `gorm:"column:started_at"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	AuditFields
}

func (Subscription) TableName() string { return "subscriptions" }

type Payment struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID      string     `gorm:"column:user_id;type:uuid"`
	AmountUSD   float64    `gorm:"column:amount_usd"`
	Currency    string     `gorm:"column:currency"`
	Network     string     `gorm:"column:network"`
	TxHash      *string    `gorm:"column:tx_hash"`
	Status      string     `gorm:"column:status"`
	ConfirmedAt *time.Time `gorm:"column:confirmed_at"`
	AuditFields
}

func (Payment) TableName() string { return "payments" }

type AlertRule struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string     `gorm:"column:user_id;type:uuid"`
	Type      string     `gorm:"column:type"`
	Symbol    string     `gorm:"column:symbol"`
	Condition string     `gorm:"column:condition"`
	Threshold float64    `gorm:"column:threshold"`
	Channel   string     `gorm:"column:channel"`
	IsActive  bool       `gorm:"column:is_active"`
	CooldownS int        `gorm:"column:cooldown_s"`
	LastFired *time.Time `gorm:"column:last_fired"`
	AuditFields
}

func (AlertRule) TableName() string { return "alert_rules" }

type Watchlist struct {
	ID     string `gorm:"column:id;type:uuid;primaryKey"`
	UserID string `gorm:"column:user_id;type:uuid"`
	Symbol string `gorm:"column:symbol"`
	AuditFields
}

func (Watchlist) TableName() string { return "watchlists" }

type Symbol struct {
	Symbol     string     `gorm:"column:symbol;primaryKey"`
	BaseAsset  string     `gorm:"column:base_asset"`
	QuoteAsset string     `gorm:"column:quote_asset"`
	Exchange   string     `gorm:"column:exchange"`
	MarketType string     `gorm:"column:market_type"`
	IsActive   bool       `gorm:"column:is_active"`
	ListedAt   *time.Time `gorm:"column:listed_at"`
	AuditFields
}

func (Symbol) TableName() string { return "symbols" }

type KnownWallet struct {
	Address    string  `gorm:"column:address;primaryKey"`
	Chain      string  `gorm:"column:chain"`
	Label      *string `gorm:"column:label"`
	Category   *string `gorm:"column:category"`
	IsExchange bool    `gorm:"column:is_exchange"`
	AuditFields
}

func (KnownWallet) TableName() string { return "known_wallets" }
