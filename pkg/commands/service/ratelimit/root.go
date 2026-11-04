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

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a rate limiter for a particular service and version"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Delete a rate limiter by its ID"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Get a rate limiter by its ID"`
	List     ListCommand     `cmd:"" help:"List all rate limiters for a particular service and version"`
	Update   UpdateCommand   `cmd:"" help:"Update a rate limiter by its ID"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "rate-limit"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
