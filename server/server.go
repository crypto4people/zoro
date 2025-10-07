package server

import (
	"context"
	"database/sql"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gobwas/ws"
)

type Server struct {
	uptime time.Time

	logger *slog.Logger

	rate struct {
		feed func(ctx context.Context, block int64) (Rate, error)
	}

	db struct {
		conn   *sql.DB
		height atomic.Int64
		subs   atomic.Int64
	}

	proxy struct {
		reporting struct {
			internal time.Duration
			timeout  time.Duration
		}
		dialer   net.Dialer
		upgrader ws.HTTPUpgrader
		conns    atomic.Int64
		max      int
	}

	http struct {
		mux http.ServeMux
	}

	eth struct {
		offset uint8
		start  int64
		addrs  []common.Address
		signer types.Signer
		chain  *big.Int
		client *ethclient.Client
		sleep  time.Duration
	}
}

type Info struct {
	Uptime  int
	Conns   int
	Subs    int
	Chain   *big.Int
	Height  int64
	Account common.Address
}

func (server *Server) Info() Info {
	return Info{
		Conns:   int(server.proxy.conns.Load()),
		Subs:    int(server.db.subs.Load()),
		Chain:   new(big.Int).Set(server.eth.chain),
		Uptime:  int(time.Since(server.uptime).Milliseconds()) / 1000,
		Account: server.eth.addrs[0],
		Height:  server.db.height.Load(),
	}
}

func (server *Server) init(ctx context.Context) error {
	if err := server.migrate(ctx); err != nil {
		return err
	}

	if err := server.vaccum(ctx); err != nil {
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
