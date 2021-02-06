package store

import "context"

type TickerRepository interface {
	InsertTick(ctx context.Context, tick Tick) error
}
