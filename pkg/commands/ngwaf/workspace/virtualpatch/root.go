package virtualpatch

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	List     ListCommand   `cmd:"" help:"List virtual patches in a workspace"`
	Update   UpdateCommand `cmd:"" help:"Update a virtual patch"`
	Retrieve GetCommand    `cmd:"" aliases:"get" help:"Retrieve a virtual patch"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "virtualpatch"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
