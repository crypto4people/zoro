package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/itsabgr/ge"
	"golang.org/x/sync/errgroup"
)

func validateAddr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	return net.JoinHostPort(host, port), err
}

func (server *Server) serveProxy(writer http.ResponseWriter, req *http.Request, account common.Address) error {

	{
		if server.proxy.conns.Add(1) > int64(server.proxy.max) {
			ge.Assert(server.proxy.conns.Add(-1) >= 0)
			return errors.New("too may connections")
		}
		defer func() {
			ge.Assert(server.proxy.conns.Add(-1) >= 0)
		}()
	}

	ctx := req.Context()

	reporting, err := server.prepareReport(ctx, account)
	if err != nil {
		return err
	}

	if err = reporting.Report(ctx, 10); err != nil {
		_ = reporting.Close()
		return err
	}

	network := req.URL.Query().Get("net")
	switch network {
	case "tcp", "udp":
	default:
		_ = reporting.Close()
		return errors.New("invalid network")
	}
	addr, err := validateAddr(req.URL.Query().Get("addr"))
	if err != nil {
		_ = reporting.Close()
		return err
	}

	return server.doProxy(req.Context(), writer, req, reporting, network, addr)

}

func (server *Server) doProxy(ctx context.Context, writer http.ResponseWriter, req *http.Request, reporting *reporting, network, addr string) (err error) {

	ioConn, err := server.proxy.dialer.DialContext(ctx, network, addr)
	if err != nil {
		_ = reporting.Close()
		return err
	}

	wsConn, _, _, err := server.proxy.upgrader.Upgrade(req, writer)
	if err != nil {
		_ = reporting.Close()
		_ = ioConn.Close()
		return err
	}

	go func() {
		defer reporting.Close()
		defer wsConn.Close()
		defer ioConn.Close()

		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() (err error) {
			ticker := time.NewTicker(time.Minute)
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
					if err = reporting.Report(ctx, 60); err != nil {
						_ = wsConn.Close()
						_ = ioConn.Close()
						return err
					}
				}
			}
		})

		eg.Go(func() (err error) {
			var data []byte
			for {
				data, _, err = wsutil.ReadClientData(ioConn)
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
			data := make([]byte, 2048)
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
