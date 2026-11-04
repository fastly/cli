package signallist

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create an account-level signal list"`
	Delete DeleteCommand `cmd:"" help:"Delete an account signal list"`
	Get    GetCommand    `cmd:"" help:"Get an account-level signal list"`
	List   ListCommand   `cmd:"" help:"List all signal lists for your account"`
	Update UpdateCommand `cmd:"" help:"Update an account-level signal list"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "signal-list"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
