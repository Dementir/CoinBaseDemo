package store

import (
	"context"
	"github.com/pkg/errors"
)

type ticker struct {
	store *Store
}

type Tick struct {
	Timestamp int64   `db:"timestamp"`
	Symbol    string  `db:"symbol"`
	Bid       float64 `db:"bid"`
	Ask       float64 `db:"ask"`
}

func (t *ticker) InsertTick(ctx context.Context, tick Tick) error {
	query := `
insert into ticks(
	timestamp
	, symbol
	, bid
	, ask
) value (
	:timestamp
	, :symbol
	, :bid
	, :ask
);`

	stmt, err := t.store.db.PrepareNamed(query)
	if err != nil {
		return errors.Wrap(err, "prepare error")
	}

	_, err = stmt.ExecContext(ctx, tick)
	if err != nil {
		return errors.Wrap(err, "sql request error")
	}

	return nil
}
