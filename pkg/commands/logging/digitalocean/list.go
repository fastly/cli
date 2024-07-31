package digitalocean

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list DigitalOcean Spaces logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListDigitalOceansInput
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
	c.CmdClause = parent.Command("list", "List DigitalOcean Spaces logging endpoints on a Fastly service version")

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

	o, err := c.Globals.APIClient.ListDigitalOceans(&c.Input)
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
		for _, digitalocean := range o {
			tw.AddLine(
				fastly.ToValue(digitalocean.ServiceID),
				fastly.ToValue(digitalocean.ServiceVersion),
				fastly.ToValue(digitalocean.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, digitalocean := range o {
		fmt.Fprintf(out, "\tDigitalOcean %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(digitalocean.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(digitalocean.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(digitalocean.Name))
		fmt.Fprintf(out, "\t\tBucket: %s\n", fastly.ToValue(digitalocean.BucketName))
		fmt.Fprintf(out, "\t\tDomain: %s\n", fastly.ToValue(digitalocean.Domain))
		fmt.Fprintf(out, "\t\tAccess key: %s\n", fastly.ToValue(digitalocean.AccessKey))
		fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(digitalocean.SecretKey))
		fmt.Fprintf(out, "\t\tPath: %s\n", fastly.ToValue(digitalocean.Path))
		fmt.Fprintf(out, "\t\tPeriod: %d\n", fastly.ToValue(digitalocean.Period))
		fmt.Fprintf(out, "\t\tGZip level: %d\n", fastly.ToValue(digitalocean.GzipLevel))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(digitalocean.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(digitalocean.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(digitalocean.ResponseCondition))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(digitalocean.MessageType))
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", fastly.ToValue(digitalocean.TimestampFormat))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(digitalocean.Placement))
		fmt.Fprintf(out, "\t\tPublic key: %s\n", fastly.ToValue(digitalocean.PublicKey))
	}
	fmt.Fprintln(out)

	return nil
}
