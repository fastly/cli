package openstack

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ListCommand calls the Fastly API to list OpenStack logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListOpenstackInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List OpenStack logging endpoints on a Fastly service version")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
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
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
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

	openstacks, err := c.Globals.APIClient.ListOpenstack(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(openstacks)
			if err != nil {
				return err
			}
			out.Write(data)
			return nil
		}

		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, openstack := range openstacks {
			tw.AddLine(openstack.ServiceID, openstack.ServiceVersion, openstack.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, openstack := range openstacks {
		fmt.Fprintf(out, "\tOpenstack %d/%d\n", i+1, len(openstacks))
		fmt.Fprintf(out, "\t\tService ID: %s\n", openstack.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", openstack.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", openstack.Name)
		fmt.Fprintf(out, "\t\tBucket: %s\n", openstack.BucketName)
		fmt.Fprintf(out, "\t\tAccess key: %s\n", openstack.AccessKey)
		fmt.Fprintf(out, "\t\tUser: %s\n", openstack.User)
		fmt.Fprintf(out, "\t\tURL: %s\n", openstack.URL)
		fmt.Fprintf(out, "\t\tPath: %s\n", openstack.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", openstack.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", openstack.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", openstack.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", openstack.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", openstack.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", openstack.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", openstack.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", openstack.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", openstack.PublicKey)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", openstack.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
