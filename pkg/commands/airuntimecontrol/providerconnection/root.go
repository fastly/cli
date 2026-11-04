package providerconnection

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a provider connection"`
	Delete DeleteCommand `cmd:"" aliases:"remove" help:"Delete a provider connection. This is rejected if the connection is still referenced by a virtual key's failover policy"`
	Get    GetCommand    `cmd:"" help:"Get a provider connection"`
	List   ListCommand   `cmd:"" help:"List provider connections"`
	Update UpdateCommand `cmd:"" help:"Update a provider connection"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "provider-connection"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
