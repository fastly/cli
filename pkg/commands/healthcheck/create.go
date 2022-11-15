package healthcheck

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create healthchecks.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data
	// required
	name           string
	serviceVersion cmd.OptionalServiceVersion
	// optional
	autoClone        cmd.OptionalAutoClone
	checkInterval    cmd.OptionalInt
	comment          cmd.OptionalString
	expectedResponse cmd.OptionalInt
	host             cmd.OptionalString
	httpVersion      cmd.OptionalString
	initial          cmd.OptionalInt
	method           cmd.OptionalString
	path             cmd.OptionalString
	serviceName      cmd.OptionalServiceNameID
	threshold        cmd.OptionalInt
	timeout          cmd.OptionalInt
	window           cmd.OptionalInt
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("create", "Create a healthcheck on a Fastly service version").Alias("add")
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Required().StringVar(&c.name)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("method", "Which HTTP method to use").Action(c.method.Set).StringVar(&c.method.Value)
	c.CmdClause.Flag("host", "Which host to check").Action(c.host.Set).StringVar(&c.host.Value)
	c.CmdClause.Flag("path", "The path to check").Action(c.path.Set).StringVar(&c.path.Value)
	c.CmdClause.Flag("http-version", "Whether to use version 1.0 or 1.1 HTTP").Action(c.httpVersion.Set).StringVar(&c.httpVersion.Value)
	c.CmdClause.Flag("timeout", "Timeout in milliseconds").Action(c.timeout.Set).IntVar(&c.timeout.Value)
	c.CmdClause.Flag("check-interval", "How often to run the healthcheck in milliseconds").Action(c.checkInterval.Set).IntVar(&c.checkInterval.Value)
	c.CmdClause.Flag("expected-response", "The status code expected from the host").Action(c.expectedResponse.Set).IntVar(&c.expectedResponse.Value)
	c.CmdClause.Flag("window", "The number of most recent healthcheck queries to keep for this healthcheck").Action(c.window.Set).IntVar(&c.window.Value)
	c.CmdClause.Flag("threshold", "How many healthchecks must succeed to be considered healthy").Action(c.threshold.Set).IntVar(&c.threshold.Value)
	c.CmdClause.Flag("initial", "When loading a config, the initial number of probes to be seen as OK").Action(c.initial.Set).IntVar(&c.initial.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}
	input := fastly.CreateHealthCheckInput{
		Name:           fastly.String(c.name),
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	}

	if c.comment.WasSet {
		input.Comment = fastly.String(c.comment.Value)
	}
	if c.method.WasSet {
		input.Method = fastly.String(c.method.Value)
	}
	if c.host.WasSet {
		input.Host = fastly.String(c.host.Value)
	}
	if c.path.WasSet {
		input.Path = fastly.String(c.path.Value)
	}
	if c.httpVersion.WasSet {
		input.HTTPVersion = fastly.String(c.httpVersion.Value)
	}
	if c.timeout.WasSet {
		input.Timeout = fastly.Int(c.timeout.Value)
	}
	if c.checkInterval.WasSet {
		input.CheckInterval = fastly.Int(c.checkInterval.Value)
	}
	if c.expectedResponse.WasSet {
		input.ExpectedResponse = fastly.Int(c.expectedResponse.Value)
	}
	if c.window.WasSet {
		input.Window = fastly.Int(c.window.Value)
	}
	if c.threshold.WasSet {
		input.Threshold = fastly.Int(c.threshold.Value)
	}
	if c.initial.WasSet {
		input.Initial = fastly.Int(c.initial.Value)
	}

	h, err := c.Globals.APIClient.CreateHealthCheck(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Created healthcheck %s (service %s version %d)", h.Name, h.ServiceID, h.ServiceVersion)
	return nil
}
