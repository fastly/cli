package healthcheck

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// CreateCommand calls the Fastly API to create healthchecks.
type CreateCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.CreateHealthCheckInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("create", "Create a healthcheck on a Fastly service version").Alias("add")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Required().StringVar(&c.Input.Name)
	c.CmdClause.Flag("comment", "A descriptive note").StringVar(&c.Input.Comment)
	c.CmdClause.Flag("method", "Which HTTP method to use").StringVar(&c.Input.Method)
	c.CmdClause.Flag("host", "Which host to check").StringVar(&c.Input.Host)
	c.CmdClause.Flag("path", "The path to check").StringVar(&c.Input.Path)
	c.CmdClause.Flag("http-version", "Whether to use version 1.0 or 1.1 HTTP").StringVar(&c.Input.HTTPVersion)
	c.CmdClause.Flag("timeout", "Timeout in milliseconds").UintVar(&c.Input.Timeout)
	c.CmdClause.Flag("check-interval", "How often to run the healthcheck in milliseconds").UintVar(&c.Input.CheckInterval)
	c.CmdClause.Flag("expected-response", "The status code expected from the host").UintVar(&c.Input.ExpectedResponse)
	c.CmdClause.Flag("window", "The number of most recent healthcheck queries to keep for this healthcheck").UintVar(&c.Input.Window)
	c.CmdClause.Flag("threshold", "How many healthchecks must succeed to be considered healthy").UintVar(&c.Input.Threshold)
	c.CmdClause.Flag("initial", "When loading a config, the initial number of probes to be seen as OK").UintVar(&c.Input.Initial)
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	h, err := c.Globals.Client.CreateHealthCheck(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created healthcheck %s (service %s version %d)", h.Name, h.ServiceID, h.ServiceVersion)
	return nil
}
