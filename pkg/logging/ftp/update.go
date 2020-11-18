package ftp

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update FTP logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	Address           common.OptionalString
	Port              common.OptionalUint
	Username          common.OptionalString
	Password          common.OptionalString
	PublicKey         common.OptionalString
	Path              common.OptionalString
	Period            common.OptionalUint
	GzipLevel         common.OptionalUint8
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	ResponseCondition common.OptionalString
	TimestampFormat   common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update an FTP logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the FTP logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the FTP logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("address", "An hostname or IPv4 address").Action(c.Address.Set).StringVar(&c.Address.Value)
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("username", "The username for the server (can be anonymous)").Action(c.Username.Set).StringVar(&c.Username.Value)
	c.CmdClause.Flag("password", "The password for the server (for anonymous use an email address)").Action(c.Password.Set).StringVar(&c.Password.Value)
	c.CmdClause.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.PublicKey.Set).StringVar(&c.PublicKey.Value)
	c.CmdClause.Flag("path", "The path to upload log files to. If the path ends in / then it is treated as a directory").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Period.Set).UintVar(&c.Period.Value)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.GzipLevel.Set).Uint8Var(&c.GzipLevel.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.TimestampFormat.Set).StringVar(&c.TimestampFormat.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateFTPInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	ftp, err := c.Globals.Client.GetFTP(&fastly.GetFTPInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateFTPInput{
		ServiceID:         ftp.ServiceID,
		ServiceVersion:    ftp.ServiceVersion,
		Name:              ftp.Name,
		NewName:           fastly.String(ftp.Name),
		Address:           fastly.String(ftp.Address),
		Port:              fastly.Uint(ftp.Port),
		PublicKey:         fastly.String(ftp.PublicKey),
		Username:          fastly.String(ftp.Username),
		Password:          fastly.String(ftp.Password),
		Path:              fastly.String(ftp.Path),
		Period:            fastly.Uint(ftp.Period),
		FormatVersion:     fastly.Uint(ftp.FormatVersion),
		GzipLevel:         fastly.Uint8(ftp.GzipLevel),
		Format:            fastly.String(ftp.Format),
		ResponseCondition: fastly.String(ftp.ResponseCondition),
		TimestampFormat:   fastly.String(ftp.TimestampFormat),
		Placement:         fastly.String(ftp.Placement),
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Address.Valid {
		input.Address = fastly.String(c.Address.Value)
	}

	if c.Port.Valid {
		input.Port = fastly.Uint(c.Port.Value)
	}

	if c.Username.Valid {
		input.Username = fastly.String(c.Username.Value)
	}

	if c.Password.Valid {
		input.Password = fastly.String(c.Password.Value)
	}

	if c.PublicKey.Valid {
		input.PublicKey = fastly.String(c.PublicKey.Value)
	}

	if c.Path.Valid {
		input.Path = fastly.String(c.Path.Value)
	}

	if c.Period.Valid {
		input.Period = fastly.Uint(c.Period.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.GzipLevel.Valid {
		input.GzipLevel = fastly.Uint8(c.GzipLevel.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
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
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	ftp, err := c.Globals.Client.UpdateFTP(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated FTP logging endpoint %s (service %s version %d)", ftp.Name, ftp.ServiceID, ftp.ServiceVersion)
	return nil
}
