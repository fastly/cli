package googlepubsub

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create Google Cloud Pub/Sub logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int
	User         string
	SecretKey    string
	Topic        string
	ProjectID    string

	// optional
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	Placement         common.OptionalString
	ResponseCondition common.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a Google Cloud Pub/Sub logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the Google Cloud Pub/Sub logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)

	c.CmdClause.Flag("user", "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON").Required().StringVar(&c.User)
	c.CmdClause.Flag("secret-key", "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON").Required().StringVar(&c.SecretKey)
	c.CmdClause.Flag("topic", "The Google Cloud Pub/Sub topic to which logs will be published").Required().StringVar(&c.Topic)
	c.CmdClause.Flag("project-id", "The ID of your Google Cloud Platform project").Required().StringVar(&c.ProjectID)

	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput() (*fastly.CreatePubsubInput, error) {
	var input fastly.CreatePubsubInput

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	input.Service = serviceID
	input.Version = c.Version
	input.Name = fastly.String(c.EndpointName)
	input.User = fastly.String(c.User)
	input.SecretKey = fastly.String(c.SecretKey)
	input.Topic = fastly.String(c.Topic)
	input.ProjectID = fastly.String(c.ProjectID)

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.Valid {
		input.Placement = fastly.String(c.Placement.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	d, err := c.Globals.Client.CreatePubsub(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Google Cloud Pub/Sub logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
