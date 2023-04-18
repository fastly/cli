package ftp

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

// ListCommand calls the Fastly API to list FTP logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest       manifest.Data
	Input          fastly.ListFTPsInput
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
	c.CmdClause = parent.Command("list", "List FTP endpoints on a Fastly service version")

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

	o, err := c.Globals.APIClient.ListFTPs(&c.Input)
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
		for _, ftp := range o {
			tw.AddLine(ftp.ServiceID, ftp.ServiceVersion, ftp.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, ftp := range o {
		fmt.Fprintf(out, "\tFTP %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", ftp.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", ftp.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", ftp.Name)
		fmt.Fprintf(out, "\t\tAddress: %s\n", ftp.Address)
		fmt.Fprintf(out, "\t\tPort: %d\n", ftp.Port)
		fmt.Fprintf(out, "\t\tUsername: %s\n", ftp.Username)
		fmt.Fprintf(out, "\t\tPassword: %s\n", ftp.Password)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", ftp.PublicKey)
		fmt.Fprintf(out, "\t\tPath: %s\n", ftp.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", ftp.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", ftp.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", ftp.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", ftp.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", ftp.ResponseCondition)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", ftp.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", ftp.Placement)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", ftp.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
