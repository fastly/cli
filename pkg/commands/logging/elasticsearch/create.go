package elasticsearch

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create an Elasticsearch logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	Index          string
	URL            string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	Pipeline          cmd.OptionalString
	RequestMaxEntries cmd.OptionalInt
	RequestMaxBytes   cmd.OptionalInt
	User              cmd.OptionalString
	Password          cmd.OptionalString
	TLSCACert         cmd.OptionalString
	TLSClientCert     cmd.OptionalString
	TLSClientKey      cmd.OptionalString
	TLSHostname       cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	Placement         cmd.OptionalString
	ResponseCondition cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create an Elasticsearch logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("index", `The name of the Elasticsearch index to send documents (logs) to. The index must follow the Elasticsearch index format rules (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html). We support strftime (http://man7.org/linux/man-pages/man3/strftime.3.html) interpolated variables inside braces prefixed with a pound symbol. For example, #{%F} will interpolate as YYYY-MM-DD with today's date`).Required().StringVar(&c.Index)
	c.CmdClause.Flag("url", "The URL to stream logs to. Must use HTTPS.").Required().StringVar(&c.URL)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	c.CmdClause.Flag("pipeline", "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing. For example my_pipeline_id. Learn more about creating a pipeline in the Elasticsearch docs (https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest.html)").Action(c.Password.Set).StringVar(&c.Pipeline.Value)
	common.TLSCACert(c.CmdClause, &c.TLSCACert)
	common.TLSClientCert(c.CmdClause, &c.TLSClientCert)
	common.TLSClientKey(c.CmdClause, &c.TLSClientKey)
	common.TLSHostname(c.CmdClause, &c.TLSHostname)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.Placement(c.CmdClause, &c.Placement)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("request-max-entries", "Maximum number of logs to append to a batch, if non-zero. Defaults to 10k").Action(c.RequestMaxEntries.Set).IntVar(&c.RequestMaxEntries.Value)
	c.CmdClause.Flag("request-max-bytes", "Maximum size of log batch, if non-zero. Defaults to 100MB").Action(c.RequestMaxBytes.Set).IntVar(&c.RequestMaxBytes.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateElasticsearchInput, error) {
	var input fastly.CreateElasticsearchInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Name = fastly.String(c.EndpointName)
	input.Index = fastly.String(c.Index)
	input.URL = fastly.String(c.URL)

	if c.Pipeline.WasSet {
		input.Pipeline = fastly.String(c.Pipeline.Value)
	}

	if c.RequestMaxEntries.WasSet {
		input.RequestMaxEntries = fastly.Int(c.RequestMaxEntries.Value)
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = fastly.Int(c.RequestMaxBytes.Value)
	}

	if c.User.WasSet {
		input.User = fastly.String(c.User.Value)
	}

	if c.Password.WasSet {
		input.Password = fastly.String(c.Password.Value)
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
		input.FormatVersion = fastly.Int(c.FormatVersion.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	d, err := c.Globals.APIClient.CreateElasticsearch(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created Elasticsearch logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
