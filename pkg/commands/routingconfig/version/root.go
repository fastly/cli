package version

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	List           ListCommand           `cmd:"" help:"List a routing config's inactive versions"`
	Activate       ActivateCommand       `cmd:"" help:"Reactivate a previous version of a routing config"`
	DeleteInactive DeleteInactiveCommand `cmd:"" help:"Permanently delete all inactive versions of a routing config. This destroys rollback history and cannot be undone"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "version"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
