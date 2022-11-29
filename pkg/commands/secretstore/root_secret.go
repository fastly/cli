package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
)

// RootNameSecret is the base command name for secret operations.
const RootNameSecret = "secret-store-entry"

// NewSecretRootCommand returns a new command registered in the parent.
func NewSecretRootCommand(parent cmd.Registerer, globals *config.Data) *SecretRootCommand {
	c := SecretRootCommand{
		Base: cmd.Base{
			Globals: globals,
		},
	}

	c.CmdClause = parent.Command(RootNameSecret, "Manipulate Fastly Secret Store secrets")

	return &c
}

// SecretRootCommand is the parent command for all 'secret' subcommands.
// It should be installed under the primary root command.
type SecretRootCommand struct {
	cmd.Base
	// no flags
}

// Exec implements the command interface.
func (c *SecretRootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
