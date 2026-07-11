package alert

import "time"

type AlertRule struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"` // price, funding, volume...
	Symbol    string    `json:"symbol"`
	Condition string    `json:"condition"` // gt, lt, cross_up, cross_down
	Threshold float64   `json:"threshold"`
	Channel   string    `json:"channel"` // telegram, webhook
	IsActive  bool      `json:"is_active"`
	CooldownS int       `json:"cooldown_s"`
	LastFired time.Time `json:"last_fired,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
