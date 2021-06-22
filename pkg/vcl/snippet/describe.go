package snippet

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Get the uploaded VCL for a particular service and version").Alias("get")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("name", "The name of the VCL").Required().StringVar(&c.name)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("snippet-id", "Alphanumeric string identifying a VCL Snippet").Short('i').Required().StringVar(&c.snippetID)

	// Optional Flags
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	dynamic        cmd.OptionalBool
	manifest       manifest.Data
	name           string
	serviceVersion cmd.OptionalServiceVersion
	snippetID      string
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
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
		input := c.constructDynamicInput(serviceID, serviceVersion.Number)
		v, err := c.Globals.Client.GetDynamicSnippet(input)
		if err != nil {
			return err
		}
		c.printDynamic(out, v)
		return nil
	}

	input := c.constructInput(serviceID, serviceVersion.Number)
	v, err := c.Globals.Client.GetSnippet(input)
	if err != nil {
		return err
	}
	c.print(out, v)
	return nil
}

// constructDynamicInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructDynamicInput(serviceID string, serviceVersion int) *fastly.GetDynamicSnippetInput {
	var input fastly.GetDynamicSnippetInput

	input.ID = c.snippetID
	input.ServiceID = serviceID

	return &input
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) *fastly.GetSnippetInput {
	var input fastly.GetSnippetInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// print displays the 'dynamic' information returned from the API.
func (c *DescribeCommand) printDynamic(out io.Writer, v *fastly.DynamicSnippet) {
	fmt.Fprintf(out, "Service ID: %s\n", v.ServiceID)
	fmt.Fprintf(out, "ID: %s\n", v.ID)
	fmt.Fprintf(out, "Content: %s\n", v.Content)
	if v.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", v.CreatedAt)
	}
	if v.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", v.UpdatedAt)
	}
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, v *fastly.Snippet) {
	fmt.Fprintf(out, "Service ID: %s\n", v.ServiceID)
	fmt.Fprintf(out, "Service Version: %d\n", v.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", v.Name)
	fmt.Fprintf(out, "ID: %s\n", v.ID)
	fmt.Fprintf(out, "Priority: %d\n", v.Priority)
	fmt.Fprintf(out, "Dynamic: %t\n", cmd.IntToBool(v.Dynamic))
	fmt.Fprintf(out, "Type: %s\n", v.Type)
	fmt.Fprintf(out, "Content: %s\n", v.Content)
	if v.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", v.CreatedAt)
	}
	if v.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", v.UpdatedAt)
	}
	if v.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", v.DeletedAt)
	}
}
