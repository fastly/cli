package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootNameStore is the base command name for secret store operations.
const RootNameStore = "secret-store"

// CommandName is the string to be used to invoke this command.
const CommandName = "secret-store"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command(CommandName, "Manipulate Fastly Secret Stores")

	return &c
}

// RootCommand is the parent command for all 'store' subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
