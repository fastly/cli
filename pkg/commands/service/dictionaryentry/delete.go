package dictionaryentry

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a service.
type DeleteCommand struct {
	argparser.Base
	Input fastly.DeleteDictionaryItemInput `kong:"-"`

	// Required.
	DictionaryID string `required:"" help:"Dictionary ID"`
	Key          string `required:"" help:"Dictionary item key"`

	// Optional.
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DeleteCommand) AfterApply() error {
	c.Input.DictionaryID = c.DictionaryID
	c.Input.ItemKey = c.Key
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID
	err = c.Globals.APIClient.DeleteDictionaryItem(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Deleted dictionary item %s (service %s, dictionary %s)", c.Input.ItemKey, c.Input.ServiceID, c.Input.DictionaryID)
	return nil
}
