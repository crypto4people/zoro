package cli

import (
	"zoro/cli/client"
	"zoro/cli/server"

	"github.com/urfave/cli/v3"
)

func App() *cli.Command {
	return &cli.Command{
		Name:  "zoro-cli",
		Usage: "CLI for Zoro — run your own server or connect to Internet freely",
		Commands: []*cli.Command{
			client.Cmd,
			server.Cmd,
		},
		HideHelpCommand: false,
	}
}
