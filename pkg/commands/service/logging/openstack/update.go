package openstack

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/logflags"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an OpenStack logging endpoint.
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
	TimestampFormat   argparser.OptionalString
	URL               argparser.OptionalString
	User              argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update an OpenStack logging endpoint on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "The name of the OpenStack logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("access-key", "Your OpenStack account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("bucket", "The name of the Openstack Space").Action(c.BucketName.Set).StringVar(&c.BucketName.Value)
	logflags.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	logflags.Format(c.CmdClause, &c.Format)
	logflags.FormatVersion(c.CmdClause, &c.FormatVersion)
	logflags.GzipLevel(c.CmdClause, &c.GzipLevel)
	logflags.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("new-name", "New name of the OpenStack logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	logflags.Path(c.CmdClause, &c.Path)
	logflags.Period(c.CmdClause, &c.Period)
	logflags.Placement(c.CmdClause, &c.Placement)
	logflags.ProcessingRegion(c.CmdClause, &c.ProcessingRegion, "OpenStack")
	logflags.PublicKey(c.CmdClause, &c.PublicKey)
	logflags.ResponseCondition(c.CmdClause, &c.ResponseCondition)
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
	logflags.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	c.CmdClause.Flag("url", "Your OpenStack auth url.").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("user", "The username for your OpenStack account.").Action(c.User.Set).StringVar(&c.User.Value)

	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateOpenstackInput, error) {
	input := fastly.UpdateOpenstackInput{
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

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.User.WasSet {
		input.User = &c.User.Value
	}

	if c.URL.WasSet {
		input.URL = &c.URL.Value
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

	openstack, err := c.Globals.APIClient.UpdateOpenstack(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Updated OpenStack logging endpoint %s (service %s version %d)",
		fastly.ToValue(openstack.Name),
		fastly.ToValue(openstack.ServiceID),
		fastly.ToValue(openstack.ServiceVersion),
	)
	return nil
}
