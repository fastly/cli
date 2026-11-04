package newrelicotlp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required.
	EndpointName   string `name:"name" required:"" help:"The name for the real-time logging configuration to update"`
	Version        string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	AutoClone         argparser.OptionalAutoClone     `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Format            argparser.OptionalString        `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt           `help:"The version of the custom logging format used for the configured endpoint"`
	Key               argparser.OptionalString        `help:"The Insert API key from the Account page of your New Relic account"`
	NewName           argparser.OptionalString        `help:"The name for the real-time logging configuration"`
	Placement         argparser.OptionalString        `help:"Where in the generated VCL the logging call should be placed"`
	Processingregion  argparser.OptionalString        `name:"processing-region" help:"The region where logs will be processed before streaming to New Relic. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	Region            argparser.OptionalString        `help:"The region to which to stream logs"`
	ResponseCondition argparser.OptionalString        `help:"The name of an existing condition in the configured endpoint"`
	URL               argparser.OptionalString        `help:"URL of the New Relic Trace Observer, if you are using New Relic Infinite Tracing"`
	ServiceID         *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName       argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
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

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	l, err := c.Globals.APIClient.UpdateNewRelicOTLP(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	var prev string
	if c.NewName.WasSet {
		prev = fmt.Sprintf("previously: %s, ", c.EndpointName)
	}

	text.Success(
		out,
		"Updated New Relic OTLP logging endpoint '%s' (%sservice: %s, version: %d)",
		fastly.ToValue(l.Name),
		prev,
		fastly.ToValue(l.ServiceID),
		fastly.ToValue(l.ServiceVersion),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) *fastly.UpdateNewRelicOTLPInput {
	var input fastly.UpdateNewRelicOTLPInput

	input.Name = c.EndpointName
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.Format.WasSet {
		input.Format = new(argparser.Content(c.Format.Value))
	}
	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}
	if c.Key.WasSet {
		input.Token = &c.Key.Value
	}
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}
	if c.Placement.WasSet {
		if strings.TrimSpace(c.Placement.Value) == "" {
			input.Placement = fastly.NullValue[string]()
		} else {
			input.Placement = fastly.NewNullable(c.Placement.Value)
		}
	}
	if c.Processingregion.WasSet {
		input.ProcessingRegion = &c.Processingregion.Value
	}
	if c.Region.WasSet {
		input.Region = &c.Region.Value
	}
	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}
	if c.URL.WasSet {
		input.URL = &c.URL.Value
	}

	return &input
}
