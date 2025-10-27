package client

import (
	"bufio"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/json/badoption"
	"github.com/sagernet/sing/common/network"
)

func connect(ctx context.Context, server, account, socksAddr string) (*box.Box, error) {
	if !strings.Contains(socksAddr, ":") {
		return nil, errors.New("invalid listen address")
	}

	inReg := include.InboundRegistry()
	outReg := include.OutboundRegistry()
	endReg := include.EndpointRegistry()
	svcReg := include.ServiceRegistry()
	dnsReg := include.DNSTransportRegistry()

	sbCtx := box.Context(ctx, inReg, outReg, endReg, dnsReg, svcReg)

	socksAddrSplitted := strings.Split(socksAddr, ":")

	socksListen := netip.MustParseAddr(socksAddrSplitted[0])
	socksPort, err := strconv.ParseUint(socksAddrSplitted[1], 10, 16)
	if err != nil {
		return nil, err
	}

	serverSplitted := strings.Split(server, ":")
	if len(serverSplitted) < 3 {
		return nil, fmt.Errorf("server must be like http://host:port or https://host:port")
	}

	host := strings.ReplaceAll(serverSplitted[1], "/", "")
	serverPort, err := strconv.ParseUint(serverSplitted[2], 10, 16)
	if err != nil {
		return nil, err
	}

	inbounds := []option.Inbound{
		{
			Type: "socks",
			Tag:  "socks-in",
			Options: &option.SocksInboundOptions{
				ListenOptions: option.ListenOptions{
					Listen:     (*badoption.Addr)(&socksListen),
					ListenPort: uint16(socksPort),
				},
			},
		},
	}

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
				Auth:                        account,
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

	box, err := box.New(box.Options{
		Context: sbCtx,
		Options: option.Options{
			Inbounds:  inbounds,
			Outbounds: outbounds,
			Route: &option.RouteOptions{
				Rules: rules,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return box, nil
}

func signTx(priv *ecdsa.PrivateKey, nonce uint64, to common.Address, chainID, amount, gasPrice *big.Int) (*types.Transaction, error) {
	if gasPrice == nil || gasPrice.Sign() <= 0 {
		return nil, fmt.Errorf("invalid gas price")
	}

	if amount == nil || amount.Sign() < 0 {
		return nil, fmt.Errorf("invalid amount")
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    new(big.Int).Set(amount),
		Gas:      21_000,
		GasPrice: new(big.Int).Set(gasPrice),
	})

	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), priv)
	if err != nil {
		return nil, fmt.Errorf("sign: %w", err)
	}
	return signed, nil
}

func confirm(msg string) (bool, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return false, fmt.Errorf("no TTY available for confirmation: %w", err)
	}
	defer tty.Close()

	fmt.Fprintf(os.Stdout, "%s [y/N]: ", msg)
	reader := bufio.NewReader(tty)
	line, _ := reader.ReadString('\n')
	resp := strings.TrimSpace(strings.ToLower(line))
	return resp == "y" || resp == "yes", nil
}
