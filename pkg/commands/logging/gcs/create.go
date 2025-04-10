package gcs

import (
	"fmt"
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

// CreateCommand calls the Fastly API to create a GCS logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// Optional.
	AccountName       argparser.OptionalString
	AutoClone         argparser.OptionalAutoClone
	Bucket            argparser.OptionalString
	CompressionCodec  argparser.OptionalString
	EndpointName      argparser.OptionalString // Can't shadow argparser.Base method Name().
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	GzipLevel         argparser.OptionalInt
	MessageType       argparser.OptionalString
	Path              argparser.OptionalString
	Period            argparser.OptionalInt
	Placement         argparser.OptionalString
	ProjectID         argparser.OptionalString
	ResponseCondition argparser.OptionalString
	SecretKey         argparser.OptionalString
	TimestampFormat   argparser.OptionalString
	User              argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a GCS logging endpoint on a Fastly service version").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the GCS logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	common.AccountName(c.CmdClause, &c.AccountName)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("bucket", "The bucket of the GCS bucket").Action(c.Bucket.Set).StringVar(&c.Bucket.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.Placement(c.CmdClause, &c.Placement)
	c.CmdClause.Flag("project-id", "The google project ID").Action(c.ProjectID.Set).StringVar(&c.ProjectID.Value)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("secret-key", "Your GCS account secret key. The private_key field in your service account authentication JSON").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
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
	c.CmdClause.Flag("user", "Your GCS service account email address. The client_email field in your service account authentication JSON").Action(c.User.Set).StringVar(&c.User.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateGCSInput, error) {
	input := fastly.CreateGCSInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.AccountName.WasSet {
		input.AccountName = &c.AccountName.Value
	}
	if c.Bucket.WasSet {
		input.Bucket = &c.Bucket.Value
	}
	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
	}
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.Format.WasSet {
		input.Format = fastly.ToPointer(argparser.Content(c.Format.Value))
	}
	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}
	if c.GzipLevel.WasSet {
		input.GzipLevel = &c.GzipLevel.Value
	}
	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}
	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}
	if c.Period.WasSet {
		input.Period = &c.Period.Value
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
	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}
	if c.User.WasSet {
		input.User = &c.User.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	d, err := c.Globals.APIClient.CreateGCS(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Created GCS logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}
