package errors

import (
	"errors"
	"io"

	"github.com/fatih/color"

	"github.com/fastly/cli/v10/pkg/text"
)

// Process persists the error log to disk and deduces the error type.
func Process(err error, args []string, out io.Writer) (skipExit bool) {
	text.Break(out)

	// NOTE: We persist any error log entries to disk before attempting to handle
	// a possible error response from app.Run as there could be errors recorded
	// during the execution flow but were otherwise handled without bubbling an
	// error back the call stack, and so if the user still experiences something
	// unexpected we will have a record of any errors that happened along the way.
	logErr := Log.Persist(LogPath, args[1:])
	if logErr != nil {
		Deduce(logErr).Print(color.Error)
	}

	// IMPORTANT: Deduce/Print needs to happen before checking for Skip.
	// This is so the help output can be printed.
	Deduce(err).Print(color.Error)

	exitError := SkipExitError{}
	if errors.As(err, &exitError) {
		return exitError.Skip
	}
	return false
}
