package healthcheck

import (
	"io"

	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand wraps the RootCommand from the servicehealthcheck package.
type RootCommand struct {
	*servicehealthcheck.RootCommand
}

// NewRootCommand returns a usable command registered under the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{servicehealthcheck.NewRootCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	return c.RootCommand.Exec(in, out)
}
