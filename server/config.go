package server

import (
	"context"
	"database/sql"
	"log/slog"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	MaxConns    int
	BlockOffset uint8
	StartBlock  int64
	BlockTime   time.Duration
	Accounts    []common.Address
	Database    *sql.DB
	EthClient   *ethclient.Client
	Logger      *slog.Logger
	RateFeed    func(ctx context.Context, block int64) (Rate, error)
}

func New(ctx context.Context, conf Config) (app *Server, err error) {

	app = &Server{
		logger: conf.Logger,
		uptime: time.Now(),
	}

	app.rate.feed = conf.RateFeed

	app.eth.client = conf.EthClient

	app.eth.chain, err = app.eth.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	app.proxy.dialer.Timeout = time.Second * 5
	app.db.conn = conf.Database
	app.proxy.max = conf.MaxConns
	app.eth.offset = conf.BlockOffset
	app.eth.start = conf.StartBlock
	app.eth.addrs = slices.Clone(conf.Accounts)
	app.eth.signer = types.LatestSignerForChainID(app.eth.chain)
	app.eth.sleep = conf.BlockTime

	err = app.init(ctx)

	if err != nil {
		return nil, err
	}

	return app, nil

}
