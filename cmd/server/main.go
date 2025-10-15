package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
	"zoro/acme"
	"zoro/server"

	"github.com/caarlos0/env/v11"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/itsabgr/ge"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

var config struct {
	Listen   string   `env:"LISTEN"`
	Contract string   `env:"CONTRACT"`
	Accounts []string `env:"ACCOUNTS,required" envSeperator:","`
	Database string   `env:"DATABASE,required"`
	Nodes    []string `env:"NODES" envSeperator:","`
	DataDir  string   `env:"DATA_DIR" envDefault:"./data"`
	Domain   *string  `env:"DOMAIN"`
	Source   *net.IP  `env:"SOURCE"`
}

var flagC = flag.String("c", "config.yaml", "")

func getOutboundIP() (net.IP, error) {
	conn, err := net.DialTCP("tcp", nil, ge.Must(net.ResolveTCPAddr("tcp", "example.com:80")))
	if err != nil {
		return nil, err
	}

	return net.IP(strings.Split(conn.LocalAddr().String(), ":")[0]), nil
}

func main() {
	// parse config
	ctx, _ := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)

	flag.Parse()
	ge.Throw(env.Parse(&config))

	if config.Source == nil {
		source, err := getOutboundIP()
		if err != nil {
			slog.Warn("failed to obtain outbound ip address, using 0.0.0.0", slog.String("error", err.Error()))
			source = net.ParseIP("0.0.0.0")
		}
		config.Source = &source
	}

	// DB
	db := ge.Must(sql.Open("sqlite3", "file:"+filepath.Join(config.DataDir, config.Database)))
	defer db.Close()

	// ethereum pool
	ethClients := []*ethclient.Client{}
	for _, node := range config.Nodes {
		ethClients = append(ethClients, ge.Must(ethclient.Dial(node)))
	}
	ethClientPool := ge.Must(server.NewPool(ctx, ethClients))
	defer ethClientPool.Close()

	// server setup
	var cert, key string
	var err error

	domain := ""
	if config.Domain != nil {
		domain = *config.Domain
	}

	if domain != "" {
		cert, key, err = acme.ObtainCertificateForDomain(ctx, domain, config.DataDir)
		if err != nil {
			slog.Warn("failed to obtain certificate, fallback to HTTP", slog.String("error", err.Error()))
		}
	}

	if config.Listen == "" {
		if cert != "" && key != "" {
			config.Listen = ":443"
		} else {
			config.Listen = ":80"
		}
	}

	// app
	app := ge.Must(server.New(ctx, server.Config{
		LocalAddr: *config.Source,
		Contract:  common.HexToAddress(config.Contract),
		Accounts:  lo.Map(config.Accounts, func(addr string, _ int) common.Address { return common.HexToAddress(addr) }),
		Database:  db,
		Nodes:     ethClientPool,
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
				slog.Warn("watcher failed", slog.String("err", err.Error()))
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
