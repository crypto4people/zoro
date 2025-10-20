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

	"github.com/ethereum/go-ethereum/common"
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
		vaccum struct {
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

	http struct {
		mux http.ServeMux
	}

	eth struct {
		rate   atomic.Pointer[Rate]
		addrs  []common.Address
		sleep  time.Duration
		client *EthClient
	}
}

type Info struct {
	Version  string
	Uptime   int
	Conns    int
	Subs     int
	Chain    string
	Height   int64
	Contract common.Address
	Account  common.Address
	Rate     Rate
}

func (server *Server) Info() Info {
	return Info{
		Version:  version,
		Conns:    int(server.proxy.congestion.Count()),
		Subs:     int(server.db.subs.Load()),
		Chain:    server.eth.client.Chain(),
		Uptime:   int(time.Since(server.uptime).Milliseconds()) / 1000,
		Account:  server.eth.addrs[0],
		Height:   server.db.height.Load(),
		Contract: server.eth.client.Contract(),
		Rate:     *server.eth.rate.Load(),
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

	//todo fill rate

	server.initMux()

	return nil
}
