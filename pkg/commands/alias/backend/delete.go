package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the servicebackend package.
type DeleteCommand struct {
	servicebackend.DeleteCommand
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service backend delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
