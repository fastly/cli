package configstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
)

// RootName is the base command name.
const RootName = "config-store-entry"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command(RootName, "Manipulate Fastly Config Store items")

	return &c
}

// RootCommand is the parent command for all subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	// no flags
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
