package azureblob

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Azure Blob Storage logging endpoints.
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
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update an Azure Blob Storage logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object").Short('n').Required().StringVar(&c.EndpointName)

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

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateBlobStorageInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	azureblob, err := c.Globals.Client.GetBlobStorage(&fastly.GetBlobStorageInput{
		Service: serviceID,
		Name:    c.EndpointName,
		Version: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateBlobStorageInput{
		Service:           azureblob.ServiceID,
		Version:           azureblob.Version,
		Name:              azureblob.Name,
		NewName:           azureblob.Name,
		Path:              azureblob.Path,
		AccountName:       azureblob.AccountName,
		Container:         azureblob.Container,
		SASToken:          azureblob.SASToken,
		Period:            azureblob.Period,
		TimestampFormat:   azureblob.TimestampFormat,
		GzipLevel:         azureblob.GzipLevel,
		PublicKey:         azureblob.PublicKey,
		Format:            azureblob.Format,
		FormatVersion:     azureblob.FormatVersion,
		MessageType:       azureblob.MessageType,
		Placement:         azureblob.Placement,
		ResponseCondition: azureblob.ResponseCondition,
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.Path.Valid {
		input.Path = c.Path.Value
	}

	if c.AccountName.Valid {
		input.AccountName = c.AccountName.Value
	}

	if c.Container.Valid {
		input.Container = c.Container.Value
	}

	if c.SASToken.Valid {
		input.SASToken = c.SASToken.Value
	}

	if c.Period.Valid {
		input.Period = c.Period.Value
	}

	if c.GzipLevel.Valid {
		input.GzipLevel = c.GzipLevel.Value
	}

	if c.Format.Valid {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.MessageType.Valid {
		input.MessageType = c.MessageType.Value
	}

	if c.TimestampFormat.Valid {
		input.TimestampFormat = c.TimestampFormat.Value
	}

	if c.Placement.Valid {
		input.Placement = c.Placement.Value
	}

	if c.PublicKey.Valid {
		input.PublicKey = c.PublicKey.Value
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

	text.Success(out, "Updated Azure Blob Storage logging endpoint %s (service %s version %d)", azureblob.Name, azureblob.ServiceID, azureblob.Version)
	return nil
}
