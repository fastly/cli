package errors

import (
	"io"

	"github.com/fastly/cli/pkg/text"
)

// ExitError is an error that can cause the os.Exit(1) to be skipped.
// An example is 'help' output (e.g. --help).
type ExitError struct {
	Skip bool
	Err  error
}

// Unwrap returns the inner error.
func (ee ExitError) Unwrap() error {
	return ee.Err
}

// Error prints the inner error string.
func (ee ExitError) Error() string {
	if ee.Err == nil {
		return ""
	}
	return ee.Err.Error()
}

// Print the error to the io.Writer for human consumption.
// The inner error is always printed via text.Output with an "Error: " prefix
// and a "." suffix.
func (ee ExitError) Print(w io.Writer) {
	if ee.Err != nil {
		text.Error(w, "%s.", ee.Err.Error())
	}
}
