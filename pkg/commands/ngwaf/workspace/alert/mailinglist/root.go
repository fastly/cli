package mailinglist

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a Mailing List alert"`
	Delete DeleteCommand `cmd:"" help:"Delete a Mailing List alert"`
	Get    GetCommand    `cmd:"" help:"Get a Mailing List alert"`
	List   ListCommand   `cmd:"" help:"List Mailing List alerts"`
	Update UpdateCommand `cmd:"" help:"Update a Mailing List alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "mailinglist"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

// AddressConfigFlags contains Address configurations used by mailing lists.
type AddressConfigFlags struct {
	Address string
}
