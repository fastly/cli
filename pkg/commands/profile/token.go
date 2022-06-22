package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

// TokenCommand represents a Kingpin command.
type TokenCommand struct {
	cmd.Base

	clientFactory APIClientFactory
	profile       string
}

// NewTokenCommand returns a new command registered in the parent.
func NewTokenCommand(parent cmd.Registerer, globals *config.Data) *TokenCommand {
	var c TokenCommand
	c.Globals = globals
	c.CmdClause = parent.Command("token", "Print user token")
	c.CmdClause.Flag("user", "Profile user to print token for").Short('u').StringVar(&c.profile)
	return &c
}

// Exec implements the command interface.
func (c *TokenCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.profile == "" {
		if name, p := profile.Default(c.Globals.File.Profiles); name != "" {
			fmt.Println(p.Token)
			return nil
		}
		text.Error(out, "no profiles available")
		return nil
	}

	if name, p := profile.Get(c.profile, c.Globals.File.Profiles); name != "" {
		fmt.Println(p.Token)
		return nil
	}
	text.Error(out, "profile '%s' does not exist", c.profile)
	return nil
}
