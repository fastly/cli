package domain

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a domain"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a domain"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show detailed information about a domain"`
	List     ListCommand     `cmd:"" help:"List domains"`
	Update   UpdateCommand   `cmd:"" help:"Update a domain"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "domain"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
