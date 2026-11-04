package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UseCommand switches the default token.
type UseCommand struct {
	argparser.Base
	// Required.
	NameFlag string `arg:"" name:"name" required:"" help:"Name of the token to use as default"`
}

func (c *UseCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := c.Globals.Config.SetDefaultAuthToken(c.NameFlag); err != nil {
		return err
	}

	if err := c.Globals.Config.Write(c.Globals.ConfigPath); err != nil {
		return fmt.Errorf("error saving config: %w", err)
	}

	text.Success(out, "Default token switched to %q", c.NameFlag)
	return nil
}
