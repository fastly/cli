package ftp

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v5/fastly"
)

// DescribeCommand calls the Fastly API to describe an FTP logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetFTPInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about an FTP logging endpoint on a Fastly service version").Alias("get")
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
	})
	c.CmdClause.Flag("name", "The name of the FTP logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	ftp, err := c.Globals.Client.GetFTP(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", ftp.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", ftp.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", ftp.Name)
	fmt.Fprintf(out, "Address: %s\n", ftp.Address)
	fmt.Fprintf(out, "Port: %d\n", ftp.Port)
	fmt.Fprintf(out, "Username: %s\n", ftp.Username)
	fmt.Fprintf(out, "Password: %s\n", ftp.Password)
	fmt.Fprintf(out, "Public key: %s\n", ftp.PublicKey)
	fmt.Fprintf(out, "Path: %s\n", ftp.Path)
	fmt.Fprintf(out, "Period: %d\n", ftp.Period)
	fmt.Fprintf(out, "GZip level: %d\n", ftp.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", ftp.Format)
	fmt.Fprintf(out, "Format version: %d\n", ftp.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", ftp.ResponseCondition)
	fmt.Fprintf(out, "Timestamp format: %s\n", ftp.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", ftp.Placement)
	fmt.Fprintf(out, "Compression codec: %s\n", ftp.CompressionCodec)

	return nil
}
