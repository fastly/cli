package openstack

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// ListCommand calls the Fastly API to list OpenStack logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListOpenstackInput
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
	c.CmdClause = parent.Command("list", "List OpenStack logging endpoints on a Fastly service version")

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
		Description: argparser.FlagServiceDesc,
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
		AllowActiveLocked:  true,
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

	o, err := c.Globals.APIClient.ListOpenstack(&c.Input)
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
		for _, openstack := range o {
			tw.AddLine(
				fastly.ToValue(openstack.ServiceID),
				fastly.ToValue(openstack.ServiceVersion),
				fastly.ToValue(openstack.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, openstack := range o {
		fmt.Fprintf(out, "\tOpenstack %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(openstack.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(openstack.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(openstack.Name))
		fmt.Fprintf(out, "\t\tBucket: %s\n", fastly.ToValue(openstack.BucketName))
		fmt.Fprintf(out, "\t\tAccess key: %s\n", fastly.ToValue(openstack.AccessKey))
		fmt.Fprintf(out, "\t\tUser: %s\n", fastly.ToValue(openstack.User))
		fmt.Fprintf(out, "\t\tURL: %s\n", fastly.ToValue(openstack.URL))
		fmt.Fprintf(out, "\t\tPath: %s\n", fastly.ToValue(openstack.Path))
		fmt.Fprintf(out, "\t\tPeriod: %d\n", fastly.ToValue(openstack.Period))
		fmt.Fprintf(out, "\t\tGZip level: %d\n", fastly.ToValue(openstack.GzipLevel))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(openstack.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(openstack.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(openstack.ResponseCondition))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(openstack.MessageType))
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", fastly.ToValue(openstack.TimestampFormat))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(openstack.Placement))
		fmt.Fprintf(out, "\t\tPublic key: %s\n", fastly.ToValue(openstack.PublicKey))
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", fastly.ToValue(openstack.CompressionCodec))
	}
	fmt.Fprintln(out)

	return nil
}
