package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand wraps the UpdateCommand from the servicedictionaryentry package.
type UpdateCommand struct {
	*servicedictionaryentry.UpdateCommand
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{servicedictionaryentry.NewUpdateCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated(out, "Use the 'service dictionary-entry update' command instead.")
	return c.UpdateCommand.Exec(in, out)
}
