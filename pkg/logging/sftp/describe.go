package sftp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe an SFTP logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetSFTPInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about an SFTP logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the SFTP logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	sftp, err := c.Globals.Client.GetSFTP(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", sftp.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", sftp.ServiceVersion)
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
	fmt.Fprintf(out, "Message type: %s\n", sftp.MessageType)
	fmt.Fprintf(out, "Response condition: %s\n", sftp.ResponseCondition)
	fmt.Fprintf(out, "Timestamp format: %s\n", sftp.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", sftp.Placement)
	fmt.Fprintf(out, "Compression codec: %s\n", sftp.CompressionCodec)

	return nil
}
