package azureblob

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// CreateCommand calls the Fastly API to create an Azure Blob Storage logging endpoint.
type CreateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName   string // Can't shadow common.Base method Name().
	Container      string
	AccountName    string
	SASToken       string
	serviceVersion common.OptionalServiceVersion

	// optional
	autoClone         common.OptionalAutoClone
	Path              common.OptionalString
	Period            common.OptionalUint
	GzipLevel         common.OptionalUint
	MessageType       common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	ResponseCondition common.OptionalString
	TimestampFormat   common.OptionalString
	Placement         common.OptionalString
	PublicKey         common.OptionalString
	FileMaxBytes      common.OptionalUint
	CompressionCodec  common.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an Azure Blob Storage logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.NewServiceVersionFlag(common.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	c.NewAutoCloneFlag(c.autoClone.Set, &c.autoClone.Value)
	c.CmdClause.Flag("container", "The name of the Azure Blob Storage container in which to store logs").Required().StringVar(&c.Container)
	c.CmdClause.Flag("account-name", "The unique Azure Blob Storage namespace in which your data objects are stored").Required().StringVar(&c.AccountName)
	c.CmdClause.Flag("sas-token", "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work").Required().StringVar(&c.SASToken)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("path", "The path to upload logs to").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).UintVar(&c.GzipLevel.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.PublicKey.Set).StringVar(&c.PublicKey.Value)
	c.CmdClause.Flag("file-max-bytes", "The maximum size of a log file in bytes").Action(c.FileMaxBytes.Set).UintVar(&c.FileMaxBytes.Value)
	c.CmdClause.Flag("compression-codec", `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`).Action(c.CompressionCodec.Set).StringVar(&c.CompressionCodec.Value)
	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput() (*fastly.CreateBlobStorageInput, error) {
	var input fastly.CreateBlobStorageInput

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	v, err := c.serviceVersion.Parse(serviceID, c.Globals.Client)
	if err != nil {
		return nil, err
	}
	v, err = c.autoClone.Parse(v, serviceID, c.Globals.Client)
	if err != nil {
		return nil, err
	}

	input.ServiceID = serviceID
	input.ServiceVersion = v.Number
	input.Name = c.EndpointName
	input.Container = c.Container
	input.AccountName = c.AccountName
	input.SASToken = c.SASToken

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.Path.WasSet {
		input.Path = c.Path.Value
	}

	if c.Period.WasSet {
		input.Period = c.Period.Value
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = c.GzipLevel.Value
	}

	if c.Format.WasSet {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = c.MessageType.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = c.TimestampFormat.Value
	}

	if c.Placement.WasSet {
		input.Placement = c.Placement.Value
	}

	if c.PublicKey.WasSet {
		input.PublicKey = c.PublicKey.Value
	}

	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = c.FileMaxBytes.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = c.CompressionCodec.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	d, err := c.Globals.Client.CreateBlobStorage(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Azure Blob Storage logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
