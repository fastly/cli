package custom

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" required:"" help:"The name of the VCL to update"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone      argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	NewName        argparser.OptionalString        `help:"New name for the VCL"`
	Content        argparser.OptionalString        `help:"VCL passed as file path or content, e.g. $(< main.vcl)"`
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	v, err := c.Globals.APIClient.UpdateVCL(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if input.NewName != nil && *input.NewName != "" {
		text.Success(out,
			"Updated custom VCL '%s' (previously: '%s', service: %s, version: %d)",
			fastly.ToValue(v.Name),
			input.Name,
			fastly.ToValue(v.ServiceID),
			fastly.ToValue(v.ServiceVersion),
		)
	} else {
		text.Success(out,
			"Updated custom VCL '%s' (service: %s, version: %d)",
			fastly.ToValue(v.Name),
			fastly.ToValue(v.ServiceID),
			fastly.ToValue(v.ServiceVersion),
		)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.UpdateVCLInput, error) {
	var input fastly.UpdateVCLInput

	input.Name = c.NameFlag
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if !c.NewName.WasSet && !c.Content.WasSet {
		return nil, fmt.Errorf("error parsing arguments: must provide either --new-name or --content to update the VCL")
	}
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}
	if c.Content.WasSet {
		input.Content = new(argparser.Content(c.Content.Value))
	}

	return &input, nil
}
