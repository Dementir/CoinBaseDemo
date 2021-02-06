package coinbase

import (
	"context"
	"encoding/json"
	"github.com/Dementir/CoinBaseDemo/internal/coingraber"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const tickerType = "ticker"

type subscribe struct {
	Type       string          `json:"type"`
	ProductIDs []string        `json:"product_ids"`
	Channels   json.RawMessage `json:"channels"`
}

type channel struct {
	Name        string   `json:"name"`
	ProductsIDs []string `json:"products_ids"`
}

type Coinbase struct {
	lg                  *zap.SugaredLogger
	cryptoCurrencyPairs []string
	ws                  *websocket.Conn
}

func New(lg *zap.SugaredLogger, cryptoCurrencyPairs []string, ws *websocket.Conn) *Coinbase {
	return &Coinbase{
		lg:                  lg,
		cryptoCurrencyPairs: cryptoCurrencyPairs,
		ws:                  ws,
	}
}

func (cb *Coinbase) Run(ctx context.Context, coinMaps map[string]chan *coingraber.Ticker) {
	err := cb.subscribe()
	if err != nil {
		cb.lg.Fatalw("cannot subscribe to coinbase", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			break
		default:
		}

		_, message, err := cb.ws.ReadMessage()
		if err != nil {
			cb.lg.Errorw("cannot read message from ws", "error", err)
		}

		ticker := new(coingraber.Ticker)
		err = json.Unmarshal(message, ticker)
		if err != nil {
			cb.lg.Errorw("cannot unmarshel message from ws", "error", err)
		}

		if ticker.Type == tickerType {
			coinMaps[ticker.ProductID] <- ticker
		}
	}
}

func (cb *Coinbase) subscribe() error {
	subscribeMessage, err := cb.getSubscribeMessage()
	if err != nil {
		return err
	}

	err = cb.ws.WriteMessage(websocket.TextMessage, subscribeMessage)
	if err != nil {
		return err
	}

	return nil
}

func (cb *Coinbase) getSubscribeMessage() ([]byte, error) {
	level := "level2"
	heartbeat := "heartbeat"
	productID := channel{
		Name:        "ticker",
		ProductsIDs: cb.cryptoCurrencyPairs,
	}

	channels := []interface{}{
		level,
		heartbeat,
		productID,
	}

	channelsJSON, err := json.Marshal(channels)
	if err != nil {
		return nil, errors.Wrap(err, "channel marshal error")
	}

	subscribeMessage := subscribe{
		Type:       "subscribe",
		ProductIDs: cb.cryptoCurrencyPairs,
		Channels:   channelsJSON,
	}

	msgJSON, err := json.Marshal(subscribeMessage)
	if err != nil {
		return nil, errors.Wrap(err, "subscribe message marshal error")
	}

	return msgJSON, nil
}
