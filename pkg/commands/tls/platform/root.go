package platform

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Upload   CreateCommand   `cmd:"" help:"Upload a new certificate"`
	Delete   DeleteCommand   `cmd:"" aliases:"remove" help:"Destroy a certificate. This disables TLS for all domains listed as SAN entries"`
	Describe DescribeCommand `cmd:"" aliases:"get" help:"Retrieve a single certificate"`
	List     ListCommand     `cmd:"" help:"List all certificates"`
	Update   UpdateCommand   `cmd:"" help:"Replace a certificate with a newly reissued certificate"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tls-platform"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
