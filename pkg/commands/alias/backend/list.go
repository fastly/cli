package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand wraps the ListCommand from the servicebackend package.
type ListCommand struct {
	*servicebackend.ListCommand
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{servicebackend.NewListCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service backend list' command instead.")
	return c.ListCommand.Exec(in, out)
}
