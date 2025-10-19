package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn struct {
	conn       net.Conn
	remoteAddr net.Addr
}

func (client *Client) Dial(ctx context.Context, net, addr string) (net.Conn, error) {
	dialAddr := fmt.Sprintf("%s/wsc?user=%s&net=%s&addr=%s", client.endpoint, client.account.Hex(), net, addr)
	conn, _, _, err := ws.Dial(ctx, dialAddr)
	if err != nil {
		return Conn{}, err
	}
	return Conn{conn: conn}, nil
}

func (conn Conn) Close() error {
	return conn.conn.Close()
}

func (conn Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

func (conn Conn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}

func (conn Conn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn Conn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn Conn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}

func (conn Conn) Read(b []byte) (n int, err error) {
	data, opCode, err := wsutil.ReadServerData(conn.conn)
	if opCode != ws.OpBinary || err != nil {
		return 0, err
	}

	copy(b, data)
	return len(data), nil
}

func (conn Conn) Write(b []byte) (n int, err error) {
	return len(b), wsutil.WriteClientBinary(conn.conn, b)
}
