package objectstorage

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/objectstorage/accesskeys"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	AccessKeys accesskeys.RootCommand `cmd:"" help:"Manipulate Fastly access keys"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "object-storage"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
