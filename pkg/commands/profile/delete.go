package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand represents a Kingpin command.
type DeleteCommand struct {
	argparser.Base

	Profile string `arg:"" required:"" help:"Profile to delete"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if !c.Globals.Flags.Quiet {
		text.Deprecated("This command will be removed in a future release. Use 'fastly auth delete' instead.\n\n")
	}

	if !c.Globals.Config.DeleteAuthToken(c.Profile) {
		return fmt.Errorf("the specified profile does not exist")
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return err
	}
	if c.Globals.Verbose() {
		text.Break(out)
	}
	text.Success(out, "Profile '%s' deleted", c.Profile)

	if c.Globals.Config.Auth.Default == "" && len(c.Globals.Config.Auth.Tokens) > 0 {
		text.Break(out)
		text.Warning(out, "At least one account profile should be set as the 'default'. Run `fastly profile update <NAME>` and ensure the profile is set to be the default.")
	}
	return nil
}
