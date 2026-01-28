package digitalocean

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/digitalocean"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand wraps the ListCommand from the newcmd package.
type ListCommand struct {
	*newcmd.ListCommand
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{newcmd.NewListCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service logging digitalocean list' command instead.")
	return c.ListCommand.Exec(in, out)
}
