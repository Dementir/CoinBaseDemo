package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/Dementir/CoinBaseDemo/internal/coinbase"
	"github.com/Dementir/CoinBaseDemo/internal/coingraber"
	"github.com/Dementir/CoinBaseDemo/internal/store"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

type config struct {
	MySQL               string   `yaml:"mysql"`
	CryptoCurrencyPairs []string `yaml:"cryptoCurrencyPairs"`
	Host                string   `yaml:"host"`
}

func loadConfigFromYaml(path string) (*config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Can't read config file: "+path)
	}
	defer f.Close()

	var conf config
	err = yaml.NewDecoder(f).Decode(&conf)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse yaml-file")
	}

	return &conf, nil
}

func main() {
	lg := NewLogger("DEBUG")
	defer lg.Sync()

	configPath := flag.String("c", "config.yaml", "set config path")
	flag.Parse()

	ctx := context.Background()

	config, err := loadConfigFromYaml(*configPath)
	if err != nil {
		lg.Fatalw("parse config error", "error", err)
	}

	db, err := sqlx.ConnectContext(ctx, "mysql", config.MySQL)
	if err != nil {
		lg.Fatalw("cannot initialize connect to db", "error", err)
	}

	dial := websocket.Dialer{}

	wssURL := fmt.Sprintf("wss://%s", config.Host)
	c, _, err := dial.Dial(wssURL, nil)
	if err != nil {
		lg.Fatalw("connect to websocket error", "error", err)
	}
	defer c.Close()

	channelMap := make(map[string]chan *coingraber.Ticker)

	for _, cryptoCurrency := range config.CryptoCurrencyPairs {
		ch := make(chan *coingraber.Ticker)

		channelMap[cryptoCurrency] = ch
	}

	cb := coinbase.New(lg, []string{
		"ETH-BTC",
		"BTC-USD",
		"BTC-EUR",
	}, c)

	go cb.Run(ctx, channelMap)

	s := store.New(db)

	graber := coingraber.NewGraber(lg, s)

	wg := sync.WaitGroup{}
	for _, ch := range channelMap {
		wg.Add(1)
		go graber.Process(ctx, ch, &wg)
	}

	wg.Wait()
}

func NewLogger(lvlStr string, opts ...zap.Option) *zap.SugaredLogger {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(lvlStr)); err != nil {
		log.Fatal(err)
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			TimeKey:     "timestamp",
			EncodeTime:  zapcore.RFC3339TimeEncoder,
			EncodeLevel: zapcore.LowercaseLevelEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
	}

	l, err := config.Build(opts...)
	if err != nil {
		log.Fatal(err)
	}

	return l.Sugar()
}
