package splunk

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a Splunk logging endpoint.
type UpdateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	NewName           cmd.OptionalString
	URL               cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalUint
	ResponseCondition cmd.OptionalString
	Placement         cmd.OptionalString
	Token             cmd.OptionalString
	TLSCACert         cmd.OptionalString
	TLSHostname       cmd.OptionalString
	TLSClientCert     cmd.OptionalString
	TLSClientKey      cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.Manifest.File.SetOutput(c.Globals.Output)
	c.Manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a Splunk logging endpoint on a Fastly service version")
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.ServiceVersion.Value,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("name", "The name of the Splunk logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterServiceIDFlag(&c.Manifest.Flag.ServiceID)
	c.CmdClause.Flag("new-name", "New name of the Splunk logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("url", "The URL to POST to.").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.TLSCACert.Set).StringVar(&c.TLSCACert.Value)
	c.CmdClause.Flag("tls-hostname", "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)").Action(c.TLSHostname.Set).StringVar(&c.TLSHostname.Value)
	c.CmdClause.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.TLSClientCert.Set).StringVar(&c.TLSClientCert.Value)
	c.CmdClause.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.TLSClientKey.Set).StringVar(&c.TLSClientKey.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "	Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("auth-token", "").Action(c.Token.Set).StringVar(&c.Token.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateSplunkInput, error) {
	input := fastly.UpdateSplunkInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.URL.WasSet {
		input.URL = fastly.String(c.URL.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.Token.WasSet {
		input.Token = fastly.String(c.Token.Value)
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = fastly.String(c.TLSCACert.Value)
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = fastly.String(c.TLSHostname.Value)
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = fastly.String(c.TLSClientCert.Value)
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = fastly.String(c.TLSClientKey.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		Client:             c.Globals.Client,
		Manifest:           c.Manifest,
		Out:                out,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	input, err := c.ConstructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	splunk, err := c.Globals.Client.UpdateSplunk(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated Splunk logging endpoint %s (service %s version %d)", splunk.Name, splunk.ServiceID, splunk.ServiceVersion)
	return nil
}
