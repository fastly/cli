package dictionaryentry

import (
	"io"

	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"

	"github.com/fastly/cli/pkg/text"
)

// CreateCommand wraps the CreateCommand from the servicedictionaryentry package.
type CreateCommand struct {
	servicedictionaryentry.CreateCommand
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service dictionary-entry create' command instead.")
	return c.CreateCommand.Exec(in, out)
}
