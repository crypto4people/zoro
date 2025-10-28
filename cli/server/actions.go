package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/crypto4people/zoro/acme"
	"github.com/crypto4people/zoro/eth"
	"github.com/crypto4people/zoro/proxy"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

func runAction(ctx context.Context, cmd *cli.Command) error {
	logger := slog.Default()

	config := cmd.String("config")
	if config == "" {
		return errors.New("config from cmd is not supported yet")
	}

	cfg, err := parseFromFile(config)
	if err != nil {
		return err
	}

	if cfg.Source == "" {
		ip, err := getOutboundIP(ctx)
		if err != nil {
			return err
		}

		cfg.source = ip
	} else {
		cfg.source = net.ParseIP(cfg.Source)
	}

	err = os.MkdirAll(cfg.DataDir, os.ModePerm)
	if err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", filepath.Join(cfg.DataDir, "zoro.db"))
	if err != nil {
		log.Fatal("db open", err)
	}

	var cert, key string
	if cfg.Domain != "" {
		cert, key, err = acme.ObtainCertificateForDomain(ctx, cfg.Domain, cfg.DataDir)
	}

	if err != nil {
		logger.Warn("failed to get acme certificate, using http", slog.String("error", err.Error()))
	}

	originalEthClient, err := ethclient.DialContext(ctx, cfg.EthNode)
	if err != nil {
		return err
	}
	defer originalEthClient.Close()

	ethClient, err := eth.NewDirectClient(
		ctx,
		originalEthClient,
		eth.DirectClientConfig{
			Accounts:  cfg.Accounts,
			RateFeed:  cfg.Contract,
			BlockTime: 10 * time.Second,
		})
	if err != nil {
		return err
	}

	server, err := proxy.NewServer(ctx, proxy.ServerConfig{
		LocalAddr:     cfg.source,
		Contract:      cfg.Contract,
		Accounts:      cfg.Accounts,
		Database:      db,
		Client:        ethClient,
		Logger:        slog.Default(),
		StartingBlock: cfg.StartingBlock,
	})
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:    cfg.Listen,
		Handler: server,
	}

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		if cert != "" && key != "" {
			err = httpServer.ListenAndServeTLS(cert, key)
		} else {
			err = httpServer.ListenAndServe()
		}
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	})

	eg.Go(func() error {
		return server.Watch(egCtx)
	})

	eg.Go(func() error {
		<-egCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Shutdown(shutdownCtx)
		return nil
	})

	if err = eg.Wait(); err != nil {
		return err
	}

	return nil
}

func getOutboundIP(ctx context.Context) (net.IP, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "example.com:80")
	if err != nil {
		return nil, err
	}
	return net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0]), nil
}
