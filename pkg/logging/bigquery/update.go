package bigquery

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update BigQuery logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	ProjectID         common.OptionalString
	Dataset           common.OptionalString
	Table             common.OptionalString
	User              common.OptionalString
	SecretKey         common.OptionalString
	Template          common.OptionalString
	Placement         common.OptionalString
	ResponseCondition common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a BigQuery logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the BigQuery logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the BigQuery logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("project-id", "Your Google Cloud Platform project ID").Action(c.ProjectID.Set).StringVar(&c.ProjectID.Value)
	c.CmdClause.Flag("dataset", "Your BigQuery dataset").Action(c.Dataset.Set).StringVar(&c.Dataset.Value)
	c.CmdClause.Flag("table", "Your BigQuery table").Action(c.Table.Set).StringVar(&c.Table.Value)
	c.CmdClause.Flag("user", "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON.").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("secret-key", "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON.").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("template-suffix", "BigQuery table name suffix template").Action(c.Template.Set).StringVar(&c.Template.Value)
	c.CmdClause.Flag("format", "Apache style log formatting. Must produce JSON that matches the schema of your BigQuery table").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateBigQueryInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	bq, err := c.Globals.Client.GetBigQuery(&fastly.GetBigQueryInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateBigQueryInput{
		ServiceID:         bq.ServiceID,
		ServiceVersion:    bq.ServiceVersion,
		Name:              bq.Name,
		NewName:           fastly.String(bq.Name),
		ProjectID:         fastly.String(bq.ProjectID),
		Dataset:           fastly.String(bq.Dataset),
		Table:             fastly.String(bq.Table),
		Template:          fastly.String(bq.Template),
		User:              fastly.String(bq.User),
		SecretKey:         fastly.String(bq.SecretKey),
		Format:            fastly.String(bq.Format),
		ResponseCondition: fastly.String(bq.ResponseCondition),
		Placement:         fastly.String(bq.Placement),
		FormatVersion:     fastly.Uint(bq.FormatVersion),
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.ProjectID.WasSet {
		input.ProjectID = fastly.String(c.ProjectID.Value)
	}

	if c.Dataset.WasSet {
		input.Dataset = fastly.String(c.Dataset.Value)
	}

	if c.Table.WasSet {
		input.Table = fastly.String(c.Table.Value)
	}

	if c.User.WasSet {
		input.User = fastly.String(c.User.Value)
	}

	if c.SecretKey.WasSet {
		input.SecretKey = fastly.String(c.SecretKey.Value)
	}

	if c.Template.WasSet {
		input.Template = fastly.String(c.Template.Value)
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

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	bq, err := c.Globals.Client.UpdateBigQuery(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated BigQuery logging endpoint %s (service %s version %d)", bq.Name, bq.ServiceID, bq.ServiceVersion)
	return nil
}
