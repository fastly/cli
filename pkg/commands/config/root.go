package config

import (
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	location bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("config", "Display the Fastly CLI configuration")
	c.CmdClause.Flag("location", "Print the location of the CLI configuration file").Short('l').BoolVar(&c.location)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.location {
		fmt.Println(config.FilePath)
		return nil
	}

	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	fmt.Println(string(data))
	return nil
}
