package loggly

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Loggly logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	Token             common.OptionalString
	ResponseCondition common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Loggly logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Loggly logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Loggly logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("auth-token", "The token to use for authentication (https://www.loggly.com/docs/customer-token-authentication-token/)").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateLogglyInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	loggly, err := c.Globals.Client.GetLoggly(&fastly.GetLogglyInput{
		Service: serviceID,
		Name:    c.EndpointName,
		Version: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateLogglyInput{
		Service:           loggly.ServiceID,
		Version:           loggly.Version,
		Name:              loggly.Name,
		NewName:           fastly.String(loggly.Name),
		Format:            fastly.String(loggly.Format),
		FormatVersion:     fastly.Uint(loggly.FormatVersion),
		Token:             fastly.String(loggly.Token),
		ResponseCondition: fastly.String(loggly.ResponseCondition),
		Placement:         fastly.String(loggly.Placement),
	}

	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.Token.Valid {
		input.Token = fastly.String(c.Token.Value)
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.Valid {
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

	loggly, err := c.Globals.Client.UpdateLoggly(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Loggly logging endpoint %s (service %s version %d)", loggly.Name, loggly.ServiceID, loggly.Version)
	return nil
}
