package bigquery

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/logging/bigquery"

	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand wraps the UpdateCommand from the newcmd package.
type UpdateCommand struct {
	newcmd.UpdateCommand
}

// Exec implements the command interface.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service logging bigquery update' command instead.")
	return c.UpdateCommand.Exec(in, out)
}
