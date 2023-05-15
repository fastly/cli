package kvstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
)

// RootName is the base command name for kv store operations.
const RootName = "kv-store"

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	// no flags
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(RootName, "Manipulate Fastly KV Stores").Alias("object-store")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
