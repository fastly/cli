package backend

import (
	"io"

	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"

	"github.com/fastly/cli/pkg/text"
)

// CreateCommand wraps the CreateCommand from the servicebackend package.
type CreateCommand struct {
	servicebackend.CreateCommand
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service backend create' command instead.")
	return c.CreateCommand.Exec(in, out)
}
