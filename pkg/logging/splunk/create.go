package splunk

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create Splunk logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.CreateSplunkInput
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a Splunk logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the Splunk logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)

	c.CmdClause.Flag("url", "The URL to POST to").Required().StringVar(&c.Input.URL)

	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").StringVar(&c.Input.TLSCACert)
	c.CmdClause.Flag("tls-hostname", "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)").StringVar(&c.Input.TLSHostname)
	c.CmdClause.Flag("format", "Apache style log formatting").StringVar(&c.Input.Format)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").UintVar(&c.Input.FormatVersion)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").StringVar(&c.Input.ResponseCondition)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").StringVar(&c.Input.Placement)
	c.CmdClause.Flag("auth-token", "A Splunk token for use in posting logs over HTTP to your collector").StringVar(&c.Input.Token)

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	d, err := c.Globals.Client.CreateSplunk(&c.Input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Splunk logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
