package serviceresource

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// UpdateCommand calls the Fastly API to update a dictionary.
type UpdateCommand struct {
	cmd.Base
	jsonOutput

	autoClone      cmd.OptionalAutoClone
	input          fastly.UpdateResourceInput
	manifest       manifest.Data
	serviceVersion cmd.OptionalServiceVersion
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *global.Data, data manifest.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
		input: fastly.UpdateResourceInput{
			// Kingpin requires the following to be initialized.
			Name: new(string),
		},
	}
	c.CmdClause = parent.Command("update", "Update a resource link for a Fastly service version")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "id",
		Description: "ID of resource link",
		Dst:         &c.input.ID,
		Required:    true,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: "Resource name (alias)",
		Dst:         c.input.Name,
		Required:    true,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Short:       's',
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Required:    true,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    cmd.OptionalServiceNameID{}, // ServiceID flag is required, no need to lookup service by name.
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      c.manifest.Flag.ServiceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersion.Number

	resource, err := c.Globals.APIClient.UpdateResource(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"ID":              c.input.ID,
			"Service ID":      c.input.ServiceID,
			"Service Version": c.input.ServiceVersion,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, resource); ok {
		return err
	}

	text.Success(out, "Updated service resource link %s on service %s version %s", resource.ID, resource.ServiceID, resource.ServiceVersion)
	return nil
}
