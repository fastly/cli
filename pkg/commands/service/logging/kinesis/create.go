package kinesis

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an Amazon Kinesis logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data `kong:"-"`

	// Required.
	ServiceName    argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion argparser.OptionalServiceVersion `kong:"-"`

	// required, but mutually exclusive

	// mutual exclusions
	// AccessKey + SecretKey or IAMRole must be provided
	AccessKey argparser.OptionalString `help:"The access key associated with the target Amazon Kinesis stream"`
	SecretKey argparser.OptionalString `help:"The secret key associated with the target Amazon Kinesis stream"`
	IAMRole   argparser.OptionalString `help:"The IAM role ARN for logging"`

	// Optional.
	AutoClone         argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
	EndpointName      argparser.OptionalString    `name:"name" short:"n" help:"The name of the Kinesis logging object. Used as a primary key for API access"` // Can't shadow argparser.Base method Name().
	Format            argparser.OptionalString    `help:"Apache style log formatting. Your log must produce valid JSON. Can be a string or a file path to a file containing formatting"`
	FormatVersion     argparser.OptionalInt       `help:"The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1"`
	Placement         argparser.OptionalString    `help:"Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none, which suppresses the logging call (the only valid value for Compute services). This field is not required and has no default value. When updating a Delivery service, pass an empty string to reset the endpoint back to automatic placement"`
	ProcessingRegion  argparser.OptionalString    `help:"The region where logs will be processed before streaming to Kinesis. One of 'none', 'eu', or 'us'. Defaults to 'none' for no specific region"`
	Region            argparser.OptionalString    `help:"The region where logs are received and stored by Kinesis"`
	ResponseCondition argparser.OptionalString    `help:"The name of an existing condition in the configured endpoint, or leave blank to always execute"`
	StreamName        argparser.OptionalString    `help:"The Amazon Kinesis stream to send logs to"`

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.ServiceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateKinesisInput, error) {
	var input fastly.CreateKinesisInput

	input.ServiceID = serviceID
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.StreamName.WasSet {
		input.StreamName = &c.StreamName.Value
	}
	if c.Region.WasSet {
		input.Region = &c.Region.Value
	}
	input.ServiceVersion = serviceVersion

	// The following block checks for invalid permutations of the ways in
	// which the AccessKey + SecretKey and IAMRole flags can be
	// provided. This is necessary because either the AccessKey and
	// SecretKey or the IAMRole is required, but they are mutually
	// exclusive. The kingpin library lacks a way to express this constraint
	// via the flag specification API so we enforce it manually here.
	switch {
	case !c.AccessKey.WasSet && !c.SecretKey.WasSet && !c.IAMRole.WasSet:
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags or the --iam-role flag must be provided")
	case (c.AccessKey.WasSet || c.SecretKey.WasSet) && c.IAMRole.WasSet:
		// Enforce mutual exclusion
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag")
	case c.AccessKey.WasSet && !c.SecretKey.WasSet:
		return nil, fmt.Errorf("error parsing arguments: required flag --secret-key not provided")
	case !c.AccessKey.WasSet && c.SecretKey.WasSet:
		return nil, fmt.Errorf("error parsing arguments: required flag --access-key not provided")
	}

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}

	if c.IAMRole.WasSet {
		input.IAMRole = &c.IAMRole.Value
	}

	if c.Format.WasSet {
		input.Format = new(argparser.Content(c.Format.Value))
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}

	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}

	return &input, nil
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
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	d, err := c.Globals.APIClient.CreateKinesis(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(
		out,
		"Created Kinesis logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *CreateCommand) FlagOrder() []string {
	return []string{
		"name", "version", "access-key", "secret-key", "iam-role", "autoclone",
		"format", "format-version", "region", "response-condition",
		"placement", "processing-region", "stream-name", "service-id",
		"service-name",
	}
}
