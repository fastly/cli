package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the servicedictionaryentry package.
type DeleteCommand struct {
	servicedictionaryentry.DeleteCommand
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service dictionary-entry delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
