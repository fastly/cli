package secretstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
)

// RootNameSecret is the base command name for secret operations.
const RootNameSecret = "secret-store-entry"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command(RootNameSecret, "Manipulate Fastly Secret Store secrets")

	return &c
}

// RootCommand is the parent command for all 'secret' subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	// no flags
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
