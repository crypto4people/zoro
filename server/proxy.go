package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"golang.org/x/sync/errgroup"
)

func validateAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	return net.JoinHostPort(host, port), err
}

func (server *Server) serveProxy(writer http.ResponseWriter, req *http.Request, account common.Address) error {

	congestionCtx, cancelCongestion := server.proxy.congestion.Push(account)

	if congestionCtx == nil {
		return errors.New("congestion")
	}

	ctx := req.Context()

	reporting, err := server.prepareReport(ctx, account)
	if err != nil {
		cancelCongestion()
		return err
	}

	if err = reporting.Report(ctx,
		server.proxy.reporting.timeout,
		server.proxy.reporting.internal,
	); err != nil {
		cancelCongestion()
		_ = reporting.Close()
		return err
	}

	network := req.URL.Query().Get("net")
	switch network {
	case "tcp", "udp":
	default:
		cancelCongestion()
		_ = reporting.Close()
		return errors.New("invalid network")
	}
	addr, err := validateAddr(req.URL.Query().Get("addr"))
	if err != nil {
		cancelCongestion()
		_ = reporting.Close()
		return err
	}

	server.logger.Debug("new conn",
		slog.String("net", network),
		slog.String("addr", addr),
		slog.String("user", account.Hex()),
	)

	return server.doProxy(ctx, congestionCtx, cancelCongestion, writer, req, reporting, network, addr)

}

func (server *Server) doProxy(ctx context.Context, congestionCtx context.Context, cancelCongestion context.CancelFunc, writer http.ResponseWriter, req *http.Request, reporting *reporting, network, addr string) (err error) {
	ioConn, err := server.proxy.dialer.DialContext(ctx, network, addr)
	if err != nil {
		cancelCongestion()
		_ = reporting.Close()
		return err
	}

	wsConn, _, _, err := server.proxy.upgrader.Upgrade(req, writer)
	if err != nil {
		cancelCongestion()
		_ = reporting.Close()
		_ = ioConn.Close()
		return err
	}

	go func() {

		eg, ctx := errgroup.WithContext(congestionCtx)
		eg.Go(func() (err error) {
			defer func() {
				ioConn.Close()
				closeWebsocket(wsConn)
				reporting.Close()
				cancelCongestion()
			}()

			ticker := time.NewTicker(server.proxy.reporting.internal)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
					if err = reporting.Report(ctx,
						server.proxy.reporting.timeout,
						server.proxy.reporting.internal,
					); err != nil {
						server.logger.Debug("reporing failed", slog.String("err", err.Error()))
						return err
					}
				}
			}
		})

		eg.Go(func() (err error) {
			var data []byte
			for {
				data, _, err = wsutil.ReadClientData(wsConn)
				if err != nil {
					return err
				}
				_, err = ioConn.Write(data)
				if err != nil {
					return err
				}
			}
		})

		eg.Go(func() (err error) {
			data := make([]byte, 4096)
			var n int
			for {
				n, err = ioConn.Read(data)
				if err != nil {
					return err
				}
				err = wsutil.WriteServerMessage(wsConn, ws.OpBinary, data[:n])
				if err != nil {
					return err
				}
			}
		})

		eg.Wait()

	}()

	return nil

}

func closeWebsocket(conn io.WriteCloser) {
	wsutil.WriteServerMessage(conn, ws.OpClose, nil)
	conn.Close()
}
