package newrelic

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v4/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all of the New Relic Logs logging objects for a particular service and version")
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
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input := c.constructInput(serviceID, serviceVersion.Number)

	l, err := c.Globals.Client.ListNewRelic(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, serviceVersion.Number, l)
	} else {
		c.printSummary(out, l)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput(serviceID string, serviceVersion int) *fastly.ListNewRelicInput {
	var input fastly.ListNewRelicInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, serviceVersion int, ls []*fastly.NewRelic) {
	fmt.Fprintf(out, "Service Version: %d\n", serviceVersion)

	for _, l := range ls {
		fmt.Fprintf(out, "\nName: %s\n", l.Name)
		fmt.Fprintf(out, "\nToken: %s\n", l.Token)
		fmt.Fprintf(out, "\nFormat: %s\n", l.Format)
		fmt.Fprintf(out, "\nFormat Version: %d\n", l.FormatVersion)
		fmt.Fprintf(out, "\nPlacement: %s\n", l.Placement)
		fmt.Fprintf(out, "\nRegion: %s\n", l.Region)
		fmt.Fprintf(out, "\nResponse Condition: %s\n\n", l.ResponseCondition)

		if l.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", l.CreatedAt)
		}
		if l.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", l.UpdatedAt)
		}
		if l.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", l.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, ls []*fastly.NewRelic) {
	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "VERSION", "NAME")
	for _, l := range ls {
		t.AddLine(l.ServiceID, l.ServiceVersion, l.Name)
	}
	t.Print()
}
