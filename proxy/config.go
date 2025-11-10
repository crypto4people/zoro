package proxy

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"time"

	"github.com/crypto4people/zoro/eth"
	"github.com/gobwas/ws"

	"github.com/ethereum/go-ethereum/common"
)

type ServerConfig struct {
	LocalAddr     net.IP
	Contract      common.Address
	Accounts      []common.Address
	Database      *sql.DB
	Client        eth.Client
	Logger        *slog.Logger
	StartingBlock int64
}

func NewServer(ctx context.Context, conf ServerConfig) (app *Server, err error) {
	app = &Server{
		logger: conf.Logger,
		uptime: time.Now(),
	}

	app.ethClient = conf.Client
	app.ethStartingBlock = conf.StartingBlock

	app.proxy.reporting.interval = time.Second * 10
	app.proxy.reporting.timeout = time.Second * 5
	app.proxy.dialer.LocalAddr = &net.TCPAddr{
		IP:   conf.LocalAddr,
		Port: 0,
	}
	app.proxy.dialer.Timeout = time.Second * 3
	app.proxy.dialer.KeepAlive = time.Second * 15
	app.proxy.upgrader = ws.HTTPUpgrader{}
	app.proxy.upgrader.Timeout = time.Second * 3
	app.proxy.congestion.Init(2000, 20)

	app.db.conn = conf.Database
	app.db.vacuum.interval = time.Hour

	err = app.init(ctx)

	if err != nil {
		return nil, err
	}

	return app, nil
}
