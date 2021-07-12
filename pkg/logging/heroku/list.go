package heroku

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Heroku logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListHerokusInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Heroku endpoints on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	herokus, err := c.Globals.Client.ListHerokus(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, heroku := range herokus {
			tw.AddLine(heroku.ServiceID, heroku.ServiceVersion, heroku.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, heroku := range herokus {
		fmt.Fprintf(out, "\tHeroku %d/%d\n", i+1, len(herokus))
		fmt.Fprintf(out, "\t\tService ID: %s\n", heroku.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", heroku.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", heroku.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", heroku.URL)
		fmt.Fprintf(out, "\t\tToken: %s\n", heroku.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", heroku.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", heroku.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", heroku.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", heroku.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
