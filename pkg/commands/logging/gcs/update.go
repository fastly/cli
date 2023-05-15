package gcs

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

// UpdateCommand calls the Fastly API to update a GCS logging endpoint.
type UpdateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// Required.
	EndpointName   string // Can't shadow cmd.Base method Name().
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// Optional.
	AccountName       cmd.OptionalString
	AutoClone         cmd.OptionalAutoClone
	Bucket            cmd.OptionalString
	CompressionCodec  cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	GzipLevel         cmd.OptionalInt
	MessageType       cmd.OptionalString
	NewName           cmd.OptionalString
	Path              cmd.OptionalString
	Period            cmd.OptionalInt
	Placement         cmd.OptionalString
	ResponseCondition cmd.OptionalString
	SecretKey         cmd.OptionalString
	TimestampFormat   cmd.OptionalString
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
	c.CmdClause = parent.Command("update", "Update a GCS logging endpoint on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "The name of the GCS logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	common.AccountName(c.CmdClause, &c.AccountName)
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("bucket", "The bucket of the GCS bucket").Action(c.Bucket.Set).StringVar(&c.Bucket.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("new-name", "New name of the GCS logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("path", "The path to upload logs to (default '/')").Action(c.Path.Set).StringVar(&c.Path.Value)
	common.Period(c.CmdClause, &c.Period)
	common.Placement(c.CmdClause, &c.Placement)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("secret-key", "Your GCS account secret key. The private_key field in your service account authentication JSON").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
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
	c.CmdClause.Flag("user", "Your GCS service account email address. The client_email field in your service account authentication JSON").Action(c.User.Set).StringVar(&c.User.Value)
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateGCSInput, error) {
	input := fastly.UpdateGCSInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
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
	if c.Format.WasSet {
		input.Format = &c.Format.Value
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
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
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

	gcs, err := c.Globals.APIClient.UpdateGCS(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated GCS logging endpoint %s (service %s version %d)", gcs.Name, gcs.ServiceID, gcs.ServiceVersion)
	return nil
}
