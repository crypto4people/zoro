package proxy

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"zoro/eth"
	"zoro/rate"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gobwas/ws"
)

const version = "beta"

type Server struct {
	uptime time.Time

	logger *slog.Logger

	db struct {
		conn   *sql.DB
		height atomic.Int64
		subs   atomic.Int64
		vacuum struct {
			interval time.Duration
			mutex    sync.Mutex
			last     time.Time
		}
	}

	proxy struct {
		reporting struct {
			internal time.Duration
			timeout  time.Duration
		}
		dialer     net.Dialer
		upgrader   ws.HTTPUpgrader
		congestion congestionMap
	}

	httpMux http.ServeMux

	ethClient        eth.Client
	ethStartingBlock int64
}

type Info struct {
	Version  string
	Uptime   int64
	Conns    int64
	Subs     int64
	Chain    string
	Height   int64
	RateFeed common.Address
	Account  common.Address
	Balance  string
	GasPrice string
	Rate     rate.Rate
}

func (server *Server) Info(ctx context.Context, user common.Address) (*Info, error) {

	price, err := server.ethClient.GasPrice(ctx)
	if err != nil {
		return nil, err
	}

	rate, err := server.ethClient.RateAt(ctx, rpc.LatestBlockNumber.Int64())
	if err != nil {
		return nil, err
	}

	balance, err := server.ethClient.BalanceOf(ctx, user)
	if err != nil {
		return nil, err
	}

	var account common.Address
	for account = range server.ethClient.Accounts() {
		break
	}

	return &Info{
		Version:  version,
		Conns:    int64(server.proxy.congestion.Count()),
		Subs:     int64(server.db.subs.Load()),
		Chain:    server.ethClient.Chain(),
		Uptime:   time.Since(server.uptime).Milliseconds() / 1000,
		Account:  account,
		Height:   server.db.height.Load(),
		RateFeed: server.ethClient.RateFeed(),
		Rate:     *rate,
		Balance:  balance.Text(10),
		GasPrice: price.Text(10),
	}, nil
}

func (server *Server) init(ctx context.Context) error {
	if err := server.migrate(ctx); err != nil {
		return err
	}

	if err := server.vacuum(ctx); err != nil {
		return err
	}

	if err := server.countSubs(ctx); err != nil {
		return err
	}

	if _, err := server.getLocalHeight(ctx); err != nil {
		return err
	}

	server.initMux()

	return nil
}
