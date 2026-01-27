package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the servicedictionaryentry package.
type DeleteCommand struct {
	*servicedictionaryentry.DeleteCommand
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{servicedictionaryentry.NewDeleteCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service dictionary-entry delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
