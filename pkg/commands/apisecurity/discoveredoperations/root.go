package discoveredoperations

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	List   ListCommand   `cmd:"" help:"List discovered operations"`
	Update UpdateCommand `cmd:"" help:"Update the status of discovered operation(s)"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "discovered-operations"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
