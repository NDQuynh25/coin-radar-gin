// Package models contains GORM mappings for relational application tables.
// These structs never create or alter schema; migrations/ owns the database schema.
package models

import "time"

type User struct {
	ID           string     `gorm:"column:id;type:uuid;primaryKey"`
	TelegramID   *int64     `gorm:"column:telegram_id"`
	Email        *string    `gorm:"column:email"`
	Username     *string    `gorm:"column:username"`
	PasswordHash *string    `gorm:"column:password_hash"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
	CreatedBy    *string    `gorm:"column:created_by;type:uuid"`
	UpdatedBy    *string    `gorm:"column:updated_by;type:uuid"`
}

func (User) TableName() string { return "users" }

type Subscription struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string     `gorm:"column:user_id;type:uuid"`
	Plan      string     `gorm:"column:plan"`
	Status    string     `gorm:"column:status"`
	StartedAt time.Time  `gorm:"column:started_at"`
	ExpiresAt *time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedBy *string    `gorm:"column:created_by;type:uuid"`
	UpdatedBy *string    `gorm:"column:updated_by;type:uuid"`
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
	CreatedAt   time.Time  `gorm:"column:created_at"`
	ConfirmedAt *time.Time `gorm:"column:confirmed_at"`
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
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedBy *string    `gorm:"column:created_by;type:uuid"`
	UpdatedBy *string    `gorm:"column:updated_by;type:uuid"`
}

func (AlertRule) TableName() string { return "alert_rules" }

type Watchlist struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	UserID    string    `gorm:"column:user_id;type:uuid"`
	Symbol    string    `gorm:"column:symbol"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (Watchlist) TableName() string { return "watchlists" }
