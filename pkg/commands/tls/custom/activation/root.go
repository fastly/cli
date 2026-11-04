package activation

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Enable   CreateCommand   `cmd:"" aliases:"add" help:"Enable TLS for a particular TLS domain and certificate combination"`
	Disable  DeleteCommand   `cmd:"" aliases:"remove" help:"Disable TLS on the domain associated with this TLS activation"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show a TLS configuration"`
	List     ListCommand     `cmd:"" help:"List all TLS activations"`
	Update   UpdateCommand   `cmd:"" help:"Update the certificate used to terminate TLS traffic for the domain associated with this TLS activation"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "activation"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
