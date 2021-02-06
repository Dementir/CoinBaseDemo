package store

import "github.com/jmoiron/sqlx"

type Store struct {
	db               *sqlx.DB
	tickerRepository TickerRepository
}

type Storage interface {
	TickerRepository() TickerRepository
}

func New(db *sqlx.DB) Storage {
	return &Store{
		db: db,
	}
}

func (s *Store) TickerRepository() TickerRepository {
	if s.tickerRepository != nil {
		return s.tickerRepository
	}

	s.tickerRepository = &ticker{
		store: s,
	}

	return s.tickerRepository
}
