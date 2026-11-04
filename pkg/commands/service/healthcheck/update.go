package healthcheck

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update healthchecks.
type UpdateCommand struct {
	argparser.Base
	input          fastly.UpdateHealthCheckInput
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Healthcheck name"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	AutoClone     argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	CheckInterval argparser.OptionalInt       `help:"How often to run the healthcheck in milliseconds"`

	Comment          argparser.OptionalString        `help:"A descriptive note"`
	ExpectedResponse argparser.OptionalInt           `help:"The status code expected from the host"`
	Host             argparser.OptionalString        `help:"Which host to check"`
	HTTPVersion      argparser.OptionalString        `help:"Whether to use version 1.0 or 1.1 HTTP"`
	Initial          argparser.OptionalInt           `help:"When loading a config, the initial number of probes to be seen as OK"`
	Method           argparser.OptionalString        `help:"Which HTTP method to use"`
	NewName          argparser.OptionalString        `help:"Healthcheck name"`
	Path             argparser.OptionalString        `help:"The path to check"`
	ServiceID        *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName      argparser.OptionalServiceNameID `help:"The name of the service"`
	Threshold        argparser.OptionalInt           `help:"How many healthchecks must succeed to be considered healthy"`
	Timeout          argparser.OptionalInt           `help:"Timeout in milliseconds"`
	Window           argparser.OptionalInt           `help:"The number of most recent healthcheck queries to keep for this healthcheck"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	if c.NewName.WasSet {
		c.input.NewName = &c.NewName.Value
	}
	if c.Comment.WasSet {
		c.input.Comment = &c.Comment.Value
	}
	if c.Method.WasSet {
		c.input.Method = &c.Method.Value
	}
	if c.Host.WasSet {
		c.input.Host = &c.Host.Value
	}
	if c.Path.WasSet {
		c.input.Path = &c.Path.Value
	}
	if c.HTTPVersion.WasSet {
		c.input.HTTPVersion = &c.HTTPVersion.Value
	}
	if c.Timeout.WasSet {
		c.input.Timeout = &c.Timeout.Value
	}
	if c.CheckInterval.WasSet {
		c.input.CheckInterval = &c.CheckInterval.Value
	}
	if c.ExpectedResponse.WasSet {
		c.input.ExpectedResponse = &c.ExpectedResponse.Value
	}
	if c.Window.WasSet {
		c.input.Window = &c.Window.Value
	}
	if c.Threshold.WasSet {
		c.input.Threshold = &c.Threshold.Value
	}
	if c.Initial.WasSet {
		c.input.Initial = &c.Initial.Value
	}

	h, err := c.Globals.APIClient.UpdateHealthCheck(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Updated healthcheck %s (service %s version %d)",
		fastly.ToValue(h.Name),
		fastly.ToValue(h.ServiceID),
		fastly.ToValue(h.ServiceVersion),
	)
	return nil
}
