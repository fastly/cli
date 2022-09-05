package sftp

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v6/fastly"
)

// DescribeCommand calls the Fastly API to describe an SFTP logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetSFTPInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about an SFTP logging endpoint on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "The name of the SFTP logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	sftp, err := c.Globals.APIClient.GetSFTP(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(sftp)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error: unable to write data to stdout: %w", err)
		}
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", sftp.ServiceID)
	}
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
