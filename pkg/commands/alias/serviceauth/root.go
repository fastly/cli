package serviceauth

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" hidden:"" help:"Create service authorization"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" hidden:"" help:"Delete service authorization"`
	Describe DescribeCommand `cmd:"" aliases:"get" hidden:"" help:"Show service authorization"`
	List     ListCommand     `cmd:"" hidden:"" help:"List service authorizations"`
	Update   UpdateCommand   `cmd:"" hidden:"" help:"Update service authorization"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "service-auth"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
