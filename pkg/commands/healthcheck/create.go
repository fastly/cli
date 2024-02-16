package healthcheck

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// CreateCommand calls the Fastly API to create healthchecks.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	autoClone        argparser.OptionalAutoClone
	checkInterval    argparser.OptionalInt
	comment          argparser.OptionalString
	expectedResponse argparser.OptionalInt
	host             argparser.OptionalString
	httpVersion      argparser.OptionalString
	initial          argparser.OptionalInt
	method           argparser.OptionalString
	name             argparser.OptionalString
	path             argparser.OptionalString
	serviceName      argparser.OptionalServiceNameID
	threshold        argparser.OptionalInt
	timeout          argparser.OptionalInt
	window           argparser.OptionalInt
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a healthcheck on a Fastly service version").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("check-interval", "How often to run the healthcheck in milliseconds").Action(c.checkInterval.Set).IntVar(&c.checkInterval.Value)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("expected-response", "The status code expected from the host").Action(c.expectedResponse.Set).IntVar(&c.expectedResponse.Value)
	c.CmdClause.Flag("host", "Which host to check").Action(c.host.Set).StringVar(&c.host.Value)
	c.CmdClause.Flag("http-version", "Whether to use version 1.0 or 1.1 HTTP").Action(c.httpVersion.Set).StringVar(&c.httpVersion.Value)
	c.CmdClause.Flag("initial", "When loading a config, the initial number of probes to be seen as OK").Action(c.initial.Set).IntVar(&c.initial.Value)
	c.CmdClause.Flag("method", "Which HTTP method to use").Action(c.method.Set).StringVar(&c.method.Value)
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("path", "The path to check").Action(c.path.Set).StringVar(&c.path.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("threshold", "How many healthchecks must succeed to be considered healthy").Action(c.threshold.Set).IntVar(&c.threshold.Value)
	c.CmdClause.Flag("timeout", "Timeout in milliseconds").Action(c.timeout.Set).IntVar(&c.timeout.Value)
	c.CmdClause.Flag("window", "The number of most recent healthcheck queries to keep for this healthcheck").Action(c.window.Set).IntVar(&c.window.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}
	input := fastly.CreateHealthCheckInput{
		ServiceID:      serviceID,
		ServiceVersion: fastly.ToValue(serviceVersion.Number),
	}

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.comment.WasSet {
		input.Comment = &c.comment.Value
	}
	if c.method.WasSet {
		input.Method = &c.method.Value
	}
	if c.host.WasSet {
		input.Host = &c.host.Value
	}
	if c.path.WasSet {
		input.Path = &c.path.Value
	}
	if c.httpVersion.WasSet {
		input.HTTPVersion = &c.httpVersion.Value
	}
	if c.timeout.WasSet {
		input.Timeout = &c.timeout.Value
	}
	if c.checkInterval.WasSet {
		input.CheckInterval = &c.checkInterval.Value
	}
	if c.expectedResponse.WasSet {
		input.ExpectedResponse = &c.expectedResponse.Value
	}
	if c.window.WasSet {
		input.Window = &c.window.Value
	}
	if c.threshold.WasSet {
		input.Threshold = &c.threshold.Value
	}
	if c.initial.WasSet {
		input.Initial = &c.initial.Value
	}

	h, err := c.Globals.APIClient.CreateHealthCheck(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created healthcheck %s (service %s version %d)", fastly.ToValue(h.Name), fastly.ToValue(h.ServiceID), fastly.ToValue(h.ServiceVersion))
	return nil
}
