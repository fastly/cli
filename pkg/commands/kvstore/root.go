package kvstore

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" help:"Create a KV Store"`
	Delete   DeleteCommand   `cmd:"" help:"Delete a KV Store"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Describe a KV Store"`
	List     ListCommand     `cmd:"" help:"List KV Stores"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "kv-store"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
