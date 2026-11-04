package imageoptimizerdefaults

import (
	"io"

	newcmd "github.com/fastly/cli/pkg/commands/service/imageoptimizerdefaults"

	"github.com/fastly/cli/pkg/text"
)

// GetCommand wraps the GetCommand from the newcmd package.
type GetCommand struct {
	newcmd.GetCommand
}

// Exec implements the command interface.
func (c *GetCommand) Exec(in io.Reader, out io.Writer) error {
	text.Deprecated("Use the 'service imageoptimizerdefaults get' command instead.")
	return c.GetCommand.Exec(in, out)
}
