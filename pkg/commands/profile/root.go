package profile

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Create CreateCommand `cmd:"" help:"Create user profile (deprecated: use 'fastly auth login' or 'fastly auth add' instead)"`
	Delete DeleteCommand `cmd:"" help:"Delete user profile (deprecated: use 'fastly auth delete' instead)"`
	List   ListCommand   `cmd:"" help:"List user profiles (deprecated: use 'fastly auth list' instead)"`
	Switch SwitchCommand `cmd:"" help:"Switch user profile (deprecated: use 'fastly auth use' instead)"`
	Token  TokenCommand  `cmd:"" help:"Print API token (deprecated: use 'fastly auth show' instead)"`
	Update UpdateCommand `cmd:"" help:"Update user profile (deprecated: use 'fastly auth login' or 'fastly auth add' instead)"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "profile"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
