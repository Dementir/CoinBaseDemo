package coingraber

import (
	"context"
	"github.com/Dementir/CoinBaseDemo/internal/store"
	"go.uber.org/zap"
	"sync"
	"time"
)

/*
{
    "type": "ticker",
    "trade_id": 20153558,
    "sequence": 3262786978,
    "time": "2017-09-02T17:05:49.250000Z",
    "product_id": "BTC-USD",
    "price": "4388.01000000",
    "side": "buy", // Taker side
    "last_size": "0.03000000",
    "best_bid": "4388",
    "best_ask": "4388.01"
}
*/
type Ticker struct {
	Type      string    `json:"type"`
	TradeID   int64     `json:"trade_id"`
	Sequence  int64     `json:"sequence"`
	Time      time.Time `json:"time"`
	ProductID string    `json:"product_id"`
	Price     float64   `json:"price,string"`
	Side      string    `json:"side"`
	LastSize  float64   `json:"last_size,string"`
	BestBid   float64   `json:"best_bid,string"`
	BestAsk   float64   `json:"best_ask,string"`
}

type Graber struct {
	lg *zap.SugaredLogger
	s  store.Storage
}

func NewGraber(lg *zap.SugaredLogger, s store.Storage) *Graber {
	return &Graber{
		lg: lg,
		s:  s,
	}
}

func (g *Graber) Process(ctx context.Context, ch <-chan *Ticker, wg *sync.WaitGroup) {
	for tick := range ch {
		select {
		case <-ctx.Done():
			break
		default:
		}

		c, cancel := context.WithTimeout(ctx, time.Second*20)

		dbTick := store.Tick{
			Timestamp: tick.Time.UnixNano() / int64(time.Millisecond),
			Symbol:    tick.ProductID,
			Bid:       tick.BestBid,
			Ask:       tick.BestAsk,
		}

		err := g.s.TickerRepository().InsertTick(c, dbTick)
		if err != nil {
			g.lg.Errorw("tick save error", "error", err)
			cancel()
			continue
		}

		cancel()
	}

	wg.Done()
}
