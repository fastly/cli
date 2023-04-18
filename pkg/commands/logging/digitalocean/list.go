package digitalocean

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// ListCommand calls the Fastly API to list DigitalOcean Spaces logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest       manifest.Data
	Input          fastly.ListDigitalOceansInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("list", "List DigitalOcean Spaces logging endpoints on a Fastly service version")

	// required
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// optional
	c.RegisterFlagBool(c.JSONFlag()) // --json
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
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
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
	c.Input.ServiceVersion = serviceVersion.Number

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
			tw.AddLine(digitalocean.ServiceID, digitalocean.ServiceVersion, digitalocean.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, digitalocean := range o {
		fmt.Fprintf(out, "\tDigitalOcean %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", digitalocean.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", digitalocean.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", digitalocean.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", digitalocean.BucketName)
		fmt.Fprintf(out, "\t\tDomain: %s\n", digitalocean.Domain)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", digitalocean.AccessKey)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", digitalocean.SecretKey)
		fmt.Fprintf(out, "\t\tPath: %s\n", digitalocean.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", digitalocean.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", digitalocean.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", digitalocean.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", digitalocean.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", digitalocean.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", digitalocean.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", digitalocean.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", digitalocean.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", digitalocean.PublicKey)
	}
	fmt.Fprintln(out)

	return nil
}
