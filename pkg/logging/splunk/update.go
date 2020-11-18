package splunk

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update Splunk logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	URL               common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	ResponseCondition common.OptionalString
	Placement         common.OptionalString
	Token             common.OptionalString
	TLSCACert         common.OptionalString
	TLSHostname       common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Splunk logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Splunk logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Splunk logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("url", "The URL to POST to.").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.TLSCACert.Set).StringVar(&c.TLSCACert.Value)
	c.CmdClause.Flag("tls-hostname", "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)").Action(c.TLSHostname.Set).StringVar(&c.TLSHostname.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "	Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("auth-token", "").Action(c.Token.Set).StringVar(&c.Token.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateSplunkInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	splunk, err := c.Globals.Client.GetSplunk(&fastly.GetSplunkInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateSplunkInput{
		ServiceID:         splunk.ServiceID,
		ServiceVersion:    splunk.ServiceVersion,
		Name:              splunk.Name,
		NewName:           fastly.String(splunk.Name),
		URL:               fastly.String(splunk.URL),
		Format:            fastly.String(splunk.Format),
		FormatVersion:     fastly.Uint(splunk.FormatVersion),
		ResponseCondition: fastly.String(splunk.ResponseCondition),
		Placement:         fastly.String(splunk.Placement),
		Token:             fastly.String(splunk.Token),
		TLSCACert:         fastly.String(splunk.TLSCACert),
		TLSHostname:       fastly.String(splunk.TLSHostname),
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.URL.Valid {
		input.URL = fastly.String(c.URL.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.Valid {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.Token.Valid {
		input.Token = fastly.String(c.Token.Value)
	}

	if c.TLSCACert.Valid {
		input.TLSCACert = fastly.String(c.TLSCACert.Value)
	}

	if c.TLSHostname.Valid {
		input.TLSHostname = fastly.String(c.TLSHostname.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	splunk, err := c.Globals.Client.UpdateSplunk(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Splunk logging endpoint %s (service %s version %d)", splunk.Name, splunk.ServiceID, splunk.ServiceVersion)
	return nil
}
