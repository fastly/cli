package webhook

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create           CreateCommand           `cmd:"" aliases:"add" help:"Create a Webhook alert"`
	Delete           DeleteCommand           `cmd:"" help:"Delete a Webhook alert"`
	Get              GetCommand              `cmd:"" help:"Get a Webhook alert"`
	GetSigningKey    GetSigningKeyCommand    `cmd:"" help:"Retrieves details of a webhook alert signing key"`
	List             ListCommand             `cmd:"" help:"List Webhook alerts"`
	RotateSigningKey RotateSigningKeyCommand `cmd:"" help:"Rotate webhook alert signing key"`
	Update           UpdateCommand           `cmd:"" help:"Update a Webhook alert"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "webhook"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
