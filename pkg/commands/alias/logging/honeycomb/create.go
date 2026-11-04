package honeycomb

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/honeycomb"

	"github.com/fastly/cli/pkg/text"
)

// CreateCommand wraps the CreateCommand from the newcmd package.
type CreateCommand struct {
	newcmd.CreateCommand
}

// Exec implements the command interface.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service logging honeycomb create' command instead.")
	return c.CreateCommand.Exec(in, out)
}
