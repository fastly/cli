package https

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// CreateCommand calls the Fastly API to create an HTTPS logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	URL            string
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	RequestMaxEntries cmd.OptionalUint
	RequestMaxBytes   cmd.OptionalUint
	TLSCACert         cmd.OptionalString
	TLSClientCert     cmd.OptionalString
	TLSClientKey      cmd.OptionalString
	TLSHostname       cmd.OptionalString
	MessageType       cmd.OptionalString
	ContentType       cmd.OptionalString
	HeaderName        cmd.OptionalString
	HeaderValue       cmd.OptionalString
	Method            cmd.OptionalString
	JSONFormat        cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalUint
	Placement         cmd.OptionalString
	ResponseCondition cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest.File.SetOutput(c.Globals.Output)
	c.Manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an HTTPS logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the HTTPS logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.ServiceVersion.Value,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("url", "URL that log data will be sent to. Must use the https protocol").Required().StringVar(&c.URL)
	c.RegisterServiceIDFlag(&c.Manifest.Flag.ServiceID)
	c.CmdClause.Flag("content-type", "Content type of the header sent with the request").Action(c.ContentType.Set).StringVar(&c.ContentType.Value)
	c.CmdClause.Flag("header-name", "Name of the custom header sent with the request").Action(c.HeaderName.Set).StringVar(&c.HeaderName.Value)
	c.CmdClause.Flag("header-value", "Value of the custom header sent with the request").Action(c.HeaderValue.Set).StringVar(&c.HeaderValue.Value)
	c.CmdClause.Flag("method", "HTTP method used for request. Can be POST or PUT. Defaults to POST if not specified").Action(c.Method.Set).StringVar(&c.Method.Value)
	c.CmdClause.Flag("json-format", "Enforces valid JSON formatting for log entries. Can be disabled 0, array of json (wraps JSON log batches in an array) 1, or newline delimited json (places each JSON log entry onto a new line in a batch) 2").Action(c.JSONFormat.Set).StringVar(&c.JSONFormat.Value)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.TLSCACert.Set).StringVar(&c.TLSCACert.Value)
	c.CmdClause.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.TLSClientCert.Set).StringVar(&c.TLSClientCert.Value)
	c.CmdClause.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.TLSClientKey.Set).StringVar(&c.TLSClientKey.Value)
	c.CmdClause.Flag("tls-hostname", "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)").Action(c.TLSHostname.Set).StringVar(&c.TLSHostname.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("format", "Apache style log formatting. Your log must produce valid JSON that HTTPS can ingest").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("request-max-entries", "Maximum number of logs to append to a batch, if non-zero. Defaults to 0 for unbounded").Action(c.RequestMaxEntries.Set).UintVar(&c.RequestMaxEntries.Value)
	c.CmdClause.Flag("request-max-bytes", "Maximum size of log batch, if non-zero. Defaults to 0 for unbounded").Action(c.RequestMaxBytes.Set).UintVar(&c.RequestMaxBytes.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateHTTPSInput, error) {
	var input fastly.CreateHTTPSInput

	input.ServiceID = serviceID
	input.Name = c.EndpointName
	input.URL = c.URL
	input.ServiceVersion = serviceVersion

	if c.ContentType.WasSet {
		input.ContentType = c.ContentType.Value
	}

	if c.HeaderName.WasSet {
		input.HeaderName = c.HeaderName.Value
	}

	if c.HeaderValue.WasSet {
		input.HeaderValue = c.HeaderValue.Value
	}

	if c.Method.WasSet {
		input.Method = c.Method.Value
	}

	if c.JSONFormat.WasSet {
		input.JSONFormat = c.JSONFormat.Value
	}

	if c.RequestMaxEntries.WasSet {
		input.RequestMaxEntries = c.RequestMaxEntries.Value
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = c.RequestMaxBytes.Value
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = c.TLSCACert.Value
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = c.TLSClientCert.Value
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = c.TLSClientKey.Value
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = c.TLSHostname.Value
	}

	if c.Format.WasSet {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.Placement.WasSet {
		input.Placement = c.Placement.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = c.MessageType.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
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

	d, err := c.Globals.Client.CreateHTTPS(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created HTTPS logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
