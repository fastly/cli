package authtoken

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create an API token (deprecated: use the Fastly API directly)"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Revoke an API token (deprecated: use the Fastly API directly)"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Get the current API token (deprecated: use the Fastly API directly)"`
	List     ListCommand     `cmd:"" help:"List API tokens (deprecated: use the Fastly API directly)"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "auth-token"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
