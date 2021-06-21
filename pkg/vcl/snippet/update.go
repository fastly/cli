package snippet

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a VCL snippet for a particular service (and version, if not a dynamic snippet)")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("name", "The name of the VCL snippet to update").Required().StringVar(&c.name)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional flags
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("content", "VCL snippet passed as file path or content, e.g. $(cat snippet.vcl)").Action(c.content.Set).StringVar(&c.content.Value)
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.CmdClause.Flag("new-name", "New name for the VCL snippet").Action(c.newName.Set).StringVar(&c.newName.Value)
	c.CmdClause.Flag("priority", "Priority determines execution order. Lower numbers execute first").Short('p').IntVar(&c.priority)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("snippet-id", "Alphanumeric string identifying a VCL Snippet").Short('i').StringVar(&c.snippetID)
	c.CmdClause.Flag("type", "The location in generated VCL where the snippet should be placed (e.g. recv, miss, fetch etc)").StringVar(&c.location)

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	autoClone      cmd.OptionalAutoClone
	content        cmd.OptionalString
	dynamic        cmd.OptionalBool
	location       string
	manifest       manifest.Data
	name           string
	newName        cmd.OptionalString
	priority       int
	serviceVersion cmd.OptionalServiceVersion
	snippetID      string
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

	if c.dynamic.WasSet {
		input, err := c.constructDynamicInput(serviceID, serviceVersion.Number)
		if err != nil {
			return err
		}
		v, err := c.Globals.Client.UpdateDynamicSnippet(input)
		if err != nil {
			return err
		}
		text.Success(out, "Updated VCL snippet '%s' (service: %s)", v.ID, v.ServiceID)
		return nil
	}

	input, err := c.constructInput(serviceID, serviceVersion.Number)
	if err != nil {
		return err
	}
	v, err := c.Globals.Client.UpdateSnippet(input)
	if err != nil {
		return err
	}
	if input.NewName != "" {
		text.Success(out, "Updated VCL snippet '%s' (previously: '%s', service: %s, version: %d)", v.Name, input.Name, v.ServiceID, v.ServiceVersion)
	} else {
		text.Success(out, "Updated VCL snippet '%s' (service: %s, version: %d)", v.Name, v.ServiceID, v.ServiceVersion)
	}
	return nil
}

// constructDynamicInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructDynamicInput(serviceID string, serviceVersion int) (*fastly.UpdateDynamicSnippetInput, error) {
	var input fastly.UpdateDynamicSnippetInput

	input.ServiceID = serviceID
	input.ID = c.snippetID

	if c.content.WasSet {
		input.Content = cmd.Content(c.content.Value)
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput(serviceID string, serviceVersion int) (*fastly.UpdateSnippetInput, error) {
	var input fastly.UpdateSnippetInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if !c.newName.WasSet && !c.content.WasSet {
		return nil, fmt.Errorf("error parsing arguments: must provide either --new-name or --content to update the VCL snippet")
	}
	if c.newName.WasSet {
		input.NewName = c.newName.Value
	}
	if c.content.WasSet {
		input.Content = cmd.Content(c.content.Value)
	}

	return &input, nil
}
