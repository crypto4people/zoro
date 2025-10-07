package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
	"zoro/server"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/itsabgr/ge"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var config struct {
	Listen      string
	Rate        server.Rate
	MaxConns    int
	BlockOffset uint8
	StartBlock  int64
	Accounts    []common.Address
	Database    string
	BlockTime   time.Duration
	Node        string
}

var flagC = flag.String("c", "config.yaml", "")

func main() {

	flag.Parse()

	ge.Throw(yaml.Unmarshal(ge.Must(os.ReadFile(*flagC)), &config))

	ethClient := ge.Must(ethclient.Dial(config.Node))
	defer ethClient.Close()

	db := ge.Must(sql.Open("sqlite3", config.Database))
	defer db.Close()

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	app := ge.Must(server.New(ctx, server.Config{
		MaxConns:    config.MaxConns,
		BlockOffset: config.BlockOffset,
		StartBlock:  config.StartBlock,
		Accounts:    config.Accounts,
		BlockTime:   config.BlockTime,
		Database:    db,
		EthClient:   ethClient,
		RateFeed: func(ctx context.Context, block int64) (server.Rate, error) {
			return config.Rate, nil
		},
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}))

	httpServer := http.Server{
		Addr:    config.Listen,
		Handler: app,
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
		return nil
	})

	eg.Go(func() error {
		err := httpServer.ListenAndServe()
		if ctx.Err() != nil {
			return nil
		}
		return err
	})

	eg.Go(func() (err error) {
		for {
			err = app.Watch(ctx)
			if ctx.Err() != nil {
				return nil
			}
			if err != nil {
				slog.Warn("logger failed", slog.String("err", err.Error()))
			}
		}
	})

	ge.Throw(eg.Wait())

}
