package rule

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create an account-level rule"`
	Delete DeleteCommand `cmd:"" help:"Delete an account-level rule"`
	Get    GetCommand    `cmd:"" help:"Get an account-level rule"`
	List   ListCommand   `cmd:"" help:"List all account-level rules"`
	Update UpdateCommand `cmd:"" help:"Update a workspace"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "rule"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
