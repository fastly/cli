package digitalocean

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

// UpdateCommand calls the Fastly API to update a DigitalOcean Spaces logging endpoint.
type UpdateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	NewName           cmd.OptionalString
	BucketName        cmd.OptionalString
	Domain            cmd.OptionalString
	AccessKey         cmd.OptionalString
	SecretKey         cmd.OptionalString
	Path              cmd.OptionalString
	Period            cmd.OptionalInt
	GzipLevel         cmd.OptionalInt
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	ResponseCondition cmd.OptionalString
	MessageType       cmd.OptionalString
	TimestampFormat   cmd.OptionalString
	Placement         cmd.OptionalString
	PublicKey         cmd.OptionalString
	CompressionCodec  cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("update", "Update a DigitalOcean Spaces logging endpoint on a Fastly service version")
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
	c.CmdClause.Flag("name", "The name of the DigitalOcean Spaces logging object").Short('n').Required().StringVar(&c.EndpointName)
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
	c.CmdClause.Flag("new-name", "New name of the DigitalOcean Spaces logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("bucket", "The name of the DigitalOcean Space").Action(c.BucketName.Set).StringVar(&c.BucketName.Value)
	c.CmdClause.Flag("domain", "The domain of the DigitalOcean Spaces endpoint (default 'nyc3.digitaloceanspaces.com')").Action(c.Domain.Set).StringVar(&c.Domain.Value)
	c.CmdClause.Flag("access-key", "Your DigitalOcean Spaces account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("secret-key", "Your DigitalOcean Spaces account secret key").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	common.Placement(c.CmdClause, &c.Placement)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateDigitalOceanInput, error) {
	input := fastly.UpdateDigitalOceanInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.BucketName.WasSet {
		input.BucketName = &c.BucketName.Value
	}

	if c.Domain.WasSet {
		input.Domain = &c.Domain.Value
	}

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}

	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}

	if c.Period.WasSet {
		input.Period = &c.Period.Value
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = &c.GzipLevel.Value
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

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}

	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
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

	digitalocean, err := c.Globals.APIClient.UpdateDigitalOcean(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated DigitalOcean Spaces logging endpoint %s (service %s version %d)", digitalocean.Name, digitalocean.ServiceID, digitalocean.ServiceVersion)
	return nil
}
