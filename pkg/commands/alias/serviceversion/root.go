package serviceversion

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Activate   ActivateCommand   `cmd:"" hidden:"" help:"Activate a Fastly service version"`
	Clone      CloneCommand      `cmd:"" hidden:"" help:"Clone a Fastly service version"`
	Deactivate DeactivateCommand `cmd:"" hidden:"" help:"Deactivate a Fastly service version"`
	List       ListCommand       `cmd:"" hidden:"" help:"List Fastly service versions"`
	Lock       LockCommand       `cmd:"" hidden:"" help:"Lock a Fastly service version"`
	Stage      StageCommand      `cmd:"" hidden:"" help:"Stage a Fastly service version"`
	Unstage    UnstageCommand    `cmd:"" hidden:"" help:"Unstage a Fastly service version"`
	Update     UpdateCommand     `cmd:"" hidden:"" help:"Update a Fastly service version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "service-version"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
