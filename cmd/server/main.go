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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/itsabgr/ge"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

var config struct {
	Listen   *string
	Contract common.Address
	Accounts []common.Address
	Database string
	Nodes    []string
	DataDir  string
	Domain   *string
	Source   *net.IP
}

var flagC = flag.String("c", "config.yaml", "")

func getOutboundIP(ctx context.Context) (net.IP, error) {
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

	ge.Throw(yaml.Unmarshal(ge.Must(os.ReadFile(*flagC)), &config))

	if len(config.Nodes) == 0 {
		slog.Error("nodes are either empty or nil")
		return
	}

	if config.DataDir == "" {
		config.DataDir = "./"
	}

	if config.Source == nil {
		source, err := getOutboundIP(ctx)
		if err != nil {
			slog.Warn("failed to obtain outbound ip address, using 0.0.0.0", slog.String("error", err.Error()))
			source = net.ParseIP("0.0.0.0")
		}
		config.Source = &source
	}

	if config.Listen == nil || *config.Listen == "" {
		listen := "0.0.0.0:443"
		if config.Domain == nil || *config.Domain == "" {
			listen = "0.0.0.0:80"
		}
		config.Listen = &listen
	}

	if config.Contract == (common.Address{}) {
		slog.Warn("ratefeed contract is empty")
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
	if config.Domain != nil || *config.Domain != "" {
		cert, key, err = acme.ObtainCertificateForDomain(*config.Domain, config.DataDir)
		if err != nil {
			// change listen addr?
			slog.Warn("failed to obtain certificate, fallback to http", slog.String("error", err.Error()))
		}
	}

	app := ge.Must(server.New(ctx, server.Config{
		LocalAddr: *config.Source,
		Contract:  config.Contract,
		Accounts:  config.Accounts,
		Database:  db,
		Nodes:     ethClientPool,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}))

	httpServer := http.Server{
		Addr:    *config.Listen,
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

	slog.Info("server started", slog.String("address", *config.Listen))
	ge.Throw(eg.Wait())

}
