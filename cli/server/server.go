package server

import "github.com/urfave/cli/v3"

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run zoro server",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Required: true, Usage: "zoro.json"},
	},
	Action: runAction,
}

var Cmd = &cli.Command{
	Name:  "server",
	Usage: "server → Setup your own zoro node",
	Commands: []*cli.Command{
		runCmd,
	},
}
