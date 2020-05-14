package logshuttle

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Logshuttle logging endpoints.
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
	URL               common.OptionalString
	ResponseCondition common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Logshuttle logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Logshuttle logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Logshuttle logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)

	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("url", "Your Log Shuttle endpoint url").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("auth-token", "The data authentication token associated with this endpoint").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateLogshuttleInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	logshuttle, err := c.Globals.Client.GetLogshuttle(&fastly.GetLogshuttleInput{
		Service: serviceID,
		Name:    c.EndpointName,
		Version: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateLogshuttleInput{
		Service:           logshuttle.ServiceID,
		Version:           logshuttle.Version,
		Name:              logshuttle.Name,
		NewName:           fastly.String(logshuttle.Name),
		Format:            fastly.String(logshuttle.Format),
		FormatVersion:     fastly.Uint(logshuttle.FormatVersion),
		URL:               fastly.String(logshuttle.URL),
		Token:             fastly.String(logshuttle.Token),
		ResponseCondition: fastly.String(logshuttle.ResponseCondition),
		Placement:         fastly.String(logshuttle.Placement),
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.URL.Valid {
		input.URL = fastly.String(c.URL.Value)
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

	logshuttle, err := c.Globals.Client.UpdateLogshuttle(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Logshuttle logging endpoint %s (service %s version %d)", logshuttle.Name, logshuttle.ServiceID, logshuttle.Version)
	return nil
}
