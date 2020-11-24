package sftp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// ListCommand calls the Fastly API to list SFTP logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListSFTPsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List SFTP endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	sftps, err := c.Globals.Client.ListSFTPs(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, sftp := range sftps {
			tw.AddLine(sftp.ServiceID, sftp.ServiceVersion, sftp.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, sftp := range sftps {
		fmt.Fprintf(out, "\tSFTP %d/%d\n", i+1, len(sftps))
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
	}
	fmt.Fprintln(out)

	return nil
}
