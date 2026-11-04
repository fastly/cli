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
)

// DescribeCommand calls the Fastly API to describe an operation.
type DescribeCommand struct {
	argparser.Base

	// Required.
	ServiceID   *string                         `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	OperationID string                          `required:"" help:"The unique identifier of the operation"`

	// Optional.
	argparser.JSONOutput
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := &operations.DescribeInput{
		ServiceID:   &serviceID,
		OperationID: &c.OperationID,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	o, err := operations.Describe(context.TODO(), fc, input)
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

	return c.print(out, o)
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, o *operations.Operation) error {
	fmt.Fprintf(out, "\nOperation ID: %s\n", o.ID)
	fmt.Fprintf(out, "Method: %s\n", strings.ToUpper(o.Method))
	fmt.Fprintf(out, "Domain: %s\n", o.Domain)
	fmt.Fprintf(out, "Path: %s\n", o.Path)
	fmt.Fprintf(out, "Description: %s\n", o.Description)
	fmt.Fprintf(out, "Status: %s\n", o.Status)
	fmt.Fprintf(out, "Tag IDs: %s\n", strings.Join(o.TagIDs, ", "))
	fmt.Fprintf(out, "RPS: %.2f\n\n", o.RPS)

	if o.CreatedAt != "" {
		fmt.Fprintf(out, "Created At: %s\n", o.CreatedAt)
	}

	if o.UpdatedAt != "" {
		fmt.Fprintf(out, "Updated At: %s\n", o.UpdatedAt)
	}

	if o.LastSeenAt != "" {
		fmt.Fprintf(out, "Last Seen At: %s\n", o.LastSeenAt)
	}

	return nil
}
