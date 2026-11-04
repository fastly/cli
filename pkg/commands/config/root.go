package config

import (
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base

	Location bool `short:"l" help:"Print the location of the CLI configuration file"`
	Reset    bool `short:"r" help:"Reset the config to a version compatible with the current CLI version"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "config"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	if c.Reset {
		if err := c.Globals.Config.UseStatic(config.FilePath); err != nil {
			return err
		}
	}

	if c.Location {
		if c.Globals.Flags.Verbose {
			text.Break(out)
		}
		fmt.Fprintln(out, c.Globals.ConfigPath)
		return nil
	}

	data, err := os.ReadFile(c.Globals.ConfigPath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	fmt.Fprintln(out, string(data))
	return nil
}
