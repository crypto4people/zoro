package cli

import (
	"context"
	"net/http"
	"zoro/proxy"

	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	return &cli.Command{
		Name:  "zoro-cli",
		Usage: "CLI client for Zoro — get server info, list subs, connect via sing-box",
		Commands: []*cli.Command{
			cmdGetServer(),
			cmdGetSubs(),
			cmdConnect(),
		},
		HideHelpCommand: false,
	}
}

func cmdGetServer() *cli.Command {
	return &cli.Command{
		Name:  "info",
		Usage: "Fetch and display server info",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server", Required: true, Usage: "Server base URL (e.g. https://zoro.example.com)"},
			&cli.StringFlag{Name: "account", Required: true, Usage: "Account address/string"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			server := cmd.String("server")
			account := cmd.String("account")

			client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(account))
			info, err := client.GetInfo(ctx)
			if err != nil {
				return err
			}

			printServerInfo(info)
			return nil
		},
	}
}

func cmdGetSubs() *cli.Command {
	return &cli.Command{
		Name:  "subs",
		Usage: "List subscriptions for an account",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server", Required: true, Usage: "Server base URL"},
			&cli.StringFlag{Name: "account", Required: true, Usage: "Account address/string"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			server := cmd.String("server")
			account := cmd.String("account")

			client := proxy.NewClient(http.DefaultClient, server, common.HexToAddress(account))
			subs, err := client.GetSubs(ctx)
			if err != nil {
				return err
			}

			printSubs(subs)
			return nil
		},
	}
}

func cmdConnect() *cli.Command {
	return &cli.Command{
		Name:  "connect",
		Usage: "Start local SOCKS5 and tunnel via sing-box",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server", Required: true},
			&cli.StringFlag{Name: "account", Required: true},
			&cli.StringFlag{Name: "local-socks", Value: "127.0.0.1:1080"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			server := cmd.String("server")
			account := cmd.String("account")
			localSocks := cmd.String("local-socks")

			box, err := connect(ctx, server, account, localSocks)
			if err != nil {
				return err
			}

			box.Start()
			<-ctx.Done()
			return ctx.Err()
		},
	}
}
