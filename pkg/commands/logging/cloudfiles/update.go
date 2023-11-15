package cloudfiles

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a Cloudfiles logging endpoint.
type UpdateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// Required.
	EndpointName   string // Can't shadow cmd.Base method Name().
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// Optional.
	AutoClone         cmd.OptionalAutoClone
	NewName           cmd.OptionalString
	User              cmd.OptionalString
	AccessKey         cmd.OptionalString
	BucketName        cmd.OptionalString
	Path              cmd.OptionalString
	Region            cmd.OptionalString
	Placement         cmd.OptionalString
	Period            cmd.OptionalInt
	GzipLevel         cmd.OptionalInt
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	ResponseCondition cmd.OptionalString
	MessageType       cmd.OptionalString
	TimestampFormat   cmd.OptionalString
	PublicKey         cmd.OptionalString
	CompressionCodec  cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Cloudfiles logging endpoint on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "The name of the Cloudfiles logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("access-key", "Your Cloudfile account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("bucket", "The name of your Cloudfiles container").Action(c.BucketName.Set).StringVar(&c.BucketName.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("new-name", "New name of the Cloudfiles logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.Placement(c.CmdClause, &c.Placement)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	c.CmdClause.Flag("region", "The region to stream logs to. One of: DFW-Dallas, ORD-Chicago, IAD-Northern Virginia, LON-London, SYD-Sydney, HKG-Hong Kong").Action(c.Region.Set).StringVar(&c.Region.Value)
	c.CmdClause.Flag("user", "The username for your Cloudfile account").Action(c.User.Set).StringVar(&c.User.Value)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateCloudfilesInput, error) {
	input := fastly.UpdateCloudfilesInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}

	if c.User.WasSet {
		input.User = &c.User.Value
	}

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.BucketName.WasSet {
		input.BucketName = &c.BucketName.Value
	}

	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}

	if c.Region.WasSet {
		input.Region = &c.Region.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
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
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
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

	cloudfiles, err := c.Globals.APIClient.UpdateCloudfiles(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Updated Cloudfiles logging endpoint %s (service %s version %d)", cloudfiles.Name, cloudfiles.ServiceID, cloudfiles.ServiceVersion)
	return nil
}
