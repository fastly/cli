package sftp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// DescribeCommand calls the Fastly API to describe an SFTP logging endpoint.
type DescribeCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.GetSFTPInput
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent common.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an SFTP logging endpoint on a Fastly service version").Alias("get")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	c.CmdClause.Flag("name", "The name of the SFTP logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	sftp, err := c.Globals.Client.GetSFTP(&c.Input)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", sftp.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", sftp.Version)
	fmt.Fprintf(out, "Name: %s\n", sftp.Name)
	fmt.Fprintf(out, "Address: %s\n", sftp.Address)
	fmt.Fprintf(out, "Port: %d\n", sftp.Port)
	fmt.Fprintf(out, "User: %s\n", sftp.User)
	fmt.Fprintf(out, "Password: %s\n", sftp.Password)
	fmt.Fprintf(out, "Public key: %s\n", sftp.PublicKey)
	fmt.Fprintf(out, "Secret key: %s\n", sftp.SecretKey)
	fmt.Fprintf(out, "SSH known hosts: %s\n", sftp.SSHKnownHosts)
	fmt.Fprintf(out, "Path: %s\n", sftp.Path)
	fmt.Fprintf(out, "Period: %d\n", sftp.Period)
	fmt.Fprintf(out, "GZip level: %d\n", sftp.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", sftp.Format)
	fmt.Fprintf(out, "Format version: %d\n", sftp.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", sftp.ResponseCondition)
	fmt.Fprintf(out, "Timestamp format: %s\n", sftp.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", sftp.Placement)

	return nil
}
