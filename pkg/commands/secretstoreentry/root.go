package secretstoreentry

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootNameSecret is the base command name for secret operations.
const RootNameSecret = "secret-store-entry"

// CommandName is the string to be used to invoke this command.
const CommandName = "secret-store-entry"

// RootCommand is the parent command for all 'secret' subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" help:"Create a new secret within specified store"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Retrieve a single secret"`
	Delete   DeleteCommand   `cmd:"" help:"Delete a secret"`
	List     ListCommand     `cmd:"" help:"List secrets within a specified store"`
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
