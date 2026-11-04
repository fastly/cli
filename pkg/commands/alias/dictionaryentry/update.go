package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand wraps the UpdateCommand from the servicedictionaryentry package.
type UpdateCommand struct {
	servicedictionaryentry.UpdateCommand
}

// Exec implements the command interface.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service dictionary-entry update' command instead.")
	return c.UpdateCommand.Exec(in, out)
}
