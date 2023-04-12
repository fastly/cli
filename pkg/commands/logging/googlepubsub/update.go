package googlepubsub

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// UpdateCommand calls the Fastly API to update a Google Cloud Pub/Sub logging endpoint.
type UpdateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AccountName       cmd.OptionalString
	AutoClone         cmd.OptionalAutoClone
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	NewName           cmd.OptionalString
	Placement         cmd.OptionalString
	ProjectID         cmd.OptionalString
	ResponseCondition cmd.OptionalString
	SecretKey         cmd.OptionalString
	Topic             cmd.OptionalString
	User              cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		Manifest: m,
	}
	c.CmdClause = parent.Command("update", "Update a Google Cloud Pub/Sub logging endpoint on a Fastly service version")

	// required
	c.CmdClause.Flag("name", "The name of the Google Cloud Pub/Sub logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// optional
	common.AccountName(c.CmdClause, &c.AccountName)
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	c.CmdClause.Flag("new-name", "New name of the Google Cloud Pub/Sub logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	common.Placement(c.CmdClause, &c.Placement)
	c.CmdClause.Flag("project-id", "The ID of your Google Cloud Platform project").Action(c.ProjectID.Set).StringVar(&c.ProjectID.Value)
	c.CmdClause.Flag("secret-key", "Your Google Cloud Platform account secret key. The private_key field in your service account authentication JSON").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
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
	c.CmdClause.Flag("topic", "The Google Cloud Pub/Sub topic to which logs will be published").Action(c.Topic.Set).StringVar(&c.Topic.Value)
	c.CmdClause.Flag("user", "Your Google Cloud Platform service account email address. The client_email field in your service account authentication JSON").Action(c.User.Set).StringVar(&c.User.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdatePubsubInput, error) {
	input := fastly.UpdatePubsubInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	if c.AccountName.WasSet {
		input.AccountName = &c.AccountName.Value
	}
	if c.Format.WasSet {
		input.Format = &c.Format.Value
	}
	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}
	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}
	if c.ProjectID.WasSet {
		input.ProjectID = &c.ProjectID.Value
	}
	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}
	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}
	if c.Topic.WasSet {
		input.Topic = &c.Topic.Value
	}
	if c.User.WasSet {
		input.User = &c.User.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.Manifest,
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

	input, err := c.ConstructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	googlepubsub, err := c.Globals.APIClient.UpdatePubsub(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated Google Cloud Pub/Sub logging endpoint %s (service %s version %d)", googlepubsub.Name, googlepubsub.ServiceID, googlepubsub.ServiceVersion)
	return nil
}
