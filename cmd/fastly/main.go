// Package main is the entry point for the Fastly CLI.
package main

import (
	"errors"
	"os"

	"github.com/fatih/color"

	"github.com/fastly/cli/pkg/app"
	fsterr "github.com/fastly/cli/pkg/errors"
)

func main() {
	if err := app.Run(os.Args, os.Stdin); err != nil {
		if skipExit := processErr(err, os.Args); skipExit {
			return
		}
		os.Exit(1)
	}
}

// processErr persists the error log to disk and deduces the error type.
func processErr(err error, args []string) (skipExit bool) {
	// NOTE: We persist any error log entries to disk before attempting to handle
	// a possible error response from app.Run as there could be errors recorded
	// during the execution flow but were otherwise handled without bubbling an
	// error back the call stack, and so if the user still experiences something
	// unexpected we will have a record of any errors that happened along the way.
	logErr := fsterr.Log.Persist(fsterr.LogPath, args[1:])
	if logErr != nil {
		fsterr.Deduce(logErr).Print(color.Error)
	}
	exitError := fsterr.SkipExitError{}
	if errors.As(err, &exitError) {
		return exitError.Skip
	}
	fsterr.Deduce(err).Print(color.Error)
	return false
}
