package alerts

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/alert"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the newcmd package.
type DeleteCommand struct {
	*newcmd.DeleteCommand
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{newcmd.NewDeleteCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service alert delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
