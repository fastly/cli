package acl

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/acl"

	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand wraps the DeleteCommand from the newcmd package.
type DeleteCommand struct {
	newcmd.DeleteCommand
}

// Exec implements the command interface.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service acl delete' command instead.")
	return c.DeleteCommand.Exec(in, out)
}
