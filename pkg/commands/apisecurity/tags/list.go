package tags

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/apisecurity/operations"
)

// ListCommand calls the Fastly API to list all operation tags.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	ServiceID   *string                         `help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
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

	if serviceID == "" {
		return errors.New("service-id is required")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}
	// Auto-paginate through all results
	var allTags []operations.OperationTag
	page := 0
	limit := 100

	input := &operations.ListTagsInput{
		ServiceID: &serviceID,
	}

	for {
		input.Page = &page
		input.Limit = &limit

		tags, err := operations.ListTags(context.TODO(), fc, input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
				"Page":       page,
			})
			return err
		}

		if tags == nil || len(tags.Data) == 0 {
			break
		}

		allTags = append(allTags, tags.Data...)

		// Check if we've fetched all results
		if len(allTags) >= tags.Meta.Total {
			break
		}

		page++
	}

	if ok, err := c.WriteJSON(out, allTags); ok {
		return err
	}

	text.PrintOperationTagsTbl(out, allTags)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *ListCommand) FlagOrder() []string {
	return []string{
		"service-id", "service-name", "json",
	}
}
