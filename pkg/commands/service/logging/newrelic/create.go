package newrelic

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	Key               argparser.OptionalString    `help:"The Insert API key from the Account page of your New Relic account"`
	NameFlag          argparser.OptionalString    `name:"name" help:"The name for the real-time logging configuration"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed"`
	Processingregion  argparser.OptionalString    `name:"processing-region" help:"The region where logs will be processed before streaming to New Relic. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	Region            argparser.OptionalString    `help:"The region where logs are received and stored by New Relic"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint"`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
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

	l, err := c.Globals.APIClient.CreateNewRelic(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(
		out,
		"Created New Relic logging endpoint '%s' (service: %s, version: %d)",
		fastly.ToValue(l.Name),
		fastly.ToValue(l.ServiceID),
		fastly.ToValue(l.ServiceVersion),
	)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateNewRelicInput {
	var input fastly.CreateNewRelicInput

	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.Key.WasSet {
		input.Token = &c.Key.Value
	}

	if c.Format.WasSet {
		input.Format = new(argparser.Content(c.Format.Value))
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
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

	return &input
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *CreateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "autoclone", "format", "format-version", "key",
		"placement", "processing-region", "region", "response-condition",
		"service-id", "service-name",
	}
}
