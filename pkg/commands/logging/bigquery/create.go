package bigquery

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create a BigQuery logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	ProjectID      string
	Dataset        string
	Table          string
	User           string
	SecretKey      string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	Template          cmd.OptionalString
	Placement         cmd.OptionalString
	ResponseCondition cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create a BigQuery logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the BigQuery logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("project-id", "Your Google Cloud Platform project ID").Required().StringVar(&c.ProjectID)
	c.CmdClause.Flag("dataset", "Your BigQuery dataset").Required().StringVar(&c.Dataset)
	c.CmdClause.Flag("table", "Your BigQuery table").Required().StringVar(&c.Table)
	c.CmdClause.Flag("user", "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON.").Required().StringVar(&c.User)
	c.CmdClause.Flag("secret-key", "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON.").Required().StringVar(&c.SecretKey)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	c.CmdClause.Flag("template-suffix", "BigQuery table name suffix template").Action(c.Template.Set).StringVar(&c.Template.Value)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.Placement(c.CmdClause, &c.Placement)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateBigQueryInput, error) {
	var input fastly.CreateBigQueryInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Name = &c.EndpointName
	input.ProjectID = &c.ProjectID
	input.Dataset = &c.Dataset
	input.User = &c.User
	input.Table = &c.Table
	input.SecretKey = &c.SecretKey

	if c.Template.WasSet {
		input.Template = &c.Template.Value
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Int(c.FormatVersion.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	d, err := c.Globals.APIClient.CreateBigQuery(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created BigQuery logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
