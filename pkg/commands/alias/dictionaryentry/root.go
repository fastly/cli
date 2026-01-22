package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand wraps the RootCommand from the servicedictionaryentry package.
type RootCommand struct {
	*servicedictionaryentry.RootCommand
}

// NewRootCommand returns a usable command registered under the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	c := RootCommand{servicedictionaryentry.NewRootCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	return c.RootCommand.Exec(in, out)
}
