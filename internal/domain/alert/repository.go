package alert

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, rule *AlertRule) error
	Update(ctx context.Context, rule *AlertRule) error
	Delete(ctx context.Context, id int64) error
	GetActiveRulesBySymbol(ctx context.Context, symbol string) ([]*AlertRule, error)
	GetRulesByUserID(ctx context.Context, userID int64) ([]*AlertRule, error)
	UpdateLastFired(ctx context.Context, id int64, firedTime time.Time) error
}
