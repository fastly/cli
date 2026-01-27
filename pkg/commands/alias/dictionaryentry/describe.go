package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the servicedictionaryentry package.
type DescribeCommand struct {
	*servicedictionaryentry.DescribeCommand
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{servicedictionaryentry.NewDescribeCommand(parent, g)}
	c.CmdClause.Hidden()
	return &c
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated(out, "Use the 'service dictionary-entry describe' command instead.")
	}
	return c.DescribeCommand.Exec(in, out)
}
