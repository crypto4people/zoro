package client

import (
	"github.com/urfave/cli/v3"
)

var serverFlag = &cli.StringFlag{Name: "server", Required: true, Usage: "Server base URL (e.g. https://zoro.example.com)"}
var accountFlag = &cli.StringFlag{Name: "account", Required: true, Usage: "Account address/string"}

var clientInfoCmd = &cli.Command{
	Name:  "info",
	Usage: "Fetch and display server info",
	Flags: []cli.Flag{
		serverFlag,
		accountFlag,
	},
	Action: infoAction,
}

var clientSubsCmd = &cli.Command{
	Name:  "subs",
	Usage: "List subscriptions for an account",
	Flags: []cli.Flag{
		serverFlag,
		accountFlag,
	},
	Action: subsAction,
}

var clientConnectCmd = &cli.Command{
	Name:  "connect",
	Usage: "Start local SOCKS5 and tunnel via sing-box",
	Flags: []cli.Flag{
		serverFlag,
		&cli.StringFlag{Name: "local-socks", Value: "127.0.0.1:1080"},
	},
	Action: connectAction,
}

var clientTxApplyCmd = &cli.Command{
	Name:  "apply",
	Usage: "Create, sign and submit a payment transaction",
	Flags: []cli.Flag{
		serverFlag,
		&cli.StringFlag{Name: "duration", Required: true},
	},
	Action: applyAction,
}

var clientTxCheckCmd = &cli.Command{
	Name:  "check",
	Usage: "Check a transaction receipt",
	Flags: []cli.Flag{
		serverFlag,
		&cli.StringFlag{Name: "hash", Required: true},
	},
	Action: checkAction,
}

var clientTxCmd = &cli.Command{
	Name:     "tx",
	Usage:    "Buy (send) or check a transaction",
	Commands: []*cli.Command{clientTxApplyCmd, clientTxCheckCmd},
}

var Cmd = &cli.Command{
	Name:  "client",
	Usage: "client → Get Server Info, Watch your subs, Buy a new subscription, Connect",
	Commands: []*cli.Command{
		clientInfoCmd,
		clientSubsCmd,
		clientConnectCmd,
		clientTxCmd,
	},
}
