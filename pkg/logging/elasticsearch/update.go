package elasticsearch

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update Elasticsearch logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	Index             common.OptionalString
	URL               common.OptionalString
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

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update an Elasticsearch logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Elasticsearch logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("index", `The name of the Elasticsearch index to send documents (logs) to. The index must follow the Elasticsearch index format rules (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html). We support strftime (http://man7.org/linux/man-pages/man3/strftime.3.html) interpolated variables inside braces prefixed with a pound symbol. For example, #{%F} will interpolate as YYYY-MM-DD with today's date`).Action(c.Index.Set).StringVar(&c.Index.Value)
	c.CmdClause.Flag("url", "The URL to stream logs to. Must use HTTPS.").Action(c.URL.Set).StringVar(&c.URL.Value)
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
func (c *UpdateCommand) createInput() (*fastly.UpdateElasticsearchInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	elasticsearch, err := c.Globals.Client.GetElasticsearch(&fastly.GetElasticsearchInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateElasticsearchInput{
		ServiceID:         elasticsearch.ServiceID,
		ServiceVersion:    elasticsearch.ServiceVersion,
		Name:              elasticsearch.Name,
		NewName:           fastly.String(elasticsearch.Name),
		ResponseCondition: fastly.String(elasticsearch.ResponseCondition),
		Format:            fastly.String(elasticsearch.Format),
		Index:             fastly.String(elasticsearch.Index),
		URL:               fastly.String(elasticsearch.URL),
		Pipeline:          fastly.String(elasticsearch.Pipeline),
		User:              fastly.String(elasticsearch.User),
		Password:          fastly.String(elasticsearch.Password),
		RequestMaxEntries: fastly.Uint(elasticsearch.RequestMaxEntries),
		RequestMaxBytes:   fastly.Uint(elasticsearch.RequestMaxBytes),
		Placement:         fastly.String(elasticsearch.Placement),
		TLSCACert:         fastly.String(elasticsearch.TLSCACert),
		TLSClientCert:     fastly.String(elasticsearch.TLSClientCert),
		TLSClientKey:      fastly.String(elasticsearch.TLSClientKey),
		TLSHostname:       fastly.String(elasticsearch.TLSHostname),
		FormatVersion:     fastly.Uint(elasticsearch.FormatVersion),
	}

	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Index.Valid {
		input.Index = fastly.String(c.Index.Value)
	}

	if c.URL.Valid {
		input.URL = fastly.String(c.URL.Value)
	}

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
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	elasticsearch, err := c.Globals.Client.UpdateElasticsearch(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Elasticsearch logging endpoint %s (service %s version %d)", elasticsearch.Name, elasticsearch.ServiceID, elasticsearch.ServiceVersion)
	return nil
}
