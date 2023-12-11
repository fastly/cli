package elasticsearch

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an Elasticsearch logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// Optional.
	AutoClone         argparser.OptionalAutoClone
	EndpointName      argparser.OptionalString // Can't shadow argparser.Base method Name().
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	Index             argparser.OptionalString
	Password          argparser.OptionalString
	Pipeline          argparser.OptionalString
	Placement         argparser.OptionalString
	RequestMaxBytes   argparser.OptionalInt
	RequestMaxEntries argparser.OptionalInt
	ResponseCondition argparser.OptionalString
	TLSCACert         argparser.OptionalString
	TLSClientCert     argparser.OptionalString
	TLSClientKey      argparser.OptionalString
	TLSHostname       argparser.OptionalString
	URL               argparser.OptionalString
	User              argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an Elasticsearch logging endpoint on a Fastly service version").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the Elasticsearch logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	c.CmdClause.Flag("index", `The name of the Elasticsearch index to send documents (logs) to. The index must follow the Elasticsearch index format rules (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html). We support strftime (http://man7.org/linux/man-pages/man3/strftime.3.html) interpolated variables inside braces prefixed with a pound symbol. For example, #{%F} will interpolate as YYYY-MM-DD with today's date`).Action(c.Index.Set).StringVar(&c.Index.Value)
	common.Placement(c.CmdClause, &c.Placement)
	c.CmdClause.Flag("pipeline", "The ID of the Elasticsearch ingest pipeline to apply pre-process transformations to before indexing. For example my_pipeline_id. Learn more about creating a pipeline in the Elasticsearch docs (https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest.html)").Action(c.Password.Set).StringVar(&c.Pipeline.Value)
	c.CmdClause.Flag("request-max-bytes", "Maximum size of log batch, if non-zero. Defaults to 100MB").Action(c.RequestMaxBytes.Set).IntVar(&c.RequestMaxBytes.Value)
	c.CmdClause.Flag("request-max-entries", "Maximum number of logs to append to a batch, if non-zero. Defaults to 10k").Action(c.RequestMaxEntries.Set).IntVar(&c.RequestMaxEntries.Value)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	common.TLSCACert(c.CmdClause, &c.TLSCACert)
	common.TLSClientCert(c.CmdClause, &c.TLSClientCert)
	common.TLSClientKey(c.CmdClause, &c.TLSClientKey)
	common.TLSHostname(c.CmdClause, &c.TLSHostname)
	c.CmdClause.Flag("url", "The URL to stream logs to. Must use HTTPS.").Action(c.URL.Set).StringVar(&c.URL.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateElasticsearchInput, error) {
	var input fastly.CreateElasticsearchInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.Index.WasSet {
		input.Index = &c.Index.Value
	}
	if c.URL.WasSet {
		input.URL = &c.URL.Value
	}

	if c.Pipeline.WasSet {
		input.Pipeline = &c.Pipeline.Value
	}

	if c.RequestMaxEntries.WasSet {
		input.RequestMaxEntries = &c.RequestMaxEntries.Value
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = &c.RequestMaxBytes.Value
	}

	if c.User.WasSet {
		input.User = &c.User.Value
	}

	if c.Password.WasSet {
		input.Password = &c.Password.Value
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = &c.TLSCACert.Value
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = &c.TLSClientCert.Value
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = &c.TLSClientKey.Value
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = &c.TLSHostname.Value
	}

	if c.Format.WasSet {
		input.Format = &c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	d, err := c.Globals.APIClient.CreateElasticsearch(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Created Elasticsearch logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}
