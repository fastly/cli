package operations

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an operation.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ServiceID   *string                         `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	OperationID string                          `required:"" help:"The unique identifier of the operation"`

	// Optional.
	Description argparser.OptionalString      `help:"Updated description of what the operation does"`
	TagIDs      argparser.OptionalStringSlice `name:"tag-ids" sep:"," help:"Comma-separated list of tag IDs to associate with the operation"`

	argparser.JSONOutput
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if !c.Description.WasSet && !c.TagIDs.WasSet {
		return fmt.Errorf("error parsing arguments: must provide at least one field to update (--description or --tag-ids)")
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	o, err := operations.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   serviceID,
			"Operation ID": c.OperationID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out, "Updated operation %s %s%s (ID: %s)", strings.ToUpper(o.Method), o.Domain, o.Path, o.ID)

	if c.Globals.Verbose() {
		fmt.Fprintln(out)
		if o.Description != "" {
			fmt.Fprintf(out, "Description: %s\n", o.Description)
		}
		if len(o.TagIDs) > 0 {
			fmt.Fprintf(out, "Tags: %d associated\n", len(o.TagIDs))
		}
		if o.UpdatedAt != "" {
			fmt.Fprintf(out, "Updated At: %s\n", o.UpdatedAt)
		}
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string) *operations.UpdateInput {
	input := &operations.UpdateInput{
		ServiceID:   &serviceID,
		OperationID: &c.OperationID,
	}

	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	if c.TagIDs.WasSet {
		input.TagIDs = &c.TagIDs.Value
	}

	return input
}
