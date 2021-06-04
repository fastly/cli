package honeycomb

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a Honeycomb logging endpoint.
type UpdateCommand struct {
	cmd.Base
	manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	serviceVersion cmd.OptionalServiceVersion

	// optional
	autoClone         cmd.OptionalAutoClone
	NewName           cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalUint
	Dataset           cmd.OptionalString
	Token             cmd.OptionalString
	ResponseCondition cmd.OptionalString
	Placement         cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a Honeycomb logging endpoint on a Fastly service version")
	c.NewServiceVersionFlag(cmd.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	c.NewAutoCloneFlag(c.autoClone.Set, &c.autoClone.Value)
	c.CmdClause.Flag("name", "The name of the Honeycomb logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("new-name", "New name of the Honeycomb logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("format", "Apache style log formatting. Your log must produce valid JSON that Honeycomb can ingest").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("dataset", "The Honeycomb Dataset you want to log to").Action(c.Dataset.Set).StringVar(&c.Dataset.Value)
	c.CmdClause.Flag("auth-token", "The Write Key from the Account page of your Honeycomb account").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateHoneycombInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	v, err := c.serviceVersion.Parse(serviceID, c.Globals.Client)
	if err != nil {
		return nil, err
	}
	v, err = c.autoClone.Parse(v, serviceID, c.Globals.Client)
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateHoneycombInput{
		ServiceID:      serviceID,
		ServiceVersion: v.Number,
		Name:           c.EndpointName,
	}

	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.Token.WasSet {
		input.Token = fastly.String(c.Token.Value)
	}

	if c.Dataset.WasSet {
		input.Dataset = fastly.String(c.Dataset.Value)
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
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	honeycomb, err := c.Globals.Client.UpdateHoneycomb(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Honeycomb logging endpoint %s (service %s version %d)", honeycomb.Name, honeycomb.ServiceID, honeycomb.ServiceVersion)
	return nil
}
