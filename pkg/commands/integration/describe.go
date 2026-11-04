package integration

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a notification integration.
type DescribeCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput

	Input fastly.GetIntegrationInput `kong:"-"`

	// Required.
	ID string `arg:"" required:"" help:"Integration ID"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.ID = c.ID
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetIntegration(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.PrintIntegration(out, o)

	return nil
}
