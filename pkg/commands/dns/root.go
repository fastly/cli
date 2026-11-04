package dns

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dns/tsigkey"
	"github.com/fastly/cli/pkg/commands/dns/zone"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	TSIGKey tsigkey.RootCommand `cmd:"" help:"Manipulate Fastly TSIG Keys"`
	Zone    zone.RootCommand    `cmd:"" help:"Manipulate Fastly DNS Zones"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "dns"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
