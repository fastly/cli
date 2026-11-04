package timeseries

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
type RootCommand struct {
	argparser.Base

	Get GetCommand `cmd:"" help:"Get workspace-level time series metrics"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "time-series"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
