package ftp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// ListCommand calls the Fastly API to list FTP logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListFTPsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List FTP endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	ftps, err := c.Globals.Client.ListFTPs(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, ftp := range ftps {
			tw.AddLine(ftp.ServiceID, ftp.Version, ftp.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.Service)
	fmt.Fprintf(out, "Version: %d\n", c.Input.Version)
	for i, ftp := range ftps {
		fmt.Fprintf(out, "\tFTP %d/%d\n", i+1, len(ftps))
		fmt.Fprintf(out, "\t\tService ID: %s\n", ftp.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", ftp.Version)
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
	}
	fmt.Fprintln(out)

	return nil
}
