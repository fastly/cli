package secretstore

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
func (c *DeleteCommand) AfterApply() error {
	c.Input.StoreID = c.StoreID
	return nil
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	argparser.JSONOutput

	Input fastly.DeleteSecretStoreInput `kong:"-"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	err := c.Globals.APIClient.DeleteSecretStore(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.Input.StoreID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted Secret Store '%s'", c.Input.StoreID)
	return nil
}
