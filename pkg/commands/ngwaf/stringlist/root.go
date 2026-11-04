package stringlist

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create an account-level string list"`
	Delete DeleteCommand `cmd:"" help:"Delete an account string list"`
	Get    GetCommand    `cmd:"" help:"Get an account-level string list"`
	List   ListCommand   `cmd:"" help:"List all string lists for your account"`
	Update UpdateCommand `cmd:"" help:"Update an account-level string list"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "string-list"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
