package service

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to create services.
type UpdateCommand struct {
	argparser.Base

	// Optional.
	Comment     argparser.OptionalString `help:"Human-readable comment"`
	input       fastly.UpdateServiceInput
	NameFlag    argparser.OptionalString        `name:"name" short:"n" help:"Service name"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
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
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.ServiceID = serviceID

	if !c.NameFlag.WasSet && !c.Comment.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --name or --comment to update service")
	}

	if c.NameFlag.WasSet {
		c.input.Name = &c.NameFlag.Value
	}
	if c.Comment.WasSet {
		c.input.Comment = &c.Comment.Value
	}

	s, err := c.Globals.APIClient.UpdateService(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   serviceID,
			"Service Name": c.NameFlag.Value,
			"Comment":      c.Comment.Value,
		})
		return err
	}

	text.Success(out, "Updated service %s", fastly.ToValue(s.ServiceID))
	return nil
}
