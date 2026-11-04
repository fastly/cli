package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand wraps the DescribeCommand from the servicedictionaryentry package.
type DescribeCommand struct {
	servicedictionaryentry.DescribeCommand
}

// Exec implements the command interface.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service dictionary-entry describe' command instead.")
	}
	return c.DescribeCommand.Exec(in, out)
}
