package healthcheck

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update healthchecks.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetHealthCheckInput

	NewName          common.OptionalString
	Comment          common.OptionalString
	Method           common.OptionalString
	Host             common.OptionalString
	Path             common.OptionalString
	HTTPVersion      common.OptionalString
	Timeout          common.OptionalUint
	CheckInterval    common.OptionalUint
	ExpectedResponse common.OptionalUint
	Window           common.OptionalUint
	Threshold        common.OptionalUint
	Initial          common.OptionalUint
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a healthcheck on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "Healthcheck name").Short('n').Required().StringVar(&c.Input.Name)

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
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	h, err := c.Globals.Client.GetHealthCheck(&c.Input)
	if err != nil {
		return err
	}

	// Copy existing values from GET to UpdateHealthCheckInput strcuture
	input := &fastly.UpdateHealthCheckInput{
		Service:          h.ServiceID,
		Version:          h.Version,
		Name:             h.Name,
		NewName:          h.Name,
		Comment:          h.Comment,
		Method:           h.Method,
		Host:             h.Host,
		Path:             h.Path,
		HTTPVersion:      h.HTTPVersion,
		Timeout:          h.Timeout,
		CheckInterval:    h.CheckInterval,
		ExpectedResponse: h.ExpectedResponse,
		Window:           h.Window,
		Threshold:        h.Threshold,
		Initial:          h.Initial,
	}

	// Set values to existing ones to prevent accidental overwrite if empty.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.Comment.Valid {
		input.Comment = c.Comment.Value
	}

	if c.Method.Valid {
		input.Method = c.Method.Value
	}

	if c.Host.Valid {
		input.Host = c.Host.Value
	}

	if c.Path.Valid {
		input.Path = c.Path.Value
	}

	if c.HTTPVersion.Valid {
		input.HTTPVersion = c.HTTPVersion.Value
	}

	if c.Timeout.Valid {
		input.Timeout = c.Timeout.Value
	}

	if c.CheckInterval.Valid {
		input.CheckInterval = c.CheckInterval.Value
	}

	if c.ExpectedResponse.Valid {
		input.ExpectedResponse = c.ExpectedResponse.Value
	}

	if c.Window.Valid {
		input.Window = c.Window.Value
	}

	if c.Threshold.Valid {
		input.Threshold = c.Threshold.Value
	}

	if c.Initial.Valid {
		input.Initial = c.Initial.Value
	}

	h, err = c.Globals.Client.UpdateHealthCheck(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated healthcheck %s (service %s version %d)", h.Name, h.ServiceID, h.Version)
	return nil
}
