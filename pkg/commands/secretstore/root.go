package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootNameStore is the base command name for secret store operations.
const RootNameStore = "secret-store"

// CommandName is the string to be used to invoke this command.
const CommandName = "secret-store"

// RootCommand is the parent command for all 'store' subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" help:"Create a new secret store"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Retrieve a single secret store"`
	Delete   DeleteCommand   `cmd:"" help:"Delete a secret store"`
	List     ListCommand     `cmd:"" help:"List secret stores"`
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
