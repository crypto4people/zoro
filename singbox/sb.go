package singbox

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	box "github.com/sagernet/sing-box"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/network"
)

type Mode uint8

const (
	ModeSocks Mode = iota
	ModeTun
)

type Settings struct {
	SocksAddr *string

	TunAddr *string

	WSCPath string // default: "/wsc"
	UseTLS  bool   // default: false
}

func New(server string, priv *ecdsa.PrivateKey, set Settings, mode Mode) (*Builder, error) {
	uri, err := url.Parse(server)
	if err != nil {
		return nil, fmt.Errorf("parse server: %w", err)
	}

	if uri.Scheme != "http" && uri.Scheme != "https" {
		return nil, fmt.Errorf("server must start with http:// or https://")
	}

	host := uri.Hostname()
	if host == "" {
		return nil, fmt.Errorf("missing host in server")
	}

	portStr := uri.Port()
	if portStr == "" {
		if uri.Scheme == "https" {
			portStr = "443"
		} else {
			portStr = "80"
		}
	}

	p64, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	if set.WSCPath == "" {
		set.WSCPath = "/wsc"
	}

	return &Builder{
		serverURL:  uri,
		serverHost: host,
		serverPort: uint16(p64),
		priv:       priv,
		mode:       mode,
		set:        set,
	}, nil
}

func (builder *Builder) NewBox(ctx context.Context) (*box.Box, error) {
	opts, err := builder.Options()
	if err != nil {
		return nil, err
	}
	inReg := include.InboundRegistry()
	outReg := include.OutboundRegistry()
	endReg := include.EndpointRegistry()
	svcReg := include.ServiceRegistry()
	dnsReg := include.DNSTransportRegistry()
	sbCtx := box.Context(ctx, inReg, outReg, endReg, dnsReg, svcReg)

	return box.New(box.Options{
		Context: sbCtx,
		Options: *opts,
	})
}

func (builder *Builder) Options() (*option.Options, error) {
	inb, err := builder.buildInbound()
	if err != nil {
		return nil, err
	}

	outb := builder.buildOutbounds()

	rules := []option.Rule{
		{
			Type: C.RuleTypeDefault,
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					Network: []string{network.NetworkUDP},
				},
				RuleAction: option.RuleAction{
					RouteOptions: option.RouteActionOptions{
						Outbound: "direct",
					},
				},
			},
		},
	}

	return &option.Options{
		Inbounds:  []option.Inbound{inb},
		Outbounds: outb,
		Route: &option.RouteOptions{
			Rules: rules,
		},
	}, nil
}

func (builder *Builder) JSON() ([]byte, error) {
	inbJSON, err := builder.buildInboundJSON()
	if err != nil {
		return nil, err
	}
	outbJSON := builder.buildOutboundsJSON()

	cfg := map[string]any{
		"inbounds":  []any{inbJSON},
		"outbounds": outbJSON,
		"route": map[string]any{
			"rules": []any{
				map[string]any{"type": "default", "outbound": "wsc-out"},
			},
		},
	}

	return json.MarshalIndent(cfg, "", "  ")
}
