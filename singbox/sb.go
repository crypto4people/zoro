package singbox

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"strconv"
	"strings"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sagernet/sing/common/network"
)

type SingBoxMode uint8

const (
	ModeSocks SingBoxMode = iota
	ModeTun
)

type Settings struct {
	SocksAddr *string
	TunAddr   *string
}

func Config(ctx context.Context, server string, privateKey *ecdsa.PrivateKey, settings *Settings, mode SingBoxMode) (*option.Options, error) {
	serverSplitted := strings.Split(server, ":")
	if len(serverSplitted) < 3 {
		return nil, fmt.Errorf("server must be like http://host:port or https://host:port")
	}

	host := strings.ReplaceAll(serverSplitted[1], "/", "")
	serverPort, err := strconv.ParseUint(serverSplitted[2], 10, 16)
	if err != nil {
		return nil, err
	}

	var socksListen netip.Addr
	var socksPort uint64
	var inbound option.Inbound
	if mode == ModeSocks {
		if settings.SocksAddr == nil {
			return nil, errors.New("socks address is required in socks mode")
		}

		socksAddr := *settings.SocksAddr
		var err error
		if !strings.Contains(socksAddr, ":") {
			return nil, errors.New("invalid listen address")
		}

		socksAddrSplitted := strings.Split(socksAddr, ":")

		socksListen = netip.MustParseAddr(socksAddrSplitted[0])
		socksPort, err = strconv.ParseUint(socksAddrSplitted[1], 10, 16)
		if err != nil {
			return nil, err
		}

		inbound = option.Inbound{
			Type: "socks",
			Tag:  "in",
			Options: &option.SocksInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     (*badoption.Addr)(&socksListen),
					ListenPort: uint16(socksPort),
				},
			},
		}
	} else {
		if settings.TunAddr == nil {
			return nil, errors.New("tun address is required in tun mode")
		}

		tunAddr := *settings.TunAddr
		if !strings.Contains(tunAddr, "/") {
			return nil, errors.New("invalid tun address")
		}

		inbound = option.Inbound{
			Type: "tun",
			Tag:  "in",
			Options: &option.TunInboundOptions{
				InterfaceName: "singbox-wsc1",
				Address:       badoption.Listable[netip.Prefix]{netip.MustParsePrefix(tunAddr)},
				AutoRoute:     true,
				AutoRedirect:  true,
			},
		}
	}

	inbounds := []option.Inbound{inbound}

	outbounds := []option.Outbound{
		{
			Type: "wsc",
			Tag:  "wsc-out",
			Options: &option.WSCOutboundOptions{
				ServerOptions: option.ServerOptions{
					Server:     host,
					ServerPort: uint16(serverPort),
				},
				DialerOptions:               option.DialerOptions{},
				OutboundTLSOptionsContainer: option.OutboundTLSOptionsContainer{TLS: &option.OutboundTLSOptions{Enabled: false}},
				Auth:                        hex.EncodeToString(privateKey.D.Bytes()),
				Path:                        "/wsc",
			},
		},
		{Type: "direct", Tag: "direct"},
		{Type: "block", Tag: "block"},
	}

	rules := []option.Rule{
		{
			DefaultOptions: option.DefaultRule{
				RawDefaultRule: option.RawDefaultRule{
					Protocol: []string{network.NetworkUDP},
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
		Inbounds:  inbounds,
		Outbounds: outbounds,
		Route: &option.RouteOptions{
			Rules: rules,
		},
	}, nil
}

func NewBox(ctx context.Context, server string, privateKey *ecdsa.PrivateKey, settings *Settings, mode SingBoxMode) (*box.Box, error) {
	opts, err := Config(ctx, server, privateKey, settings, mode)
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
