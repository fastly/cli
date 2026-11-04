package privatekey

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a TLS private key"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Destroy a TLS private key. Only private keys not already matched to any certificates can be deleted"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show a TLS private key"`
	List     ListCommand     `cmd:"" help:"List all TLS private keys"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "private-key"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
