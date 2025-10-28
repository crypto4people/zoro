package utility

import "github.com/urfave/cli/v3"

var convertCmd = &cli.Command{
	Name:   "gen-address",
	Usage:  "generates address from private key",
	Action: convertAction,
}

var Cmd = &cli.Command{
	Name:  "utility",
	Usage: "small utils to help you use zoro",
	Commands: []*cli.Command{
		convertCmd,
	},
}
