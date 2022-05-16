package config

import (
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	filePath string
	location bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("config", "Display the Fastly CLI configuration")
	c.CmdClause.Flag("location", "Print the location of the CLI configuration file").Short('l').BoolVar(&c.location)
	c.filePath = globals.Path
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.location {
		if c.Globals.Flag.Verbose {
			text.Break(out)
		}
		fmt.Fprintln(out, c.filePath)
		return nil
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	fmt.Fprintln(out, string(data))
	return nil
}
