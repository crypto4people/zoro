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
	"zoro/pool"
	"zoro/server"

	"github.com/ethereum/go-ethereum/common"
	"github.com/itsabgr/ge"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var config struct {
	Listen           string
	RatefeedContract common.Address
	Accounts         []common.Address
	Database         string
	Nodes            []string
	DataDir          string
	Domain           string
	Source           *string
}

var flagC = flag.String("c", "config.yaml", "")

func main() {
	// parse config
	flag.Parse()

	ge.Throw(yaml.Unmarshal(ge.Must(os.ReadFile(*flagC)), &config))

	if len(config.Nodes) == 0 {
		slog.Error("nodes are either empty or nil")
		return
	}

	if config.DataDir == "" {
		config.DataDir = "./"
	}

	if config.Source == nil || *config.Source == "" {
		source := "0.0.0.0"
		config.Source = &source
	}

	if config.RatefeedContract == (common.Address{}) {
		slog.Warn("ratefeed contract is empty")
	}

	// DB
	db := ge.Must(sql.Open("sqlite3", "file:"+filepath.Join(config.DataDir, config.Database)))
	defer db.Close()

	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	// ethereum pool
	ethClientPool := ge.Must(pool.New(ctx, config.Nodes))
	defer ethClientPool.Close()

	// server setup
	cert, key, err := obtainCertificateForDomain(config.Domain, config.DataDir)
	if err != nil {
		slog.Warn("failed to obtain certificate, fallback to http", slog.String("error", err.Error()))
	}

	app := ge.Must(server.New(ctx, server.Config{
		Source:           *config.Source,
		RatefeedContract: config.RatefeedContract,
		Accounts:         config.Accounts,
		Database:         db,
		EthClientPool:    ethClientPool,
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
