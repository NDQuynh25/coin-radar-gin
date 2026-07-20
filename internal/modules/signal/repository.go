package signal

import "context"

type Repository interface {
	Save(ctx context.Context, signal *Signal) error
	GetLatestSignals(ctx context.Context, limit int) ([]*Signal, error)
	GetSignalsBySymbol(ctx context.Context, symbol string, limit int) ([]*Signal, error)
}
