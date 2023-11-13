package sftp

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list SFTP logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input          fastly.ListSFTPsInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List SFTP endpoints on a Fastly service version")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
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
	c.Input.ServiceVersion = serviceVersion.Number

	o, err := c.Globals.APIClient.ListSFTPs(&c.Input)
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
		for _, sftp := range o {
			tw.AddLine(sftp.ServiceID, sftp.ServiceVersion, sftp.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, sftp := range o {
		fmt.Fprintf(out, "\tSFTP %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", sftp.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", sftp.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", sftp.Name)
		fmt.Fprintf(out, "\t\tAddress: %s\n", sftp.Address)
		fmt.Fprintf(out, "\t\tPort: %d\n", sftp.Port)
		fmt.Fprintf(out, "\t\tUser: %s\n", sftp.User)
		fmt.Fprintf(out, "\t\tPassword: %s\n", sftp.Password)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", sftp.PublicKey)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", sftp.SecretKey)
		fmt.Fprintf(out, "\t\tSSH known hosts: %s\n", sftp.SSHKnownHosts)
		fmt.Fprintf(out, "\t\tPath: %s\n", sftp.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", sftp.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", sftp.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", sftp.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", sftp.FormatVersion)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", sftp.MessageType)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", sftp.ResponseCondition)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", sftp.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", sftp.Placement)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", sftp.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
