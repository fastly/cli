package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
)

// RootNameStore is the base command name for secret store operations.
const RootNameStore = "secret-store"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command(RootNameStore, "Manipulate Fastly Secret Stores")

	return &c
}

// RootCommand is the parent command for all 'store' subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	// no flags
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
