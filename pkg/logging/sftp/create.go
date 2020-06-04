package sftp

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create SFTP logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName  string // Can't shaddow common.Base method Name().
	Version       int
	Address       string
	User          string
	SSHKnownHosts string

	// optional
	Port              common.OptionalUint
	Password          common.OptionalString
	PublicKey         common.OptionalString
	SecretKey         common.OptionalString
	Path              common.OptionalString
	Period            common.OptionalUint
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	GzipLevel         common.OptionalUint
	MessageType       common.OptionalString
	ResponseCondition common.OptionalString
	TimestampFormat   common.OptionalString
	Placement         common.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand

	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an SFTP logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the SFTP logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)

	c.CmdClause.Flag("address", "The hostname or IPv4 addres").Required().StringVar(&c.Address)
	c.CmdClause.Flag("user", "The username for the server").Required().StringVar(&c.User)
	c.CmdClause.Flag("ssh-known-hosts", "A list of host keys for all hosts we can connect to over SFTP").Required().StringVar(&c.SSHKnownHosts)

	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("password", "The password for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.Password.Set).StringVar(&c.Password.Value)
	c.CmdClause.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.PublicKey.Set).StringVar(&c.PublicKey.Value)
	c.CmdClause.Flag("secret-key", "The SSH private key for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("path", "The path to upload logs to. The directory must exist on the SFTP server before logs can be saved to it").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).UintVar(&c.GzipLevel.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput() (*fastly.CreateSFTPInput, error) {
	var input fastly.CreateSFTPInput

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	input.Service = serviceID
	input.Version = c.Version
	input.Name = fastly.String(c.EndpointName)
	input.Address = fastly.String(c.Address)
	input.User = fastly.String(c.User)
	input.SSHKnownHosts = fastly.String(c.SSHKnownHosts)

	if c.Port.Valid {
		input.Port = fastly.Uint(c.Port.Value)
	}

	if c.Password.Valid {
		input.Password = fastly.String(c.Password.Value)
	}

	if c.PublicKey.Valid {
		input.PublicKey = fastly.String(c.PublicKey.Value)
	}

	if c.SecretKey.Valid {
		input.SecretKey = fastly.String(c.SecretKey.Value)
	}

	if c.Path.Valid {
		input.Path = fastly.String(c.Path.Value)
	}

	if c.Period.Valid {
		input.Period = fastly.Uint(c.Period.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.GzipLevel.Valid {
		input.GzipLevel = fastly.Uint(c.GzipLevel.Value)
	}

	if c.MessageType.Valid {
		input.MessageType = fastly.String(c.MessageType.Value)
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.TimestampFormat.Valid {
		input.TimestampFormat = fastly.String(c.TimestampFormat.Value)
	}

	if c.Placement.Valid {
		input.Placement = fastly.String(c.Placement.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	d, err := c.Globals.Client.CreateSFTP(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created SFTP logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
