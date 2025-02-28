package scalyr

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Scalyr logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListScalyrsInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Scalyr endpoints on a Fastly service version")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.ListScalyrs(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, scalyr := range o {
			tw.AddLine(
				fastly.ToValue(scalyr.ServiceID),
				fastly.ToValue(scalyr.ServiceVersion),
				fastly.ToValue(scalyr.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, scalyr := range o {
		fmt.Fprintf(out, "\tScalyr %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(scalyr.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(scalyr.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(scalyr.Name))
		fmt.Fprintf(out, "\t\tToken: %s\n", fastly.ToValue(scalyr.Token))
		fmt.Fprintf(out, "\t\tRegion: %s\n", fastly.ToValue(scalyr.Region))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(scalyr.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(scalyr.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(scalyr.ResponseCondition))
		fmt.Fprintf(out, "\t\tProject ID: %s\n", fastly.ToValue(scalyr.ProjectID))
	}
	fmt.Fprintln(out)

	return nil
}
