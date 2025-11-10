package proxy

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
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
	cancelFuncs = append(cancelFuncs, func() { _ = wsConn.Close() })

	if err := server.authConn(wsConn, addr, network, upstreamAddr); err != nil {
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
	cancelFuncs = append(cancelFuncs, func() { _ = reporting.Close() })

	isUDP := strings.EqualFold(network, "udp")

	var (
		ioConn  net.Conn
		udpConn *net.UDPConn
	)

	if !isUDP {
		ioConn, err = server.proxy.dialer.DialContext(ctx, "tcp", upstreamAddr)
		if err != nil {
			logger.Debug("connection dropped due to failure on dialing remote", slog.String("error", err.Error()))
			return err
		}
		cancelFuncs = append(cancelFuncs, func() { _ = ioConn.Close() })
	} else {
		udpConn, err = net.ListenUDP("udp", nil)
		if err != nil {
			logger.Debug("connection dropped: listen udp", slog.String("err", err.Error()))
			return err
		}
		cancelFuncs = append(cancelFuncs, func() { _ = udpConn.Close() })
	}

	eg, egCtx := errgroup.WithContext(cgCtx)

	var once sync.Once
	closeAll := func() {
		once.Do(func() {
			if ioConn != nil {
				ioConn.Close()
			}
			if udpConn != nil {
				udpConn.Close()
			}
			wsConn.Close()
		})
	}

	var wsMu sync.Mutex
	writeWS := func(op ws.OpCode, p []byte) error {
		wsMu.Lock()
		defer wsMu.Unlock()
		return wsutil.WriteServerMessage(wsConn, op, p)
	}

	eg.Go(func() error {
		ticker := time.NewTicker(server.proxy.reporting.interval)
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
					server.proxy.reporting.interval,
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

			if !isUDP {
				if _, err := ioConn.Write(data); err != nil {
					server.logger.Debug("connection dropped: write to remote", slog.String("err", err.Error()))
					return err
				}
			} else {
				dst, payload, err := parseUDPFrame(data)
				if err != nil {
					server.logger.Debug("connection dropped: parse udp", slog.String("err", err.Error()))
					return err
				}

				if _, err := udpConn.WriteToUDPAddrPort(payload, dst); err != nil {
					server.logger.Debug("connection dropped: udp write", slog.String("err", err.Error()))
					return err
				}
			}
		}
	})

	if !isUDP {
		eg.Go(func() error {
			defer closeAll()
			buf := make([]byte, 64<<10)
			for {
				n, err := ioConn.Read(buf)
				if err != nil {
					server.logger.Debug("connection dropped: read remote", slog.String("err", err.Error()))
					return err
				}

				if err := writeWS(ws.OpBinary, buf[:n]); err != nil {
					server.logger.Debug("connection dropped: write to client", slog.String("err", err.Error()))
					return err
				}
			}
		})
	} else {
		eg.Go(func() error {
			defer closeAll()
			buf := make([]byte, 64<<10)
			for {
				_ = udpConn.SetReadDeadline(time.Now().Add(2 * time.Minute))
				n, src, err := udpConn.ReadFromUDPAddrPort(buf)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						select {
						case <-egCtx.Done():
							return egCtx.Err()
						default:
							continue
						}
					}
					server.logger.Debug("connection dropped: udp read", slog.String("err", err.Error()))
					return err
				}

				frame := composeUDPFrame(src, buf[:n])
				if err := writeWS(ws.OpBinary, frame); err != nil {
					server.logger.Debug("connection dropped: write to client", slog.String("err", err.Error()))
					return err
				}
			}
		})
	}

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

func parseUDPFrame(data []byte) (netip.AddrPort, []byte, error) {
	if len(data) <= 18 {
		return netip.AddrPort{}, nil, errors.New("invalid data: too short")
	}

	ip16 := data[:16]
	port := binary.BigEndian.Uint16(data[16:18])
	payload := data[18:]

	var ip netip.Addr
	if isAllZero(ip16[:12]) {
		ip = netip.AddrFrom4([4]byte(ip16[12:16]))
	} else {
		ip = netip.AddrFrom16([16]byte(ip16))
	}

	return netip.AddrPortFrom(ip, port), payload, nil
}

func composeUDPFrame(addr netip.AddrPort, payload []byte) []byte {
	ip := addr.Addr()

	ip16 := make([]byte, 16)
	if ip.Is4() {
		v4 := ip.As4()
		copy(ip16[12:16], v4[:])
	} else if ip.Is6() {
		v6 := ip.As16()
		copy(ip16[:], v6[:])
	}

	frame := make([]byte, 18+len(payload))
	copy(frame[0:16], ip16)
	binary.BigEndian.PutUint16(frame[16:18], addr.Port())
	copy(frame[18:], payload)

	return frame
}

func isAllZero(b []byte) bool {
	for _, x := range b {
		if x != 0 {
			return false
		}
	}
	return true
}
