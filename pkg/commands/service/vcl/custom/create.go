package custom

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	Content argparser.OptionalString `help:"VCL passed as file path or content, e.g. $(< main.vcl)"`
	Version string                   `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone      argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Main           argparser.OptionalBool          `help:"Whether the VCL is the 'main' entrypoint"`
	NameFlag       argparser.OptionalString        `name:"name" help:"The name of the VCL"`
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	v, err := c.Globals.APIClient.CreateVCL(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created custom VCL '%s' (service: %s, version: %d, main: %t)",
		fastly.ToValue(v.Name),
		fastly.ToValue(v.ServiceID),
		fastly.ToValue(v.ServiceVersion),
		fastly.ToValue(v.Main),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateVCLInput {
	input := fastly.CreateVCLInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Content.WasSet {
		input.Content = new(argparser.Content(c.Content.Value))
	}
	if c.Main.WasSet {
		input.Main = new(c.Main.Value)
	}
	return &input
}
