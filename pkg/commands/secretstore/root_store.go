package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
)

// RootNameStore is the base command name for secret store operations.
const RootNameStore = "secret-store"

// NewStoreRootCommand returns a new command registered in the parent.
func NewStoreRootCommand(parent cmd.Registerer, globals *config.Data) *StoreRootCommand {
	c := StoreRootCommand{
		Base: cmd.Base{
			Globals: globals,
		},
	}

	c.CmdClause = parent.Command(RootNameStore, "Manipulate Fastly Secret Stores")

	return &c
}

// StoreRootCommand is the parent command for all 'store' subcommands.
// It should be installed under the primary root command.
type StoreRootCommand struct {
	cmd.Base
	// no flags
}

// Exec implements the command interface.
func (c *StoreRootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
