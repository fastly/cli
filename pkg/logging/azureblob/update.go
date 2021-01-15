package azureblob

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update an Azure Blob Storage logging endpoint.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	//required
	EndpointName string
	Version      int

	// optional
	NewName           common.OptionalString
	AccountName       common.OptionalString
	Container         common.OptionalString
	SASToken          common.OptionalString
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

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update an Azure Blob Storage logging endpoint on a Fastly service version")

	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("new-name", "New name of the Azure Blob Storage logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("container", "The name of the Azure Blob Storage container in which to store logs").Action(c.Container.Set).StringVar(&c.Container.Value)
	c.CmdClause.Flag("account-name", "The unique Azure Blob Storage namespace in which your data objects are stored").Action(c.AccountName.Set).StringVar(&c.AccountName.Value)
	c.CmdClause.Flag("sas-token", "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work").Action(c.SASToken.Set).StringVar(&c.SASToken.Value)
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
func (c *UpdateCommand) createInput() (*fastly.UpdateBlobStorageInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	input := fastly.UpdateBlobStorageInput{
		ServiceID:      serviceID,
		ServiceVersion: c.Version,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Path.WasSet {
		input.Path = fastly.String(c.Path.Value)
	}

	if c.AccountName.WasSet {
		input.AccountName = fastly.String(c.AccountName.Value)
	}

	if c.Container.WasSet {
		input.Container = fastly.String(c.Container.Value)
	}

	if c.SASToken.WasSet {
		input.SASToken = fastly.String(c.SASToken.Value)
	}

	if c.Period.WasSet {
		input.Period = fastly.Uint(c.Period.Value)
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = fastly.Uint(c.GzipLevel.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.MessageType.WasSet {
		input.MessageType = fastly.String(c.MessageType.Value)
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = fastly.String(c.TimestampFormat.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.PublicKey.WasSet {
		input.PublicKey = fastly.String(c.PublicKey.Value)
	}

	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = fastly.Uint(c.FileMaxBytes.Value)
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = fastly.String(c.CompressionCodec.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	azureblob, err := c.Globals.Client.UpdateBlobStorage(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Azure Blob Storage logging endpoint %s (service %s version %d)", azureblob.Name, azureblob.ServiceID, azureblob.ServiceVersion)
	return nil
}
