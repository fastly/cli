package gcs

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v5/fastly"
)

// DescribeCommand calls the Fastly API to describe a GCS logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetGCSInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a GCS logging endpoint on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "The name of the GCS logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

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
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	gcs, err := c.Globals.Client.GetGCS(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(gcs)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", gcs.ServiceID)
	}
	fmt.Fprintf(out, "Version: %d\n", gcs.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", gcs.Name)
	fmt.Fprintf(out, "Bucket: %s\n", gcs.Bucket)
	fmt.Fprintf(out, "User: %s\n", gcs.User)
	fmt.Fprintf(out, "Secret key: %s\n", gcs.SecretKey)
	fmt.Fprintf(out, "Path: %s\n", gcs.Path)
	fmt.Fprintf(out, "Period: %d\n", gcs.Period)
	fmt.Fprintf(out, "GZip level: %d\n", gcs.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", gcs.Format)
	fmt.Fprintf(out, "Format version: %d\n", gcs.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", gcs.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", gcs.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", gcs.TimestampFormat)
	fmt.Fprintf(out, "Placement: %s\n", gcs.Placement)
	fmt.Fprintf(out, "Compression codec: %s\n", gcs.CompressionCodec)

	return nil
}
