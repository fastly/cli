package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand wraps the CreateCommand from the servicebackend package.
type CreateCommand struct {
	*servicebackend.CreateCommand
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{servicebackend.NewCreateCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service backend create' command instead.")
	return c.CreateCommand.Exec(in, out)
}
