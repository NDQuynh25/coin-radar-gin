package alert

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, rule *AlertRule) error
	Update(ctx context.Context, rule *AlertRule) error
	Delete(ctx context.Context, id string) error
	GetActiveRulesBySymbol(ctx context.Context, symbol string) ([]*AlertRule, error)
	GetRulesByUserID(ctx context.Context, userID string) ([]*AlertRule, error)
	UpdateLastFired(ctx context.Context, id string, firedTime time.Time) error
}
