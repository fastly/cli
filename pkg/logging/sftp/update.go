package sftp

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update SFTP logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string
	Version      int

	// optional
	NewName           common.OptionalString
	Address           common.OptionalString
	Port              common.OptionalUint
	PublicKey         common.OptionalString
	SecretKey         common.OptionalString
	SSHKnownHosts     common.OptionalString
	User              common.OptionalString
	Password          common.OptionalString
	Path              common.OptionalString
	Period            common.OptionalUint
	FormatVersion     common.OptionalUint
	GzipLevel         common.OptionalUint
	Format            common.OptionalString
	MessageType       common.OptionalString
	ResponseCondition common.OptionalString
	TimestampFormat   common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update an SFTP logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the SFTP logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the SFTP logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("address", "The hostname or IPv4 address").Action(c.Address.Set).StringVar(&c.Address.Value)
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.PublicKey.Set).StringVar(&c.PublicKey.Value)
	c.CmdClause.Flag("secret-key", "The SSH private key for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("ssh-known-hosts", "A list of host keys for all hosts we can connect to over SFTP").Action(c.SSHKnownHosts.Set).StringVar(&c.SSHKnownHosts.Value)
	c.CmdClause.Flag("user", "The username for the server").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("password", "The password for the server. If both password and secret_key are passed, secret_key will be used in preference").Action(c.Password.Set).StringVar(&c.Password.Value)
	c.CmdClause.Flag("path", "The path to upload logs to. The directory must exist on the SFTP server before logs can be saved to it").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).UintVar(&c.GzipLevel.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateSFTPInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	sftp, err := c.Globals.Client.GetSFTP(&fastly.GetSFTPInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateSFTPInput{
		ServiceID:         sftp.ServiceID,
		ServiceVersion:    sftp.ServiceVersion,
		Name:              sftp.Name,
		NewName:           fastly.String(sftp.Name),
		Address:           fastly.String(sftp.Address),
		Port:              fastly.Uint(sftp.Port),
		PublicKey:         fastly.String(sftp.PublicKey),
		SecretKey:         fastly.String(sftp.SecretKey),
		SSHKnownHosts:     fastly.String(sftp.SSHKnownHosts),
		User:              fastly.String(sftp.User),
		Password:          fastly.String(sftp.Password),
		Path:              fastly.String(sftp.Path),
		Period:            fastly.Uint(sftp.Period),
		Format:            fastly.String(sftp.Format),
		FormatVersion:     fastly.Uint(sftp.FormatVersion),
		GzipLevel:         fastly.Uint(uint(sftp.GzipLevel)), // TODO (v2): consistent type.
		MessageType:       fastly.String(sftp.MessageType),
		ResponseCondition: fastly.String(sftp.ResponseCondition),
		TimestampFormat:   fastly.String(sftp.TimestampFormat),
		Placement:         fastly.String(sftp.Placement),
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
		input.GzipLevel = fastly.Uint(c.GzipLevel.Value)
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

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	sftp, err := c.Globals.Client.UpdateSFTP(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated SFTP logging endpoint %s (service %s version %d)", sftp.Name, sftp.ServiceID, sftp.ServiceVersion)
	return nil
}
