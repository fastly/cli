package openstack

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an OpenStack logging endpoint.
type DeleteCommand struct {
	argparser.Base
	Input          fastly.DeleteOpenstackInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"The name of the OpenStack logging object"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone   argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DeleteCommand) AfterApply() error {
	c.Input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	if err := c.Globals.APIClient.DeleteOpenstack(context.TODO(), &c.Input); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Deleted OpenStack logging endpoint %s (service %s version %d)", c.Input.Name, c.Input.ServiceID, c.Input.ServiceVersion)
	return nil
}
