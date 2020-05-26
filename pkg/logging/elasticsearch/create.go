package elasticsearch

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// CreateCommand calls the Fastly API to create Elasticsearch logging endpoints.
type CreateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int
	Index        string
	URL          string

	// optional
	Pipeline          common.OptionalString
	RequestMaxEntries common.OptionalUint
	RequestMaxBytes   common.OptionalUint
	User              common.OptionalString
	Password          common.OptionalString
	TLSCACert         common.OptionalString
	TLSClientCert     common.OptionalString
	TLSClientKey      common.OptionalString
	TLSHostname       common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	Placement         common.OptionalString
	ResponseCondition common.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent common.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create an Elasticsearch logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)

	c.CmdClause.Flag("index", `The name of the Elasticsearch index to send documents (logs) to. The index must follow the Elasticsearch index format rules (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html). We support strftime (http://man7.org/linux/man-pages/man3/strftime.3.html) interpolated variables inside braces prefixed with a pound symbol. For example, #{%F} will interpolate as YYYY-MM-DD with today's date`).Required().StringVar(&c.Index)
	c.CmdClause.Flag("url", "The URL to stream logs to. Must use HTTPS.").Required().StringVar(&c.URL)

	c.CmdClause.Flag("pipeline", "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing. For example my_pipeline_id. Learn more about creating a pipeline in the Elasticsearch docs (https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest.html)").Action(c.Password.Set).StringVar(&c.Pipeline.Value)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.TLSCACert.Set).StringVar(&c.TLSCACert.Value)
	c.CmdClause.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.TLSClientCert.Set).StringVar(&c.TLSClientCert.Value)
	c.CmdClause.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.TLSClientKey.Set).StringVar(&c.TLSClientKey.Value)
	c.CmdClause.Flag("tls-hostname", "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)").Action(c.TLSHostname.Set).StringVar(&c.TLSHostname.Value)
	c.CmdClause.Flag("format", "Apache style log formatting. Your log must produce valid JSON that Elasticsearch can ingest").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("request-max-entries", "Maximum number of logs to append to a batch, if non-zero. Defaults to 0 for unbounded").Action(c.RequestMaxEntries.Set).UintVar(&c.RequestMaxEntries.Value)
	c.CmdClause.Flag("request-max-bytes", "Maximum size of log batch, if non-zero. Defaults to 0 for unbounded").Action(c.RequestMaxBytes.Set).UintVar(&c.RequestMaxBytes.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput() (*fastly.CreateElasticsearchInput, error) {
	var input fastly.CreateElasticsearchInput

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	input.Service = serviceID
	input.Version = c.Version
	input.Name = fastly.String(c.EndpointName)
	input.Index = fastly.String(c.Index)
	input.URL = fastly.String(c.URL)

	if c.Pipeline.Valid {
		input.Pipeline = fastly.String(c.Pipeline.Value)
	}

	if c.RequestMaxEntries.Valid {
		input.RequestMaxEntries = fastly.Uint(c.RequestMaxEntries.Value)
	}

	if c.RequestMaxBytes.Valid {
		input.RequestMaxBytes = fastly.Uint(c.RequestMaxBytes.Value)
	}

	if c.User.Valid {
		input.User = fastly.String(c.User.Value)
	}

	if c.Password.Valid {
		input.Password = fastly.String(c.Password.Value)
	}

	if c.TLSCACert.Valid {
		input.TLSCACert = fastly.String(c.TLSCACert.Value)
	}

	if c.TLSClientCert.Valid {
		input.TLSClientCert = fastly.String(c.TLSClientCert.Value)
	}

	if c.TLSClientKey.Valid {
		input.TLSClientKey = fastly.String(c.TLSClientKey.Value)
	}

	if c.TLSHostname.Valid {
		input.TLSHostname = fastly.String(c.TLSHostname.Value)
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

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	d, err := c.Globals.Client.CreateElasticsearch(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Elasticsearch logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.Version)
	return nil
}
