package usagemetrics

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Export ExportCommand `cmd:"" help:"Export usage metrics as CSV"`
	List   ListCommand   `cmd:"" help:"List usage metrics"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "usage-metrics"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
