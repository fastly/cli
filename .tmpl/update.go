package ${CLI_PACKAGE}

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v4/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "<...>")
	c.Globals = globals
	c.manifest = data

	// Required flags
	// c.CmdClause.Flag("name", "<...>").Required().StringVar(&c.name)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional flags
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("new-name", "<...>").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	autoClone      cmd.OptionalAutoClone
	manifest manifest.Data
	name           string
	newName        cmd.OptionalString
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.constructInput(serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}

	r, err := c.Globals.Client.Update${CLI_API}(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if input.NewName != nil && *input.NewName != "" {
		text.Success(out, "Updated <...> '%s' (previously: '%s', service: %s, version: %d)", r.Name, input.Name, r.ServiceID, r.ServiceVersion)
	} else {
		text.Success(out, "Updated <...> '%s' (service: %s, version: %d)", r.Name, r.ServiceID, r.ServiceVersion)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.Update${CLI_API}Input, error) {
	var input fastly.Update${CLI_API}Input

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if !c.newName.WasSet && !c.content.WasSet {
		return nil, fmt.Errorf("error parsing arguments: must provide either --new-name or --content to update the <...>")
	}
	if c.newName.WasSet {
		input.NewName = fastly.String(c.newName.Value)
	}

	return &input, nil

}
