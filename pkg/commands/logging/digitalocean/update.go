package digitalocean

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a DigitalOcean Spaces logging endpoint.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	EndpointName   string
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// Optional.
	AccessKey         argparser.OptionalString
	AutoClone         argparser.OptionalAutoClone
	BucketName        argparser.OptionalString
	CompressionCodec  argparser.OptionalString
	Domain            argparser.OptionalString
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	GzipLevel         argparser.OptionalInt
	MessageType       argparser.OptionalString
	NewName           argparser.OptionalString
	Path              argparser.OptionalString
	Period            argparser.OptionalInt
	Placement         argparser.OptionalString
	ProcessingRegion  argparser.OptionalString
	PublicKey         argparser.OptionalString
	ResponseCondition argparser.OptionalString
	SecretKey         argparser.OptionalString
	TimestampFormat   argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a DigitalOcean Spaces logging endpoint on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "The name of the DigitalOcean Spaces logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("access-key", "Your DigitalOcean Spaces account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("bucket", "The name of the DigitalOcean Space").Action(c.BucketName.Set).StringVar(&c.BucketName.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	c.CmdClause.Flag("domain", "The domain of the DigitalOcean Spaces endpoint (default 'nyc3.digitaloceanspaces.com')").Action(c.Domain.Set).StringVar(&c.Domain.Value)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("new-name", "New name of the DigitalOcean Spaces logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.Placement(c.CmdClause, &c.Placement)
	common.ProcessingRegion(c.CmdClause, &c.ProcessingRegion, "DigitalOcean Spaces")
	common.PublicKey(c.CmdClause, &c.PublicKey)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("secret-key", "Your DigitalOcean Spaces account secret key").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.ServiceName.Value,
	})
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
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
		input.Format = fastly.ToPointer(argparser.Content(c.Format.Value))
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

	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	digitalocean, err := c.Globals.APIClient.UpdateDigitalOcean(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Updated DigitalOcean Spaces logging endpoint %s (service %s version %d)",
		fastly.ToValue(digitalocean.Name),
		fastly.ToValue(digitalocean.ServiceID),
		fastly.ToValue(digitalocean.ServiceVersion),
	)
	return nil
}
