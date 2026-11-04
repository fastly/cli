package config

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show a TLS configuration"`
	List     ListCommand     `cmd:"" help:"List all TLS configurations"`
	Update   UpdateCommand   `cmd:"" help:"Update a TLS configuration"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tls-config"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
