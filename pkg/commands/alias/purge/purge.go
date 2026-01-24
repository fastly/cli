package purge

import (
	"io"

	servicepurge "github.com/fastly/cli/pkg/commands/service/purge"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// PurgeCommand wraps the PurgeCommand from the servicepurge package.
type Command struct {
	*servicepurge.PurgeCommand
}

// NewCommand returns a usable command registered under the parent.
func NewCommand(parent argparser.Registerer, g *global.Data) *Command {
	c := Command{servicepurge.NewPurgeCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *Command) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service purge' command instead.")
	return c.PurgeCommand.Exec(in, out)
}
