package custom

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List the uploaded VCLs for a particular service and version")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})

	// Optional Flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	manifest       manifest.Data
	serviceVersion cmd.OptionalServiceVersion
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
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

	vs, err := c.Globals.Client.ListVCLs(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, serviceID, serviceVersion.Number, vs)
	} else {
		c.printSummary(out, vs)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput(serviceID string, serviceVersion int) *fastly.ListVCLsInput {
	var input fastly.ListVCLsInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, serviceID string, serviceVersion int, vs []*fastly.VCL) {
	fmt.Fprintf(out, "\nService ID: %s\n", serviceID)
	fmt.Fprintf(out, "Service Version: %d\n", serviceVersion)

	for _, v := range vs {
		fmt.Fprintf(out, "\nName: %s\n", v.Name)
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
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, vs []*fastly.VCL) {
	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "VERSION", "NAME", "MAIN")
	for _, v := range vs {
		t.AddLine(v.ServiceID, v.ServiceVersion, v.Name, v.Main)
	}
	t.Print()
}
