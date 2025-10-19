package proxy

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

func (server *Server) serveProxy(writer http.ResponseWriter, logger *slog.Logger, req *http.Request, account common.Address) error {
	congestionCtx, cancelCongestion := server.proxy.congestion.Push(account)
	if congestionCtx == nil {
		logger.Debug("connection dropped due to congestion")
		return errors.New("congestion")
	}

	ctx := req.Context()

	reporting, err := server.prepareReport(ctx, account)
	if err != nil {
		logger.Debug("connection dropped due to error on prepareReport", slog.String("error", err.Error()))
		cancelCongestion()
		return err
	}

	if err = reporting.Report(ctx,
		server.proxy.reporting.timeout,
		server.proxy.reporting.internal,
	); err != nil {
		logger.Debug("connection dropped due to error on Report", slog.String("error", err.Error()))
		cancelCongestion()
		_ = reporting.Close()
		return err
	}

	network := req.URL.Query().Get("net")
	switch network {
	case "tcp", "udp":
	default:
		logger.Debug("connection dropped due to invalid network requested by user", slog.String("network", network))
		cancelCongestion()
		_ = reporting.Close()
		return errors.New("invalid network")
	}
	addr, err := validateAddr(req.URL.Query().Get("addr"))
	if err != nil {
		logger.Debug("connection dropped due to invalid address requested by user", slog.String("address", addr))
		cancelCongestion()
		_ = reporting.Close()
		return err
	}

	logger.Debug("new conn",
		slog.String("network", network),
		slog.String("address", addr),
	)

	return server.doProxy(ctx, logger, congestionCtx, cancelCongestion, writer, req, reporting, network, addr)

}

func (server *Server) doProxy(ctx context.Context, logger *slog.Logger, congestionCtx context.Context, cancelCongestion context.CancelFunc, writer http.ResponseWriter, req *http.Request, reporting *reporting, network, addr string) (err error) {
	ioConn, err := server.proxy.dialer.DialContext(ctx, network, addr)
	if err != nil {
		logger.Debug("connection dropped due to failure on dialing remote", slog.String("error", err.Error()))
		cancelCongestion()
		_ = reporting.Close()
		return err
	}

	wsConn, _, _, err := server.proxy.upgrader.Upgrade(req, writer)
	if err != nil {
		logger.Debug("connection dropped due to failure on upgrading remote connection", slog.String("error", err.Error()))
		cancelCongestion()
		_ = reporting.Close()
		_ = ioConn.Close()
		return err
	}

	go func() {

		eg, ctx := errgroup.WithContext(congestionCtx)
		eg.Go(func() (err error) {
			defer func() {
				logger.Debug("connection ended")
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
						logger.Debug("connection dropped due to failure on Report", slog.String("error", err.Error()))
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
					server.logger.Debug("connection dropped due to failure when reading client data", slog.String("err", err.Error()))
					return err
				}
				_, err = ioConn.Write(data)
				server.logger.Debug("connection dropped due to failure when writing on remote conn", slog.String("err", err.Error()))
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
					server.logger.Debug("connection dropped due to failure when reading remote data", slog.String("err", err.Error()))
					return err
				}
				err = wsutil.WriteServerMessage(wsConn, ws.OpBinary, data[:n])
				if err != nil {
					server.logger.Debug("connection dropped due to failure when writing to client", slog.String("err", err.Error()))
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
