package webhook

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// CommandName is the string to be used to invoke this command.
const CommandName = "webhook"

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create           CreateCommand           `cmd:"" aliases:"add" help:"Create a Webhook notification integration"`
	Update           UpdateCommand           `cmd:"" help:"Update a Webhook notification integration"`
	GetSigningKey    GetSigningKeyCommand    `cmd:"" help:"Get the signing key for a Webhook notification integration"`
	RotateSigningKey RotateSigningKeyCommand `cmd:"" help:"Rotate the signing key for a Webhook notification integration"`
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
