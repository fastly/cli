package item

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a custom dashboard item"`
	Describe DescribeCommand `cmd:"" aliases:"add" help:"Describe a custom dashboard item"`
	Update   UpdateCommand   `cmd:"" aliases:"add" help:"Update a custom dashboard item"`
	Delete   DeleteCommand   `cmd:"" aliases:"add" help:"Delete a custom dashboard item"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "item"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
