package ftp

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create FTP logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CreateFTPInput
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an FTP logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the FTP logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)

	c.CmdClause.Flag("address", "An hostname or IPv4 address").Required().StringVar(&c.Input.Address)
	c.CmdClause.Flag("user", "The username for the server (can be anonymous)").Required().StringVar(&c.Input.Username)
	c.CmdClause.Flag("password", "The password for the server (for anonymous use an email address)").Required().StringVar(&c.Input.Password)

	c.CmdClause.Flag("port", "The port number").UintVar(&c.Input.Port)
	c.CmdClause.Flag("path", "The path to upload log files to. If the path ends in / then it is treated as a directory").StringVar(&c.Input.Path)
	c.CmdClause.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").UintVar(&c.Input.Period)
	c.CmdClause.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Uint8Var(&c.Input.GzipLevel)
	c.CmdClause.Flag("format", "Apache style log formatting").StringVar(&c.Input.Format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").UintVar(&c.Input.FormatVersion)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").StringVar(&c.Input.ResponseCondition)
	c.CmdClause.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).StringVar(&c.Input.TimestampFormat)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").StringVar(&c.Input.Placement)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID
	d, err := c.Globals.Client.CreateFTP(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created FTP logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
