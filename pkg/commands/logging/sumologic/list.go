package sumologic

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// ListCommand calls the Fastly API to list Sumologic logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListSumologicsInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Sumologic endpoints on a Fastly service version")
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
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
		ServiceNameFlag:    c.serviceName,
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	sumologics, err := c.Globals.Client.ListSumologics(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, sumologic := range sumologics {
			tw.AddLine(sumologic.ServiceID, sumologic.ServiceVersion, sumologic.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, sumologic := range sumologics {
		fmt.Fprintf(out, "\tSumologic %d/%d\n", i+1, len(sumologics))
		fmt.Fprintf(out, "\t\tService ID: %s\n", sumologic.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", sumologic.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", sumologic.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", sumologic.URL)
		fmt.Fprintf(out, "\t\tFormat: %s\n", sumologic.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", sumologic.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", sumologic.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", sumologic.MessageType)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", sumologic.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
