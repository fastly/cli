package custom

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

	// Optional Flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	manifest       manifest.Data
	name           string
	serviceVersion cmd.OptionalServiceVersion
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

	input := c.constructInput(serviceID, serviceVersion.Number)

	v, err := c.Globals.Client.GetVCL(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.print(out, v)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) *fastly.GetVCLInput {
	var input fastly.GetVCLInput

	input.Name = c.name
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, v *fastly.VCL) {
	fmt.Fprintf(out, "\nService ID: %s\n", v.ServiceID)
	fmt.Fprintf(out, "Service Version: %d\n\n", v.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", v.Name)
	fmt.Fprintf(out, "Main: %t\n", v.Main)
	fmt.Fprintf(out, "Content: \n%s\n\n", v.Content)
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
