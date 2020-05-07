package sumologic

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create Sumologic logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CreateSumologicInput
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a Sumologic logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the Sumologic logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)

	c.CmdClause.Flag("url", "The URL to POST to").Required().StringVar(&c.Input.URL)
	c.CmdClause.Flag("format", "Apache style log formatting").StringVar(&c.Input.Format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").IntVar(&c.Input.FormatVersion)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").StringVar(&c.Input.ResponseCondition)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").StringVar(&c.Input.MessageType)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").StringVar(&c.Input.Placement)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	d, err := c.Globals.Client.CreateSumologic(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Sumologic logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
