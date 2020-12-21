package googlepubsub

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update a Google Cloud Pub/Sub logging endpoint.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shadow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	User              common.OptionalString
	SecretKey         common.OptionalString
	ProjectID         common.OptionalString
	Topic             common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	Placement         common.OptionalString
	ResponseCondition common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Google Cloud Pub/Sub logging endpoint on a Fastly service version")

	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Google Cloud Pub/Sub logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("new-name", "New name of the Google Cloud Pub/Sub logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("user", "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("secret-key", "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("topic", "The Google Cloud Pub/Sub topic to which logs will be published").Action(c.Topic.Set).StringVar(&c.Topic.Value)
	c.CmdClause.Flag("project-id", "The ID of your Google Cloud Platform project").Action(c.ProjectID.Set).StringVar(&c.ProjectID.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdatePubsubInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	googlepubsub, err := c.Globals.Client.GetPubsub(&fastly.GetPubsubInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdatePubsubInput{
		ServiceID:         googlepubsub.ServiceID,
		ServiceVersion:    googlepubsub.ServiceVersion,
		Name:              googlepubsub.Name,
		NewName:           fastly.String(googlepubsub.Name),
		User:              fastly.String(googlepubsub.User),
		SecretKey:         fastly.String(googlepubsub.SecretKey),
		ProjectID:         fastly.String(googlepubsub.ProjectID),
		Topic:             fastly.String(googlepubsub.Topic),
		Format:            fastly.String(googlepubsub.Format),
		FormatVersion:     fastly.Uint(googlepubsub.FormatVersion),
		Placement:         fastly.String(googlepubsub.Placement),
		ResponseCondition: fastly.String(googlepubsub.ResponseCondition),
	}

	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.User.WasSet {
		input.User = fastly.String(c.User.Value)
	}

	if c.SecretKey.WasSet {
		input.SecretKey = fastly.String(c.SecretKey.Value)
	}

	if c.Topic.WasSet {
		input.Topic = fastly.String(c.Topic.Value)
	}

	if c.ProjectID.WasSet {
		input.ProjectID = fastly.String(c.ProjectID.Value)
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

	googlepubsub, err := c.Globals.Client.UpdatePubsub(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Google Cloud Pub/Sub logging endpoint %s (service %s version %d)", googlepubsub.Name, googlepubsub.ServiceID, googlepubsub.ServiceVersion)
	return nil
}
