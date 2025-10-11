package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Source    string
	Accounts  []common.Address
	Database  *sql.DB
	EthClient *ethclient.Client
	Logger    *slog.Logger
	RateFeed  func(ctx context.Context, block int64) (Rate, error)
}

func New(ctx context.Context, conf Config) (app *Server, err error) {

	app = &Server{
		logger: conf.Logger,
		uptime: time.Now(),
	}

	app.rate.feed = conf.RateFeed

	app.eth.client = conf.EthClient

	chain, err := app.eth.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	app.proxy.reporting.internal = time.Second * 10
	app.proxy.reporting.timeout = time.Second * 5
	app.proxy.dialer.LocalAddr = &net.TCPAddr{
		IP:   net.ParseIP(conf.Source),
		Port: 0,
	}
	app.proxy.dialer.Timeout = time.Second * 3
	app.proxy.dialer.KeepAlive = time.Second * 15
	app.proxy.upgrader.Timeout = time.Second * 3
	app.db.conn = conf.Database
	app.eth.sleep = time.Second
	app.db.vaccum.interval = time.Hour
	app.proxy.congestion.Init(2000, 20)
	app.eth.addrs = slices.Clone(conf.Accounts)
	app.eth.signer = types.LatestSignerForChainID(chain)

	err = app.init(ctx)

	if err != nil {
		return nil, err
	}

	return app, nil

}
