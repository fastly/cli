package healthcheck

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update healthchecks.
type UpdateCommand struct {
	argparser.Base
	input          fastly.UpdateHealthCheckInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
	autoClone      argparser.OptionalAutoClone

	NewName          argparser.OptionalString
	Comment          argparser.OptionalString
	Method           argparser.OptionalString
	Host             argparser.OptionalString
	Path             argparser.OptionalString
	HTTPVersion      argparser.OptionalString
	Timeout          argparser.OptionalInt
	CheckInterval    argparser.OptionalInt
	ExpectedResponse argparser.OptionalInt
	Window           argparser.OptionalInt
	Threshold        argparser.OptionalInt
	Initial          argparser.OptionalInt
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a healthcheck on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Required().StringVar(&c.input.Name)
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
	c.CmdClause.Flag("check-interval", "How often to run the healthcheck in milliseconds").Action(c.CheckInterval.Set).IntVar(&c.CheckInterval.Value)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.Comment.Set).StringVar(&c.Comment.Value)
	c.CmdClause.Flag("expected-response", "The status code expected from the host").Action(c.ExpectedResponse.Set).IntVar(&c.ExpectedResponse.Value)
	c.CmdClause.Flag("host", "Which host to check").Action(c.Host.Set).StringVar(&c.Host.Value)
	c.CmdClause.Flag("http-version", "Whether to use version 1.0 or 1.1 HTTP").Action(c.HTTPVersion.Set).StringVar(&c.HTTPVersion.Value)
	c.CmdClause.Flag("initial", "When loading a config, the initial number of probes to be seen as OK").Action(c.Initial.Set).IntVar(&c.Initial.Value)
	c.CmdClause.Flag("method", "Which HTTP method to use").Action(c.Method.Set).StringVar(&c.Method.Value)
	c.CmdClause.Flag("new-name", "Healthcheck name").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("path", "The path to check").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("threshold", "How many healthchecks must succeed to be considered healthy").Action(c.Threshold.Set).IntVar(&c.Threshold.Value)
	c.CmdClause.Flag("timeout", "Timeout in milliseconds").Action(c.Timeout.Set).IntVar(&c.Timeout.Value)
	c.CmdClause.Flag("window", "The number of most recent healthcheck queries to keep for this healthcheck").Action(c.Window.Set).IntVar(&c.Window.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	if c.NewName.WasSet {
		c.input.NewName = &c.NewName.Value
	}
	if c.Comment.WasSet {
		c.input.Comment = &c.Comment.Value
	}
	if c.Method.WasSet {
		c.input.Method = &c.Method.Value
	}
	if c.Host.WasSet {
		c.input.Host = &c.Host.Value
	}
	if c.Path.WasSet {
		c.input.Path = &c.Path.Value
	}
	if c.HTTPVersion.WasSet {
		c.input.HTTPVersion = &c.HTTPVersion.Value
	}
	if c.Timeout.WasSet {
		c.input.Timeout = &c.Timeout.Value
	}
	if c.CheckInterval.WasSet {
		c.input.CheckInterval = &c.CheckInterval.Value
	}
	if c.ExpectedResponse.WasSet {
		c.input.ExpectedResponse = &c.ExpectedResponse.Value
	}
	if c.Window.WasSet {
		c.input.Window = &c.Window.Value
	}
	if c.Threshold.WasSet {
		c.input.Threshold = &c.Threshold.Value
	}
	if c.Initial.WasSet {
		c.input.Initial = &c.Initial.Value
	}

	h, err := c.Globals.APIClient.UpdateHealthCheck(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Updated healthcheck %s (service %s version %d)",
		fastly.ToValue(h.Name),
		fastly.ToValue(h.ServiceID),
		fastly.ToValue(h.ServiceVersion),
	)
	return nil
}
