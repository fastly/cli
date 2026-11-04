package subscription

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a new TLS subscription"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Destroy a TLS subscription. A subscription cannot be destroyed if there are domains in the TLS enabled state"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show a TLS subscription"`
	List     ListCommand     `cmd:"" help:"List all TLS subscriptions"`
	Update   UpdateCommand   `cmd:"" help:"Change the TLS domains or common name associated with this subscription, or update the TLS configuration for this set of domains"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tls-subscription"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
