package proxy

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"golang.org/x/sync/errgroup"
)

func (server *Server) doProxy(
	writer http.ResponseWriter, request *http.Request,
	logger *slog.Logger,
	addr common.Address,
) error {
	cancelFuncs := []func(){}
	defer func() {
		for _, fn := range cancelFuncs {
			fn()
		}
	}()

	ctx := request.Context()

	network, upstreamAddr, err := getParams(request)
	if err != nil {
		logger.Debug("connection dropped due to invalid params")
		return err
	}

	wsConn, _, _, err := server.proxy.upgrader.Upgrade(request, writer)
	if err != nil {
		return err
	}
	cancelFuncs = append(cancelFuncs, func() { wsConn.Close() })

	err = server.authConn(wsConn, addr, network, upstreamAddr)
	if err != nil {
		return err
	}

	cgCtx, cancel := server.proxy.congestion.Push(addr)
	if cgCtx == nil {
		logger.Debug("connection dropped due to congestion")
		return errors.New("max connections exceeded")
	}
	cancelFuncs = append(cancelFuncs, cancel)

	reporting, err := server.prepareReport(ctx, addr)
	if err != nil {
		return err
	}
	cancelFuncs = append(cancelFuncs, func() { reporting.Close() })

	ioConn, err := server.proxy.dialer.DialContext(ctx, network, upstreamAddr)
	if err != nil {
		logger.Debug("connection dropped due to failure on dialing remote", slog.String("error", err.Error()))
		return err
	}
	cancelFuncs = append(cancelFuncs, func() { ioConn.Close() })

	eg, egCtx := errgroup.WithContext(cgCtx)

	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			_ = ioConn.Close()
			_ = wsConn.Close()
		})
	}

	eg.Go(func() error {
		ticker := time.NewTicker(server.proxy.reporting.internal)
		defer ticker.Stop()
		defer closeAll()

		for {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			case <-ticker.C:
				if err := reporting.Report(
					egCtx,
					server.proxy.reporting.timeout,
					server.proxy.reporting.internal,
				); err != nil {
					logger.Debug("connection dropped due to failure on Report", slog.String("error", err.Error()))
					return err
				}
			}
		}
	})

	eg.Go(func() error {
		defer closeAll()
		for {
			data, _, err := wsutil.ReadClientData(wsConn)
			if err != nil {
				server.logger.Debug("connection dropped: read client data", slog.String("err", err.Error()))
				return err
			}
			if _, err := ioConn.Write(data); err != nil {
				server.logger.Debug("connection dropped: write to remote", slog.String("err", err.Error()))
				return err
			}
		}
	})

	eg.Go(func() error {
		defer closeAll()
		buf := make([]byte, 4096)
		for {
			n, err := ioConn.Read(buf)
			if err != nil {
				server.logger.Debug("connection dropped: read remote", slog.String("err", err.Error()))
				return err
			}
			if err := wsutil.WriteServerMessage(wsConn, ws.OpBinary, buf[:n]); err != nil {
				server.logger.Debug("connection dropped: write to client", slog.String("err", err.Error()))
				return err
			}
		}
	})

	return eg.Wait()
}

func getParams(request *http.Request) (string, string, error) {
	network := request.URL.Query().Get("net")
	if network != "tcp" && network != "udp" {
		return "", "", errors.New("invalid data")
	}

	addr := request.URL.Query().Get("addr")

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", err
	}

	return network, net.JoinHostPort(host, port), nil
}
