package key

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a virtual key"`
	Delete DeleteCommand `cmd:"" aliases:"remove" help:"Delete a virtual key and any associated security configuration"`
	Get    GetCommand    `cmd:"" help:"Get a virtual key"`
	List   ListCommand   `cmd:"" help:"List virtual keys"`
	Rotate RotateCommand `cmd:"" help:"Rotate a virtual key, generating a new access token and invalidating the previous one"`
	Update UpdateCommand `cmd:"" help:"Update a virtual key"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "key"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
