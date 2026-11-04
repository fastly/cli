package ratelimit

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" hidden:"" help:"Create a rate limiter for a particular service and version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" hidden:"" help:"Delete a rate limiter by its ID"`
	Describe DescribeCommand `cmd:"" aliases:"get" hidden:"" help:"Get a rate limiter by its ID"`
	List     ListCommand     `cmd:"" hidden:"" help:"List all rate limiters for a particular service and version"`
	Update   UpdateCommand   `cmd:"" hidden:"" help:"Update a rate limiter by its ID"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "rate-limit"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
