package configstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// CommandName is the string to be used to invoke this command.
const CommandName = "config-store-entry"

// RootCommand is the parent command for all subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"insert" help:"Create a new config store item"`
	Delete   DeleteCommand   `cmd:"" help:"Delete a config store item"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Retrieve a single config store item"`
	List     ListCommand     `cmd:"" help:"List config store items"`
	Update   UpdateCommand   `cmd:"" help:"Update a config store item"`
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
