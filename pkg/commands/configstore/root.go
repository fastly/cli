package configstore

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// CommandName is the string to be used to invoke this command.
const CommandName = "config-store"

// RootCommand is the parent command for all 'store' subcommands.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create       CreateCommand       `cmd:"" help:"Create a new config store"`
	Delete       DeleteCommand       `cmd:"" help:"Delete a config store"`
	Describe     DescribeCommand     `cmd:"" aliases:"get" help:"Retrieve a single config store"`
	List         ListCommand         `cmd:"" help:"List config stores"`
	ListServices ListServicesCommand `cmd:"" help:"List config store's services"`
	Update       UpdateCommand       `cmd:"" help:"Update a config store"`
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
