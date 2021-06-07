package loggly

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a Loggly logging endpoint.
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
	c.CmdClause = parent.Command("update", "Update a Loggly logging endpoint on a Fastly service version")
	c.SetServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.SetAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("name", "The name of the Loggly logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("new-name", "New name of the Loggly logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("auth-token", "The token to use for authentication (https://www.loggly.com/docs/customer-token-authentication-token/)").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput(serviceID string, serviceVersion int) (*fastly.UpdateLogglyInput, error) {
	input := fastly.UpdateLogglyInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
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
	serviceID, serviceVersion, err := cmd.ServiceDetails(c.manifest, c.serviceVersion, c.autoClone, c.Globals.Flag.Verbose, out, c.Globals.Client)
	if err != nil {
		return err
	}

	input, err := c.createInput(serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}

	loggly, err := c.Globals.Client.UpdateLoggly(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Loggly logging endpoint %s (service %s version %d)", loggly.Name, loggly.ServiceID, loggly.ServiceVersion)
	return nil
}
