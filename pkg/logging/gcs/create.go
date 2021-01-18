package gcs

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// CreateCommand calls the Fastly API to create a GCS logging endpoint.
type CreateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shadow common.Base method Name().
	Version      int
	Bucket       string
	User         string
	SecretKey    string

	// optional
	Path              common.OptionalString
	Period            common.OptionalUint
	GzipLevel         common.OptionalUint8
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	MessageType       common.OptionalString
	ResponseCondition common.OptionalString
	TimestampFormat   common.OptionalString
	Placement         common.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a GCS logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the GCS logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("user", "Your GCS service account email address. The client_email field in your service account authentication JSON").Required().StringVar(&c.User)
	c.CmdClause.Flag("bucket", "The bucket of the GCS bucket").Required().StringVar(&c.Bucket)
	c.CmdClause.Flag("secret-key", "Your GCS account secret key. The private_key field in your service account authentication JSON").Required().StringVar(&c.SecretKey)

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("path", "The path to upload logs to (default '/')").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).Uint8Var(&c.GzipLevel.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput() (*fastly.CreateGCSInput, error) {
	var input fastly.CreateGCSInput

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	input.ServiceID = serviceID
	input.ServiceVersion = c.Version
	input.Name = c.EndpointName
	input.Bucket = c.Bucket
	input.User = c.User
	input.SecretKey = c.SecretKey

	if c.Path.WasSet {
		input.Path = c.Path.Value
	}

	if c.Period.WasSet {
		input.Period = c.Period.Value
	}

	if c.Format.WasSet {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = c.GzipLevel.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = c.TimestampFormat.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = c.MessageType.Value
	}

	if c.Placement.WasSet {
		input.Placement = c.Placement.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	d, err := c.Globals.Client.CreateGCS(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created GCS logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
