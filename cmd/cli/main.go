package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/crypto4people/zoro/cli"
)

func main() {
	app := cli.App()

	if err := app.Run(context.Background(), os.Args); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}
