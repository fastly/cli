package resourcelink

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a resource link.
type CreateCommand struct {
	argparser.Base

	input          fastly.CreateResourceInput
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	ResourceID string `short:"r" required:"" help:"Resource ID"`
	Version    string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// At least one of the following is required.
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`

	// Optional.
	AutoClone argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	argparser.JSONOutput
	NameFlag string `name:"name" short:"n" help:"Resource link name (alias). Defaults to resource's name"`
}

// BeforeApply sets the command up before kong applies the command line.
func (c *CreateCommand) BeforeApply() error {
	c.input = fastly.CreateResourceInput{
		// Kingpin requires the following to be initialized.
		ResourceID: new(string),
		Name:       new(string),
	}
	return nil
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.input.ResourceID = &c.ResourceID
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.input.Name = &c.NameFlag
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

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
			"Service ID":      c.Globals.Manifest.Flag.ServiceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.CreateResource(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"ID":              c.input.ResourceID,
			"Service ID":      c.input.ServiceID,
			"Service Version": c.input.ServiceVersion,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.Success(out,
		"Created service resource link %q (%s) on service %s version %d",
		fastly.ToValue(o.Name),
		fastly.ToValue(o.LinkID),
		fastly.ToValue(o.ServiceID),
		fastly.ToValue(o.ServiceVersion),
	)
	return nil
}
