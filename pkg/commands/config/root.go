package config

import (
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base

	location bool
	reset    bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("config", "Display the Fastly CLI configuration")
	c.CmdClause.Flag("location", "Print the location of the CLI configuration file").Short('l').BoolVar(&c.location)
	c.CmdClause.Flag("reset", "Reset the config to a version compatible with the current CLI version").Short('r').BoolVar(&c.reset)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	if c.reset {
		if err := c.Globals.Config.UseStatic(config.FilePath); err != nil {
			return err
		}
	}

	if c.location {
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
