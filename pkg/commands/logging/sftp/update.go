package sftp

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// UpdateCommand calls the Fastly API to update an SFTP logging endpoint.
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
	Address           cmd.OptionalString
	Port              cmd.OptionalUint
	PublicKey         cmd.OptionalString
	SecretKey         cmd.OptionalString
	SSHKnownHosts     cmd.OptionalString
	User              cmd.OptionalString
	Password          cmd.OptionalString
	Path              cmd.OptionalString
	Period            cmd.OptionalUint
	FormatVersion     cmd.OptionalUint
	GzipLevel         cmd.OptionalUint8
	Format            cmd.OptionalString
	MessageType       cmd.OptionalString
	ResponseCondition cmd.OptionalString
	TimestampFormat   cmd.OptionalString
	Placement         cmd.OptionalString
	CompressionCodec  cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("update", "Update an SFTP logging endpoint on a Fastly service version")
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
	c.CmdClause.Flag("name", "The name of the SFTP logging object").Short('n').Required().StringVar(&c.EndpointName)
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
	c.CmdClause.Flag("new-name", "New name of the SFTP logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("address", "The hostname or IPv4 address").Action(c.Address.Set).StringVar(&c.Address.Value)
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	c.CmdClause.Flag("secret-key", "The SSH private key for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("ssh-known-hosts", "A list of host keys for all hosts we can connect to over SFTP").Action(c.SSHKnownHosts.Set).StringVar(&c.SSHKnownHosts.Value)
	c.CmdClause.Flag("user", "The username for the server").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("password", "The password for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.Password.Set).StringVar(&c.Password.Value)
	c.CmdClause.Flag("path", "The path to upload logs to. The directory must exist on the SFTP server before logs can be saved to it").Action(c.Path.Set).StringVar(&c.Path.Value)
	common.Period(c.CmdClause, &c.Period)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	common.Placement(c.CmdClause, &c.Placement)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateSFTPInput, error) {
	input := fastly.UpdateSFTPInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Address.WasSet {
		input.Address = fastly.String(c.Address.Value)
	}

	if c.Port.WasSet {
		input.Port = fastly.Uint(c.Port.Value)
	}

	if c.Password.WasSet {
		input.Password = fastly.String(c.Password.Value)
	}

	if c.PublicKey.WasSet {
		input.PublicKey = fastly.String(c.PublicKey.Value)
	}

	if c.SecretKey.WasSet {
		input.SecretKey = fastly.String(c.SecretKey.Value)
	}

	if c.SSHKnownHosts.WasSet {
		input.SSHKnownHosts = fastly.String(c.SSHKnownHosts.Value)
	}

	if c.User.WasSet {
		input.User = fastly.String(c.User.Value)
	}

	if c.Path.WasSet {
		input.Path = fastly.String(c.Path.Value)
	}

	if c.Period.WasSet {
		input.Period = fastly.Uint(c.Period.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.GzipLevel.WasSet {
		input.GzipLevel = fastly.Uint8(c.GzipLevel.Value)
	}

	if c.MessageType.WasSet {
		input.MessageType = fastly.String(c.MessageType.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = fastly.String(c.TimestampFormat.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = fastly.String(c.CompressionCodec.Value)
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

	sftp, err := c.Globals.APIClient.UpdateSFTP(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated SFTP logging endpoint %s (service %s version %d)", sftp.Name, sftp.ServiceID, sftp.ServiceVersion)
	return nil
}
