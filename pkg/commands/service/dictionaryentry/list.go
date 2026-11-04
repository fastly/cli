package dictionaryentry

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list dictionary items.
type ListCommand struct {
	argparser.Base

	// Required.
	DictionaryID string `required:"" help:"Dictionary ID"`

	// Optional.
	Direction string `default:"ascend" enum:"ascend,descend" hints:"ascend,descend" help:"Direction in which to sort results"`
	argparser.JSONOutput

	input       fastly.GetDictionaryItemsInput
	Page        int                             `help:"Page number of data set to fetch"`
	PerPage     int                             `help:"Number of records per page"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Sort        string                          `default:"created" help:"Field on which to sort"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.input.DictionaryID = c.DictionaryID
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.Direction = &c.Direction
	c.input.Page = &c.Page
	c.input.PerPage = &c.PerPage
	c.input.ServiceID = serviceID
	c.input.Sort = &c.Sort
	paginator := c.Globals.APIClient.GetDictionaryItems(context.TODO(), &c.input)

	var o []*fastly.DictionaryItem
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Dictionary ID":   c.input.DictionaryID,
				"Service ID":      serviceID,
				"Remaining Pages": paginator.Remaining(),
			})
			return err
		}
		o = append(o, data...)
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", c.input.ServiceID)
	}
	for i, dictionary := range o {
		text.Output(out, "Item: %d/%d", i+1, len(o))
		text.PrintDictionaryItem(out, "\t", dictionary)
		text.Break(out)
	}

	return nil
}
