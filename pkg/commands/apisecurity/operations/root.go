package operations

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// CommandName is the string to be used to invoke this command.
const CommandName = "operations"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, globals *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command(CommandName, "Manage operations associated with services")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
