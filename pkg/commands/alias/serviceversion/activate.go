package serviceversion

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/version"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ActivateCommand wraps the ActivateCommand from the newcmd package.
type ActivateCommand struct {
	*newcmd.ActivateCommand
}

// NewActivateCommand returns a usable command registered under the parent.
func NewActivateCommand(parent argparser.Registerer, g *global.Data) *ActivateCommand {
	c := ActivateCommand{newcmd.NewActivateCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *ActivateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service version activate' command instead.")
	return c.ActivateCommand.Exec(in, out)
}
