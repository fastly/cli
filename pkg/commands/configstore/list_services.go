package configstore

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListServicesCommand) AfterApply() error {
	c.input.StoreID = c.StoreID
	return nil
}

// ListServicesCommand calls the Fastly API to list appropriate resources.
type ListServicesCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	argparser.JSONOutput
	input fastly.ListConfigStoreServicesInput
}

// Exec invokes the application logic for the command.
func (c *ListServicesCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.ListConfigStoreServices(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.PrintConfigStoreServicesTbl(out, o)

	return nil
}
