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

	Activate   ActivateCommand   `cmd:"" help:"Activate a Fastly service version"`
	Clone      CloneCommand      `cmd:"" help:"Clone a Fastly service version"`
	Deactivate DeactivateCommand `cmd:"" help:"Deactivate a Fastly service version"`
	List       ListCommand       `cmd:"" help:"List Fastly service versions"`
	Lock       LockCommand       `cmd:"" help:"Lock a Fastly service version"`
	Stage      StageCommand      `cmd:"" help:"Stage a Fastly service version"`
	Unstage    UnstageCommand    `cmd:"" help:"Unstage a Fastly service version"`
	Update     UpdateCommand     `cmd:"" help:"Update a Fastly service version"`
	Validate   ValidateCommand   `cmd:"" help:"Validate a service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "version"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
