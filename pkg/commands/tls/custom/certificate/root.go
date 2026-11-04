package certificate

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create   CreateCommand   `cmd:"" aliases:"add" help:"Create a TLS certificate"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Destroy a TLS certificate. TLS certificates already enabled for a domain cannot be destroyed"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Show a TLS certificate"`
	List     ListCommand     `cmd:"" help:"List all TLS certificates"`
	Update   UpdateCommand   `cmd:"" help:"Replace a TLS certificate with a newly reissued TLS certificate, or update a TLS certificate's name"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "certificate"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
