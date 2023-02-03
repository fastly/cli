package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
)

// KeysRootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type KeysRootCommand struct {
	cmd.Base
	// no flags
}

// NewKeysRootCommand returns a new command registered in the parent.
func NewKeysRootCommand(parent cmd.Registerer, globals *config.Data) *KeysRootCommand {
	var c KeysRootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("object-store-keys", "Manipulate Fastly Object Store keys")
	return &c
}

// Exec implements the command interface.
func (c *KeysRootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
