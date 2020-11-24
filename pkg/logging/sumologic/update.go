package sumologic

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to update Sumologic logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	URL               common.OptionalString
	Format            common.OptionalString
	ResponseCondition common.OptionalString
	MessageType       common.OptionalString
	FormatVersion     common.OptionalInt // Inconsistent with other logging endpoints, but remaining as int to avoid breaking changes in fastly/go-fastly.
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Sumologic logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Sumologic logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Sumologic logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("url", "The URL to POST to").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).IntVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.MessageType.Set).StringVar(&c.MessageType.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateSumologicInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	sumologic, err := c.Globals.Client.GetSumologic(&fastly.GetSumologicInput{
		ServiceID:      serviceID,
		Name:           c.EndpointName,
		ServiceVersion: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateSumologicInput{
		ServiceID:         sumologic.ServiceID,
		ServiceVersion:    sumologic.ServiceVersion,
		Name:              sumologic.Name,
		NewName:           fastly.String(sumologic.Name),
		URL:               fastly.String(sumologic.URL),
		Format:            fastly.String(sumologic.Format),
		ResponseCondition: fastly.String(sumologic.ResponseCondition),
		MessageType:       fastly.String(sumologic.MessageType),
		FormatVersion:     fastly.Int(sumologic.FormatVersion),
		Placement:         fastly.String(sumologic.Placement),
	}

	// Set new values if set by user.
	if c.NewName.WasSet {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.URL.WasSet {
		input.URL = fastly.String(c.URL.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.MessageType.WasSet {
		input.MessageType = fastly.String(c.MessageType.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Int(c.FormatVersion.Value)
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
	sumologic, err := c.Globals.Client.UpdateSumologic(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Sumologic logging endpoint %s (service %s version %d)", sumologic.Name, sumologic.ServiceID, sumologic.ServiceVersion)
	return nil
}
