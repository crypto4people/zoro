package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"slices"
	"time"
	"zoro/pool"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Config struct {
	Source           string
	RatefeedContract common.Address
	Accounts         []common.Address
	Database         *sql.DB
	EthClientPool    pool.EthClient
	Logger           *slog.Logger
}

func New(ctx context.Context, conf Config) (app *Server, err error) {
	app = &Server{
		logger: conf.Logger,
		uptime: time.Now(),
	}

	app.rate.feed = app.NewRate

	app.eth.client = conf.EthClientPool
	app.eth.sleep = time.Second
	app.eth.ratefeedContract = conf.RatefeedContract
	app.eth.addrs = slices.Clone(conf.Accounts)
	chain, err := app.eth.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}
	app.eth.signer = types.LatestSignerForChainID(chain)

	app.proxy.reporting.internal = time.Second * 10
	app.proxy.reporting.timeout = time.Second * 5
	app.proxy.dialer.LocalAddr = &net.TCPAddr{
		IP:   net.ParseIP(conf.Source),
		Port: 0,
	}
	app.proxy.dialer.Timeout = time.Second * 3
	app.proxy.dialer.KeepAlive = time.Second * 15
	app.proxy.upgrader.Timeout = time.Second * 3
	app.proxy.congestion.Init(2000, 20)

	app.db.conn = conf.Database
	app.db.vaccum.interval = time.Hour

	err = app.init(ctx)

	if err != nil {
		return nil, err
	}

	return app, nil
}
