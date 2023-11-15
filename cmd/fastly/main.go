// Package main is the entry point for the Fastly CLI.
package main

import (
	"os"

	"github.com/fastly/cli/pkg/app"
	fsterr "github.com/fastly/cli/pkg/errors"
)

func main() {
	if err := app.Run(os.Args, os.Stdin); err != nil {
		if skipExit := fsterr.Process(err, os.Args, os.Stdout); skipExit {
			return
		}
		os.Exit(1)
	}
}
