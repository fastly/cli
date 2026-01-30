package sumologic

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/sumologic"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand wraps the CreateCommand from the newcmd package.
type CreateCommand struct {
	*newcmd.CreateCommand
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{newcmd.NewCreateCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service logging sumologic create' command instead.")
	return c.CreateCommand.Exec(in, out)
}
