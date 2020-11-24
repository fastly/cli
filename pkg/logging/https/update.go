package https

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update HTTPS logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	URL               common.OptionalString
	RequestMaxEntries common.OptionalUint
	RequestMaxBytes   common.OptionalUint
	TLSCACert         common.OptionalString
	TLSClientCert     common.OptionalString
	TLSClientKey      common.OptionalString
	TLSHostname       common.OptionalString
	MessageType       common.OptionalString
	ContentType       common.OptionalString
	HeaderName        common.OptionalString
	HeaderValue       common.OptionalString
	Method            common.OptionalString
	JSONFormat        common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	Placement         common.OptionalString
	ResponseCondition common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update an HTTPS logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the HTTPS logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the HTTPS logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("url", "URL that log data will be sent to. Must use the https protocol").Action(c.URL.Set).StringVar(&c.URL.Value)
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

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateHTTPSInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	https, err := c.Globals.Client.GetHTTPS(&fastly.GetHTTPSInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateHTTPSInput{
		ServiceID:         https.ServiceID,
		ServiceVersion:    https.ServiceVersion,
		Name:              https.Name,
		NewName:           fastly.String(https.Name),
		ResponseCondition: fastly.String(https.ResponseCondition),
		Format:            fastly.String(https.Format),
		URL:               fastly.String(https.URL),
		RequestMaxEntries: fastly.Uint(https.RequestMaxEntries),
		RequestMaxBytes:   fastly.Uint(https.RequestMaxBytes),
		ContentType:       fastly.String(https.ContentType),
		HeaderName:        fastly.String(https.HeaderName),
		HeaderValue:       fastly.String(https.HeaderValue),
		Method:            fastly.String(https.Method),
		JSONFormat:        fastly.String(https.JSONFormat),
		Placement:         fastly.String(https.Placement),
		TLSCACert:         fastly.String(https.TLSCACert),
		TLSClientCert:     fastly.String(https.TLSClientCert),
		TLSClientKey:      fastly.String(https.TLSClientKey),
		TLSHostname:       fastly.String(https.TLSHostname),
		MessageType:       fastly.String(https.MessageType),
		FormatVersion:     fastly.Uint(https.FormatVersion),
	}

	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.URL.WasSet {
		input.URL = fastly.String(c.URL.Value)
	}

	if c.ContentType.WasSet {
		input.ContentType = fastly.String(c.ContentType.Value)
	}

	if c.JSONFormat.WasSet {
		input.JSONFormat = fastly.String(c.JSONFormat.Value)
	}

	if c.HeaderName.WasSet {
		input.HeaderName = fastly.String(c.HeaderName.Value)
	}

	if c.HeaderValue.WasSet {
		input.HeaderValue = fastly.String(c.HeaderValue.Value)
	}

	if c.Method.WasSet {
		input.Method = fastly.String(c.Method.Value)
	}

	if c.RequestMaxEntries.WasSet {
		input.RequestMaxEntries = fastly.Uint(c.RequestMaxEntries.Value)
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = fastly.Uint(c.RequestMaxBytes.Value)
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = fastly.String(c.TLSCACert.Value)
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = fastly.String(c.TLSClientCert.Value)
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = fastly.String(c.TLSClientKey.Value)
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = fastly.String(c.TLSHostname.Value)
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

	if c.MessageType.WasSet {
		input.MessageType = fastly.String(c.MessageType.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	https, err := c.Globals.Client.UpdateHTTPS(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated HTTPS logging endpoint %s (service %s version %d)", https.Name, https.ServiceID, https.ServiceVersion)
	return nil
}
