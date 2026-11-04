package sftp

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list SFTP logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListSFTPsInput           `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
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
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.ListSFTPs(context.TODO(), &c.Input)
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
			tw.AddLine(
				fastly.ToValue(sftp.ServiceID),
				fastly.ToValue(sftp.ServiceVersion),
				fastly.ToValue(sftp.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, sftp := range o {
		fmt.Fprintf(out, "\tSFTP %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(sftp.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(sftp.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(sftp.Name))
		fmt.Fprintf(out, "\t\tAddress: %s\n", fastly.ToValue(sftp.Address))
		fmt.Fprintf(out, "\t\tPort: %d\n", fastly.ToValue(sftp.Port))
		fmt.Fprintf(out, "\t\tUser: %s\n", fastly.ToValue(sftp.User))
		fmt.Fprintf(out, "\t\tPassword: %s\n", fastly.ToValue(sftp.Password))
		fmt.Fprintf(out, "\t\tPublic key: %s\n", fastly.ToValue(sftp.PublicKey))
		fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(sftp.SecretKey))
		fmt.Fprintf(out, "\t\tSSH known hosts: %s\n", fastly.ToValue(sftp.SSHKnownHosts))
		fmt.Fprintf(out, "\t\tPath: %s\n", fastly.ToValue(sftp.Path))
		fmt.Fprintf(out, "\t\tPeriod: %d\n", fastly.ToValue(sftp.Period))
		fmt.Fprintf(out, "\t\tGZip level: %d\n", fastly.ToValue(sftp.GzipLevel))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(sftp.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(sftp.FormatVersion))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(sftp.MessageType))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(sftp.ResponseCondition))
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", fastly.ToValue(sftp.TimestampFormat))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(sftp.Placement))
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", fastly.ToValue(sftp.CompressionCodec))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(sftp.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
