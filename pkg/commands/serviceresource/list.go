package serviceresource

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// ListCommand calls the Fastly API to list service resource links
type ListCommand struct {
	cmd.Base
	jsonOutput

	input          fastly.ListResourcesInput
	manifest       manifest.Data
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent
func NewListCommand(parent cmd.Registerer, globals *global.Data, data manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("list", "List all resource links for a Fastly service version")

	// Required.
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
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := cmd.ServiceID(cmd.OptionalServiceNameID{}, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	serviceVersion, err := c.serviceVersion.Parse(serviceID, c.Globals.APIClient)
	if err != nil {
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = serviceVersion.Number

	resources, err := c.Globals.APIClient.ListResources(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      c.input.ServiceID,
			"Service Version": c.input.ServiceVersion,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, resources); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "Service ID: %s\n", c.input.ServiceID)
	}
	text.Output(out, "Service Version: %d\n", c.input.ServiceVersion)

	for i, resource := range resources {
		fmt.Fprintf(out, "Resource Link %d/%d\n", i+1, len(resources))
		text.PrintResource(out, "\t", resource)
		fmt.Fprintln(out)
	}

	return nil
}
