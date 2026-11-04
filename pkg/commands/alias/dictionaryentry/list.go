package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/text"
)

// ListCommand wraps the ListCommand from the servicedictionaryentry package.
type ListCommand struct {
	servicedictionaryentry.ListCommand
}

// Exec implements the command interface.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service dictionary-entry list' command instead.")
	}
	return c.ListCommand.Exec(in, out)
}
