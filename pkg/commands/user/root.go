package user

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a user of the Fastly API and web interface"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a user of the Fastly API and web interface"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Get a specific user of the Fastly API and web interface"`
	List     ListCommand     `cmd:"" help:"List all users from a specified customer id"`
	Update   UpdateCommand   `cmd:"" help:"Update a user of the Fastly API and web interface"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "user"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
