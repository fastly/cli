package tsigkey

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a TSIG key"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a TSIG key"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Describe a TSIG key"`
	List     ListCommand     `cmd:"" help:"List TSIG keys"`
	Update   UpdateCommand   `cmd:"" help:"Update a TSIG key"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tsig-key"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
