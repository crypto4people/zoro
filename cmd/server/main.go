package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	Listen           string
	Rate             server.Rate
	RatefeedContract common.Address
	Accounts         []common.Address
	Database         string
	Node             string
	DataDir          string
	Domain           string
	Source           *string
}

var flagC = flag.String("c", "config.yaml", "")

func main() {

	flag.Parse()

	ge.Throw(yaml.Unmarshal(ge.Must(os.ReadFile(*flagC)), &config))

	ethClient := ge.Must(ethclient.Dial(config.Node))
	defer ethClient.Close()

	db := ge.Must(sql.Open("sqlite3", filepath.Join(config.DataDir, config.Database)))
	defer db.Close()

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	source := "0.0.0.0"
	if config.Source != nil {
		source = *config.Source
	}

	cert, key, err := obtainCertificateForDomain(config.Domain, config.DataDir)
	if err != nil {
		slog.Warn("failed to obtain certificate, fallback to http", slog.String("error", err.Error()))
	}

	app := ge.Must(server.New(ctx, server.Config{
		Source:           source,
		RatefeedContract: config.RatefeedContract,
		Accounts:         config.Accounts,
		Database:         db,
		EthClient:        ethClient,
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
		var err error
		if cert != "" && key != "" {
			err = httpServer.ListenAndServeTLS(cert, key)
		} else {
			err = httpServer.ListenAndServe()
		}

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
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second * 5):
					continue
				}
			}
		}
	})

	slog.Info("server started", slog.String("address", config.Listen))
	ge.Throw(eg.Wait())

}
