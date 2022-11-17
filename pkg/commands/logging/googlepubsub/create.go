package googlepubsub

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

// CreateCommand calls the Fastly API to create a Google Cloud Pub/Sub logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	EndpointName      cmd.OptionalString // Can't shadow cmd.Base method Name().
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	Placement         cmd.OptionalString
	ProjectID         cmd.OptionalString
	ResponseCondition cmd.OptionalString
	SecretKey         cmd.OptionalString
	Topic             cmd.OptionalString
	User              cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create a Google Cloud Pub/Sub logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the Google Cloud Pub/Sub logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
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
	c.CmdClause.Flag("user", "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("secret-key", "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("topic", "The Google Cloud Pub/Sub topic to which logs will be published").Action(c.Topic.Set).StringVar(&c.Topic.Value)
	c.CmdClause.Flag("project-id", "The ID of your Google Cloud Platform project").Action(c.ProjectID.Set).StringVar(&c.ProjectID.Value)
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
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.Placement(c.CmdClause, &c.Placement)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreatePubsubInput, error) {
	var input fastly.CreatePubsubInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.User.WasSet {
		input.User = &c.User.Value
	}
	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}
	if c.Topic.WasSet {
		input.Topic = &c.Topic.Value
	}
	if c.ProjectID.WasSet {
		input.ProjectID = &c.ProjectID.Value
	}

	if c.Format.WasSet {
		input.Format = &c.Format.Value
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
		c.Globals.ErrLog.Add(err)
		return err
	}

	d, err := c.Globals.APIClient.CreatePubsub(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created Google Cloud Pub/Sub logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
