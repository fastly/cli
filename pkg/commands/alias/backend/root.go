package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand wraps the RootCommand from the servicebackend package.
type RootCommand struct {
	*servicebackend.RootCommand
}

// NewRootCommand returns a usable command registered under the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{servicebackend.NewRootCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	return c.RootCommand.Exec(in, out)
}
