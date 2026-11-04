package tags

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" help:"Create an operation tag"`
	Delete DeleteCommand `cmd:"" help:"Delete an operation tag"`
	Get    GetCommand    `cmd:"" help:"Get an operation tag"`
	List   ListCommand   `cmd:"" help:"List all operation tags"`
	Update UpdateCommand `cmd:"" help:"Update an operation tag"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tags"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
