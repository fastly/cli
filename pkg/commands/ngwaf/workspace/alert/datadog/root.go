package datadog

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a Datadog alert"`
	Delete DeleteCommand `cmd:"" help:"Delete a Datadog alert"`
	Get    GetCommand    `cmd:"" help:"Get a Datadog alert"`
	List   ListCommand   `cmd:"" help:"List Datadog alerts"`
	Update UpdateCommand `cmd:"" help:"Update a Datadog alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "datadog"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

// ConfigFlags contains Datadog specific configuration flags.
type ConfigFlags struct {
	Key  string
	Site string
}
