package kinesis

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Amazon Kinesis logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListKinesisInput
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
	c.CmdClause = parent.Command("list", "List Kinesis endpoints on a Fastly service version")

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

	o, err := c.Globals.APIClient.ListKinesis(&c.Input)
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
		for _, kinesis := range o {
			tw.AddLine(
				fastly.ToValue(kinesis.ServiceID),
				fastly.ToValue(kinesis.ServiceVersion),
				fastly.ToValue(kinesis.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kinesis := range o {
		fmt.Fprintf(out, "\tKinesis %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(kinesis.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(kinesis.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(kinesis.Name))
		fmt.Fprintf(out, "\t\tStream name: %s\n", fastly.ToValue(kinesis.StreamName))
		fmt.Fprintf(out, "\t\tRegion: %s\n", fastly.ToValue(kinesis.Region))
		if kinesis.AccessKey != nil || kinesis.SecretKey != nil {
			fmt.Fprintf(out, "\t\tAccess key: %s\n", fastly.ToValue(kinesis.AccessKey))
			fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(kinesis.SecretKey))
		}
		if kinesis.IAMRole != nil {
			fmt.Fprintf(out, "\t\tIAM role: %s\n", fastly.ToValue(kinesis.IAMRole))
		}
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(kinesis.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(kinesis.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(kinesis.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(kinesis.Placement))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(kinesis.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
