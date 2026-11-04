package accesskeys

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" help:"Create an access key"`
	Delete DeleteCommand `cmd:"" help:"Delete an access key"`
	Get    GetCommand    `cmd:"" help:"Get an access key"`
	List   ListCommand   `cmd:"" aliases:"list-access-keys" help:"List all access keys"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "access-keys"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
