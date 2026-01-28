package cloudfiles

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/cloudfiles"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand wraps the UpdateCommand from the newcmd package.
type UpdateCommand struct {
	*newcmd.UpdateCommand
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{newcmd.NewUpdateCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service logging cloudfiles update' command instead.")
	return c.UpdateCommand.Exec(in, out)
}
