package resourcelink

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// CreateCommand calls the Fastly API to create a resource link.
type CreateCommand struct {
	cmd.Base
	cmd.JSONOutput

	autoClone      cmd.OptionalAutoClone
	input          fastly.CreateResourceInput
	manifest       manifest.Data
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *CreateCommand {
	c := CreateCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
		input: fastly.CreateResourceInput{
			// Kingpin requires the following to be initialized.
			ResourceID: new(string),
			Name:       new(string),
		},
	}
	c.CmdClause = parent.Command("create", "Create a Fastly service resource link").Alias("link")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "resource-id",
		Short:       'r',
		Description: flagResourceIDDescription,
		Dst:         c.input.ResourceID,
		Required:    true,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// At least one of the following is required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Short:       's',
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceName,
		Action:      c.serviceName.Set,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})

	// Optional.
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        "name",
		Short:       'n',
		Description: flagNameDescription,
		Dst:         c.input.Name,
	})

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
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

	resource, err := c.Globals.APIClient.CreateResource(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"ID":              c.input.ResourceID,
			"Service ID":      c.input.ServiceID,
			"Service Version": c.input.ServiceVersion,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, resource); ok {
		return err
	}

	text.Success(out, "Created service resource link %q (%s) on service %s version %s", resource.Name, resource.ID, resource.ServiceID, resource.ServiceVersion)
	return nil
}
