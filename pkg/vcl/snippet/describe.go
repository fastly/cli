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
	c.CmdClause = parent.Command("describe", "Get the uploaded VCL snippet for a particular service and version").Alias("get")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional Flags
	c.CmdClause.Flag("dynamic", "Whether the VCL snippet is dynamic or versioned").Action(c.dynamic.Set).BoolVar(&c.dynamic.Value)
	c.CmdClause.Flag("name", "The name of the VCL snippet").StringVar(&c.name)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("snippet-id", "Alphanumeric string identifying a VCL Snippet").Short('i').StringVar(&c.snippetID)

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
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.dynamic.WasSet {
		input, err := c.constructDynamicInput(serviceID, serviceVersion.Number)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		v, err := c.Globals.Client.GetDynamicSnippet(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.printDynamic(out, v)
		return nil
	}

	input, err := c.constructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	v, err := c.Globals.Client.GetSnippet(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	c.print(out, v)
	return nil
}

// constructDynamicInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructDynamicInput(serviceID string, serviceVersion int) (*fastly.GetDynamicSnippetInput, error) {
	var input fastly.GetDynamicSnippetInput

	input.ID = c.snippetID
	input.ServiceID = serviceID

	if c.snippetID == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --snippet-id with a dynamic VCL snippet")
	}

	return &input, nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) (*fastly.GetSnippetInput, error) {
	var input fastly.GetSnippetInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.name == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --name with a versioned VCL snippet")
	}

	return &input, nil
}

// print displays the 'dynamic' information returned from the API.
func (c *DescribeCommand) printDynamic(out io.Writer, v *fastly.DynamicSnippet) {
	fmt.Fprintf(out, "\nService ID: %s\n", v.ServiceID)
	fmt.Fprintf(out, "ID: %s\n", v.ID)
	fmt.Fprintf(out, "Content: \n%s\n", v.Content)
	if v.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", v.CreatedAt)
	}
	if v.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", v.UpdatedAt)
	}
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, v *fastly.Snippet) {
	fmt.Fprintf(out, "\nService ID: %s\n", v.ServiceID)
	fmt.Fprintf(out, "Service Version: %d\n", v.ServiceVersion)

	fmt.Fprintf(out, "\nName: %s\n", v.Name)
	fmt.Fprintf(out, "ID: %s\n", v.ID)
	fmt.Fprintf(out, "Priority: %d\n", v.Priority)
	fmt.Fprintf(out, "Dynamic: %t\n", cmd.IntToBool(v.Dynamic))
	fmt.Fprintf(out, "Type: %s\n", v.Type)
	fmt.Fprintf(out, "Content: \n%s\n", v.Content)
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
