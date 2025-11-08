package indexer

import "github.com/urfave/cli/v3"

var runCmd = &cli.Command{
	Name:  "run",
	Usage: "run zoro indexer",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Required: true, Usage: "zoro-indexer.json"},
	},
	Action: runAction,
}

var Cmd = &cli.Command{
	Name:  "indexer",
	Usage: "indexer",
	Commands: []*cli.Command{
		runCmd,
	},
}
