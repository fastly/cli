package threshold

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a workspace threshold"`
	Delete DeleteCommand `cmd:"" help:"Deletes a workspace threshold"`
	Get    GetCommand    `cmd:"" help:"Retrieves a workspace threshold"`
	List   ListCommand   `cmd:"" help:"List workspace thresholds"`
	Update UpdateCommand `cmd:"" help:"Update a workspace threshold"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "threshold"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
