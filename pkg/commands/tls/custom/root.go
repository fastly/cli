package custom

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/tls/custom/activation"
	"github.com/fastly/cli/pkg/commands/tls/custom/certificate"
	"github.com/fastly/cli/pkg/commands/tls/custom/domain"
	"github.com/fastly/cli/pkg/commands/tls/custom/privatekey"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Activation  activation.RootCommand  `cmd:"" help:"Upload and manage TLS activations"`
	Certificate certificate.RootCommand `cmd:"" help:"Upload and manage TLS certificates"`
	Domain      domain.RootCommand      `cmd:"" help:"Manage TLS domains"`
	PrivateKey  privatekey.RootCommand  `cmd:"" help:"Upload and manage private keys used to sign certificates"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "tls-custom"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
