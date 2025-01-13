package stats

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand dispatches all "stats" commands.
type RootCommand struct {
	argparser.Base
}

// CommandName is the string to be used to invoke this command
const CommandName = "stats"

// NewRootCommand returns a new top level "stats" command.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "View historical and realtime statistics for a Fastly service")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
