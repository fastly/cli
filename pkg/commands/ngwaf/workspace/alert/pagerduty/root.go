package pagerduty

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" aliases:"add" help:"Create a PagerDuty alert"`
	Delete DeleteCommand `cmd:"" help:"Delete a PagerDuty alert"`
	Get    GetCommand    `cmd:"" help:"Get a PagerDuty alert"`
	List   ListCommand   `cmd:"" help:"List PagerDuty alerts"`
	Update UpdateCommand `cmd:"" help:"Update a PagerDuty alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "pagerduty"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
