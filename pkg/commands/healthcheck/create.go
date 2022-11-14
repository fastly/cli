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
	input            fastly.CreateHealthCheckInput
	autoClone        cmd.OptionalAutoClone
	checkInterval    cmd.OptionalUint
	expectedResponse cmd.OptionalUint
	initial          cmd.OptionalUint
	manifest         manifest.Data
	serviceName      cmd.OptionalServiceNameID
	serviceVersion   cmd.OptionalServiceVersion
	threshold        cmd.OptionalUint
	timeout          cmd.OptionalUint
	window           cmd.OptionalUint
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
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Required().StringVar(&c.input.Name)
	c.CmdClause.Flag("comment", "A descriptive note").StringVar(&c.input.Comment)
	c.CmdClause.Flag("method", "Which HTTP method to use").StringVar(&c.input.Method)
	c.CmdClause.Flag("host", "Which host to check").StringVar(&c.input.Host)
	c.CmdClause.Flag("path", "The path to check").StringVar(&c.input.Path)
	c.CmdClause.Flag("http-version", "Whether to use version 1.0 or 1.1 HTTP").StringVar(&c.input.HTTPVersion)
	c.CmdClause.Flag("timeout", "Timeout in milliseconds").Action(c.timeout.Set).UintVar(&c.timeout.Value)
	c.CmdClause.Flag("check-interval", "How often to run the healthcheck in milliseconds").Action(c.checkInterval.Set).UintVar(&c.checkInterval.Value)
	c.CmdClause.Flag("expected-response", "The status code expected from the host").Action(c.expectedResponse.Set).UintVar(&c.expectedResponse.Value)
	c.CmdClause.Flag("window", "The number of most recent healthcheck queries to keep for this healthcheck").Action(c.window.Set).UintVar(&c.window.Value)
	c.CmdClause.Flag("threshold", "How many healthchecks must succeed to be considered healthy").Action(c.threshold.Set).UintVar(&c.threshold.Value)
	c.CmdClause.Flag("initial", "When loading a config, the initial number of probes to be seen as OK").Action(c.initial.Set).UintVar(&c.initial.Value)
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

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersion.Number

	if c.timeout.WasSet {
		c.input.Timeout = fastly.Uint(c.timeout.Value)
	}
	if c.checkInterval.WasSet {
		c.input.CheckInterval = fastly.Uint(c.checkInterval.Value)
	}
	if c.expectedResponse.WasSet {
		c.input.ExpectedResponse = fastly.Uint(c.expectedResponse.Value)
	}
	if c.window.WasSet {
		c.input.Window = fastly.Uint(c.window.Value)
	}
	if c.threshold.WasSet {
		c.input.Threshold = fastly.Uint(c.threshold.Value)
	}
	if c.initial.WasSet {
		c.input.Initial = fastly.Uint(c.initial.Value)
	}

	h, err := c.Globals.APIClient.CreateHealthCheck(&c.input)
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
