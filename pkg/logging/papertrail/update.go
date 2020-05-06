package papertrail

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Papertrail logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	Input fastly.GetPapertrailInput

	NewName common.OptionalString

	Address           common.OptionalString
	Port              common.OptionalUint
	FormatVersion     common.OptionalUint
	Format            common.OptionalString
	ResponseCondition common.OptionalString
	Placement         common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Papertrail logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the Papertrail logging object").Short('n').Required().StringVar(&c.Input.Name)

	c.CmdClause.Flag("new-name", "New name of the Papertrail logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("address", "A hostname or IPv4 address").Action(c.Address.Set).StringVar(&c.Address.Value)
	c.CmdClause.Flag("port", "The port number").Action(c.Port.Set).UintVar(&c.Port.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("format", "Apache style log formatting").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Placement.Set).StringVar(&c.Placement.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	papertrail, err := c.Globals.Client.GetPapertrail(&c.Input)
	if err != nil {
		return err
	}

	input := &fastly.UpdatePapertrailInput{
		Service:           papertrail.ServiceID,
		Version:           papertrail.Version,
		Name:              papertrail.Name,
		NewName:           papertrail.Name,
		Address:           papertrail.Address,
		Port:              papertrail.Port,
		FormatVersion:     papertrail.FormatVersion,
		Format:            papertrail.Format,
		ResponseCondition: papertrail.ResponseCondition,
		Placement:         papertrail.Placement,
	}

	// Set new values if set by user.
	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.NewName.Valid {
		input.NewName = c.NewName.Value
	}

	if c.Address.Valid {
		input.Address = c.Address.Value
	}

	if c.Port.Valid {
		input.Port = c.Port.Value
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = c.FormatVersion.Value
	}

	if c.Format.Valid {
		input.Format = c.Format.Value
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = c.ResponseCondition.Value
	}

	if c.Placement.Valid {
		input.Placement = c.Placement.Value
	}

	papertrail, err = c.Globals.Client.UpdatePapertrail(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Papertrail logging endpoint %s (service %s version %d)", papertrail.Name, papertrail.ServiceID, papertrail.Version)
	return nil
}
