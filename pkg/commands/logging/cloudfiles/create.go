package cloudfiles

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// CreateCommand calls the Fastly API to create a Cloudfiles logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	Token          string
	User           string
	AccessKey      string
	BucketName     string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	Path              cmd.OptionalString
	Region            cmd.OptionalString
	PublicKey         cmd.OptionalString
	Period            cmd.OptionalUint
	GzipLevel         cmd.OptionalUint8
	MessageType       cmd.OptionalString
	TimestampFormat   cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalUint
	ResponseCondition cmd.OptionalString
	Placement         cmd.OptionalString
	CompressionCodec  cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create a Cloudfiles logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the Cloudfiles logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
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
	c.CmdClause.Flag("user", "The username for your Cloudfile account").Required().StringVar(&c.User)
	c.CmdClause.Flag("access-key", "Your Cloudfile account access key").Required().StringVar(&c.AccessKey)
	c.CmdClause.Flag("bucket", "The name of your Cloudfiles container").Required().StringVar(&c.BucketName)
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
	common.Path(c.CmdClause, &c.Path)
	c.CmdClause.Flag("region", "The region to stream logs to. One of: DFW-Dallas, ORD-Chicago, IAD-Northern Virginia, LON-London, SYD-Sydney, HKG-Hong Kong").Action(c.Region.Set).StringVar(&c.Region.Value)
	common.Placement(c.CmdClause, &c.Placement)
	common.Period(c.CmdClause, &c.Period)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateCloudfilesInput, error) {
	var input fastly.CreateCloudfilesInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Name = c.EndpointName
	input.User = c.User
	input.AccessKey = c.AccessKey
	input.BucketName = c.BucketName

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.Path.WasSet {
		input.Path = c.Path.Value
	}

	if c.Region.WasSet {
		input.Region = c.Region.Value
	}

	if c.Placement.WasSet {
		input.Placement = c.Placement.Value
	}

	if c.Period.WasSet {
		input.Period = c.Period.Value
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = c.GzipLevel.Value
	}

	if c.Format.WasSet {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = c.MessageType.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = c.TimestampFormat.Value
	}

	if c.PublicKey.WasSet {
		input.PublicKey = c.PublicKey.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = c.CompressionCodec.Value
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

	d, err := c.Globals.APIClient.CreateCloudfiles(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created Cloudfiles logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
