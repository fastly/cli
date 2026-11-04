package ratelimit

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/ratelimit"

	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand wraps the UpdateCommand from the newcmd package.
type UpdateCommand struct {
	newcmd.UpdateCommand
}

// Exec implements the command interface.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	if !c.JSONOutput.Enabled {
		text.Deprecated("Use the 'service rate-limit update' command instead.")
	}
	return c.UpdateCommand.Exec(in, out)
}
