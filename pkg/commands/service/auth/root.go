package auth

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create service authorization"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete service authorization"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show service authorization"`
	List     ListCommand     `cmd:"" help:"List service authorizations"`
	Update   UpdateCommand   `cmd:"" help:"Update service authorization"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "auth"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
