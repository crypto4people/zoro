package singbox

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
)

type Builder struct {
	serverURL  *url.URL
	serverHost string
	serverPort uint16
	priv       *ecdsa.PrivateKey
	mode       Mode
	set        Settings
}

func (builder *Builder) buildInbound() (option.Inbound, error) {
	switch builder.mode {
	case ModeSocks:
		if builder.set.SocksAddr == nil || !strings.Contains(*builder.set.SocksAddr, ":") {
			return option.Inbound{}, errors.New("socks address is required like 127.0.0.1:1080")
		}
		parts := strings.SplitN(*builder.set.SocksAddr, ":", 2)
		addr := netip.MustParseAddr(parts[0])
		p64, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			return option.Inbound{}, err
		}
		return option.Inbound{
			Type: "socks",
			Tag:  "in",
			Options: &option.SocksInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     (*badoption.Addr)(&addr),
					ListenPort: uint16(p64),
				},
			},
		}, nil

	case ModeTun:
		if builder.set.TunAddr == nil || !strings.Contains(*builder.set.TunAddr, "/") {
			return option.Inbound{}, errors.New("tun address is required like 10.66.0.1/24")
		}
		return option.Inbound{
			Type: "tun",
			Tag:  "in",
			Options: &option.TunInboundOptions{
				Address:      badoption.Listable[netip.Prefix]{netip.MustParsePrefix(*builder.set.TunAddr)},
				AutoRoute:    true,
				AutoRedirect: true,
			},
		}, nil
	default:
		return option.Inbound{}, errors.New("unknown mode")
	}
}

func (builder *Builder) buildOutbounds() []option.Outbound {
	return []option.Outbound{
		{
			Type: "wsc",
			Tag:  "wsc-out",
			Options: &option.WSCOutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     builder.serverHost,
					ServerPort: builder.serverPort,
				},
				DialerOptions: option.DialerOptions{},
				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{
					TLS: &option.OutboundTLSOptions{Enabled: builder.set.UseTLS},
				},
				Auth: hex.EncodeToString(builder.priv.D.Bytes()),
				Path: builder.set.WSCPath,
			},
		},
		{Type: "direct", Tag: "direct"},
		{Type: "block", Tag: "block"},
	}
}

func (builder *Builder) buildInboundJSON() (map[string]any, error) {
	switch builder.mode {
	case ModeSocks:
		if builder.set.SocksAddr == nil {
			return nil, errors.New("socks address is required")
		}
		return map[string]any{
			"type": "socks",
			"tag":  "in",
			"listen": map[string]any{
				"addr": strings.SplitN(*builder.set.SocksAddr, ":", 2)[0],
				"port": func() int {
					p, _ := strconv.Atoi(strings.SplitN(*builder.set.SocksAddr, ":", 2)[1])
					return p
				}(),
			},
		}, nil

	case ModeTun:
		if builder.set.TunAddr == nil {
			return nil, errors.New("tun address is required")
		}
		return map[string]any{
			"type":          "tun",
			"tag":           "in",
			"address":       []string{*builder.set.TunAddr},
			"auto_route":    true,
			"auto_redirect": true,
		}, nil
	default:
		return nil, errors.New("unknown mode")
	}
}

func (builder *Builder) buildOutboundsJSON() []any {
	return []any{
		map[string]any{
			"type":        "wsc",
			"tag":         "wsc-out",
			"server":      builder.serverHost,
			"server_port": int(builder.serverPort),
			"auth":        hex.EncodeToString(builder.priv.D.Bytes()),
			"path":        builder.set.WSCPath,
			"tls": map[string]any{
				"enabled": builder.set.UseTLS,
			},
		},
		map[string]any{"type": "direct", "tag": "direct"},
		map[string]any{"type": "block", "tag": "block"},
	}
}
