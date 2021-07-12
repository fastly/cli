package snippet

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create a snippet for a particular service and version").Alias("add")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("content", "VCL snippet passed as file path or content, e.g. $(cat snippet.vcl)").Required().StringVar(&c.content)
	c.CmdClause.Flag("name", "The name of the VCL snippet").Required().StringVar(&c.name)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("type", "The location in generated VCL where the snippet should be placed (e.g. recv, miss, fetch etc)").Required().StringVar(&c.location)

	// Optional flags
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.CmdClause.Flag("priority", "Priority determines execution order. Lower numbers execute first").Short('p').Action(c.priority.Set).IntVar(&c.priority.Value)

	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	autoClone      cmd.OptionalAutoClone
	content        string
	dynamic        cmd.OptionalBool
	location       string
	manifest       manifest.Data
	name           string
	priority       cmd.OptionalInt
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	input := c.constructInput(serviceID, serviceVersion.Number)

	v, err := c.Globals.Client.CreateSnippet(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created VCL snippet '%s' (service: %s, version: %d, dynamic: %t, type: %s, priority: %d)", v.Name, v.ServiceID, v.ServiceVersion, c.dynamic.WasSet, c.location, v.Priority)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string, serviceVersion int) *fastly.CreateSnippetInput {
	var input fastly.CreateSnippetInput

	input.Content = cmd.Content(c.content)
	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Type = fastly.SnippetType(c.location)

	if c.dynamic.WasSet {
		input.Dynamic = 1
	}
	if c.priority.WasSet {
		input.Priority = c.priority.Value
	}

	return &input
}
