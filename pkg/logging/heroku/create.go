package heroku

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// CreateCommand calls the Fastly API to create a Heroku logging endpoint.
type CreateCommand struct {
	cmd.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shadow cmd.Base method Name().
	Version      int
	Token        string
	URL          string

	// optional
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalUint
	ResponseCondition cmd.OptionalString
	Placement         cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand

	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("create", "Create a Heroku logging endpoint on a Fastly service version").Alias("add")

	c.CmdClause.Flag("name", "The name of the Heroku logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("url", "The url to stream logs to").Required().StringVar(&c.URL)
	c.CmdClause.Flag("auth-token", "The token to use for authentication (https://devcenter.heroku.com/articles/add-on-partner-log-integration)").Required().StringVar(&c.Token)

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) createInput() (*fastly.CreateHerokuInput, error) {
	var input fastly.CreateHerokuInput

	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	input.ServiceID = serviceID
	input.ServiceVersion = c.Version
	input.Name = c.EndpointName
	input.Token = c.Token
	input.URL = c.URL

	if c.Format.WasSet {
		input.Format = c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.Placement.WasSet {
		input.Placement = c.Placement.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	d, err := c.Globals.Client.CreateHeroku(input)
	if err != nil {
		return err
	}

	text.Success(out, "Created Heroku logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
