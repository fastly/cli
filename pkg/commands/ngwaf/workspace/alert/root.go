package alert

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// CommandName is the string to be used to invoke this command.
const CommandName = "alert"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

// GetDefaultEvents returns the hardcoded events value for all alerts.
// Currently the only supported value is "flag".
func GetDefaultEvents() *[]string {
	events := []string{"flag"}
	return &events
}
