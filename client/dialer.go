package client

import (
	"context"
	"net"
	"net/url"
	"path"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Conn struct {
	conn       net.Conn
	remoteAddr net.Addr
}

func (client *Client) Dial(ctx context.Context, network, addr string) (net.Conn, error) {
	url, err := url.Parse(client.endpoint)
	if err != nil {
		return nil, err
	}
	url.Path = path.Join(url.Path, "wsc")

	query := url.Query()
	query.Set("user", client.account.Hex())
	query.Set("net", network)
	query.Set("addr", addr)
	url.RawQuery = query.Encode()

	conn, _, _, err := ws.Dial(ctx, url.String())
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn}, nil
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

func (conn Conn) Read(payload []byte) (n int, err error) {
	data, opCode, err := wsutil.ReadServerData(conn.conn)
	if opCode != ws.OpBinary || err != nil {
		return 0, err
	}

	copy(payload, data)
	return len(data), nil
}

func (conn Conn) Write(payload []byte) (n int, err error) {
	return len(payload), wsutil.WriteClientBinary(conn.conn, payload)
}
