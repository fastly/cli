package slack

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a Slack alert"`
	Delete DeleteCommand `cmd:"" help:"Delete a Slack alert"`
	Get    GetCommand    `cmd:"" help:"Get a Slack alert"`
	List   ListCommand   `cmd:"" help:"List Slack alerts"`
	Update UpdateCommand `cmd:"" help:"Update a Slack alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "slack"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
