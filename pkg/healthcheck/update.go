package healthcheck

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update healthchecks.
type UpdateCommand struct {
	cmd.Base
	manifest       manifest.Data
	input          fastly.UpdateHealthCheckInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	NewName          cmd.OptionalString
	Comment          cmd.OptionalString
	Method           cmd.OptionalString
	Host             cmd.OptionalString
	Path             cmd.OptionalString
	HTTPVersion      cmd.OptionalString
	Timeout          cmd.OptionalUint
	CheckInterval    cmd.OptionalUint
	ExpectedResponse cmd.OptionalUint
	Window           cmd.OptionalUint
	Threshold        cmd.OptionalUint
	Initial          cmd.OptionalUint
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a healthcheck on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Required().StringVar(&c.input.Name)
	c.CmdClause.Flag("new-name", "Healthcheck name").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.Comment.Set).StringVar(&c.Comment.Value)
	c.CmdClause.Flag("method", "Which HTTP method to use").Action(c.Method.Set).StringVar(&c.Method.Value)
	c.CmdClause.Flag("host", "Which host to check").Action(c.Host.Set).StringVar(&c.Host.Value)
	c.CmdClause.Flag("path", "The path to check").Action(c.Path.Set).StringVar(&c.Path.Value)
	c.CmdClause.Flag("http-version", "Whether to use version 1.0 or 1.1 HTTP").Action(c.HTTPVersion.Set).StringVar(&c.HTTPVersion.Value)
	c.CmdClause.Flag("timeout", "Timeout in milliseconds").Action(c.Timeout.Set).UintVar(&c.Timeout.Value)
	c.CmdClause.Flag("check-interval", "How often to run the healthcheck in milliseconds").Action(c.CheckInterval.Set).UintVar(&c.CheckInterval.Value)
	c.CmdClause.Flag("expected-response", "The status code expected from the host").Action(c.ExpectedResponse.Set).UintVar(&c.ExpectedResponse.Value)
	c.CmdClause.Flag("window", "The number of most recent healthcheck queries to keep for this healthcheck").Action(c.Window.Set).UintVar(&c.Window.Value)
	c.CmdClause.Flag("threshold", "How many healthchecks must succeed to be considered healthy").Action(c.Threshold.Set).UintVar(&c.Threshold.Value)
	c.CmdClause.Flag("initial", "When loading a config, the initial number of probes to be seen as OK").Action(c.Initial.Set).UintVar(&c.Initial.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersion.Number

	if c.NewName.WasSet {
		c.input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Comment.WasSet {
		c.input.Comment = fastly.String(c.Comment.Value)
	}

	if c.Method.WasSet {
		c.input.Method = fastly.String(c.Method.Value)
	}

	if c.Host.WasSet {
		c.input.Host = fastly.String(c.Host.Value)
	}

	if c.Path.WasSet {
		c.input.Path = fastly.String(c.Path.Value)
	}

	if c.HTTPVersion.WasSet {
		c.input.HTTPVersion = fastly.String(c.HTTPVersion.Value)
	}

	if c.Timeout.WasSet {
		c.input.Timeout = fastly.Uint(c.Timeout.Value)
	}

	if c.CheckInterval.WasSet {
		c.input.CheckInterval = fastly.Uint(c.CheckInterval.Value)
	}

	if c.ExpectedResponse.WasSet {
		c.input.ExpectedResponse = fastly.Uint(c.ExpectedResponse.Value)
	}

	if c.Window.WasSet {
		c.input.Window = fastly.Uint(c.Window.Value)
	}

	if c.Threshold.WasSet {
		c.input.Threshold = fastly.Uint(c.Threshold.Value)
	}

	if c.Initial.WasSet {
		c.input.Initial = fastly.Uint(c.Initial.Value)
	}

	h, err := c.Globals.Client.UpdateHealthCheck(&c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Updated healthcheck %s (service %s version %d)", h.Name, h.ServiceID, h.ServiceVersion)
	return nil
}
