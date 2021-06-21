package custom

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update the uploaded VCL for a particular service and version")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("name", "The name of the VCL to update").Required().StringVar(&c.name)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional flags
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("new-name", "New name for the VCL").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("content", "VCL passed as file path or content, e.g. $(cat main.vcl)").Action(c.content.Set).StringVar(&c.content.Value)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	autoClone      cmd.OptionalAutoClone
	content        cmd.OptionalString
	manifest       manifest.Data
	name           string
	newName        cmd.OptionalString
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		return err
	}

	input, err := c.createInput(serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}

	v, err := c.Globals.Client.UpdateVCL(input)
	if err != nil {
		return err
	}

	if input.NewName != nil && *input.NewName != "" {
		text.Success(out, "Updated custom VCL '%s' (previously: '%s', service: %s, version: %d)", v.Name, input.Name, v.ServiceID, v.ServiceVersion)
	} else {
		text.Success(out, "Updated custom VCL '%s' (service: %s, version: %d)", v.Name, v.ServiceID, v.ServiceVersion)
	}
	return nil
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput(serviceID string, serviceVersion int) (*fastly.UpdateVCLInput, error) {
	var input fastly.UpdateVCLInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if !c.newName.WasSet && !c.content.WasSet {
		return nil, fmt.Errorf("error parsing arguments: must provide either --new-name or --content to update the VCL")
	}
	if c.newName.WasSet {
		input.NewName = fastly.String(c.newName.Value)
	}
	if c.content.WasSet {
		input.Content = fastly.String(Content(c.content.Value))
	}

	return &input, nil
}

// Content determines if the given flag value for --content is a file path,
// and if so read the contents from disk, otherwise presume the given value is
// the content.
func Content(flagval string) string {
	content := flagval
	if path, err := filepath.Abs(flagval); err == nil {
		if _, err := os.Stat(path); err == nil {
			if data, err := os.ReadFile(path); err == nil {
				content = string(data)
			}
		}
	}
	return content
}
