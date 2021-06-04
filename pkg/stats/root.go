package stats

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
)

// RootCommand dispatches all "stats" commands.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new top level "stats" command.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("stats", "View statistics (historical and realtime) for a Fastly service")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	panic("unreachable")
}
