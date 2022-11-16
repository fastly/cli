package sftp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create an SFTP logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	Address        string
	User           string
	SSHKnownHosts  string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	Port              cmd.OptionalInt
	Password          cmd.OptionalString
	PublicKey         cmd.OptionalString
	SecretKey         cmd.OptionalString
	Path              cmd.OptionalString
	Period            cmd.OptionalInt
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	GzipLevel         cmd.OptionalInt
	MessageType       cmd.OptionalString
	ResponseCondition cmd.OptionalString
	TimestampFormat   cmd.OptionalString
	Placement         cmd.OptionalString
	CompressionCodec  cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create an SFTP logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the SFTP logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
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
	c.CmdClause.Flag("address", "The hostname or IPv4 address").Required().StringVar(&c.Address)
	c.CmdClause.Flag("user", "The username for the server").Required().StringVar(&c.User)
	c.CmdClause.Flag("ssh-known-hosts", "A list of host keys for all hosts we can connect to over SFTP").Required().StringVar(&c.SSHKnownHosts)
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
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).IntVar(&c.Port.Value)
	c.CmdClause.Flag("password", "The password for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.Password.Set).StringVar(&c.Password.Value)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	c.CmdClause.Flag("secret-key", "The SSH private key for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("path", "The path to upload logs to. The directory must exist on the SFTP server before logs can be saved to it").Action(c.Path.Set).StringVar(&c.Path.Value)
	common.Period(c.CmdClause, &c.Period)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	common.Placement(c.CmdClause, &c.Placement)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateSFTPInput, error) {
	var input fastly.CreateSFTPInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Name = &c.EndpointName
	input.Address = &c.Address
	input.User = &c.User
	input.SSHKnownHosts = &c.SSHKnownHosts

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.Port.WasSet {
		input.Port = &c.Port.Value
	}

	if c.Password.WasSet {
		input.Password = &c.Password.Value
	}

	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
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

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
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

	d, err := c.Globals.APIClient.CreateSFTP(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created SFTP logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
