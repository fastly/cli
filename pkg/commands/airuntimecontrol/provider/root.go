package provider

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	List       ListCommand       `cmd:"" help:"List AI providers supported by AI Runtime Control"`
	ListModels ListModelsCommand `cmd:"" help:"List models available for a specific provider"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "provider"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
