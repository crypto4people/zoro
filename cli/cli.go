package cli

import (
	"zoro/cli/client"
	"zoro/cli/server"
	"zoro/cli/utility"

	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	return &cli.Command{
		EnableShellCompletion: true,
		Name:                  "zoro-cli",
		Usage:                 "CLI for Zoro — run your own server or connect to Internet freely",
		Commands: []*cli.Command{
			client.Cmd,
			server.Cmd,
			utility.Cmd,
		},
		HideHelpCommand: false,
	}
}
