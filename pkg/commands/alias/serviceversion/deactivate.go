package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeactivateCommand wraps the DeactivateCommand from the newcmd package.
type DeactivateCommand struct {
	*newcmd.DeactivateCommand
}

// NewDeactivateCommand returns a usable command registered under the parent.
func NewDeactivateCommand(parent argparser.Registerer, g *global.Data) *DeactivateCommand {
	c := DeactivateCommand{newcmd.NewDeactivateCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DeactivateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service version deactivate' command instead.")
	return c.DeactivateCommand.Exec(in, out)
}
