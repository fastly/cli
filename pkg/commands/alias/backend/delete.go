package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the servicebackend package.
type DeleteCommand struct {
	*servicebackend.DeleteCommand
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{servicebackend.NewDeleteCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service backend delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
