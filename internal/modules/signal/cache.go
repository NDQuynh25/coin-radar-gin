package signal

import (
	"context"
)

type SignalCache interface {
	PublishSignal(ctx context.Context, channel string, signalPayload string) error
	SubscribeSignal(ctx context.Context, channel string) (<-chan string, error)
}
