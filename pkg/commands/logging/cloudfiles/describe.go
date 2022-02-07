package cloudfiles

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

// DescribeCommand calls the Fastly API to describe a Cloudfiles logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetCloudfilesInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a Cloudfiles logging endpoint on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "The name of the Cloudfiles logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
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
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	cloudfiles, err := c.Globals.APIClient.GetCloudfiles(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if c.json {
		data, err := json.Marshal(cloudfiles)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", cloudfiles.ServiceID)
	}
	fmt.Fprintf(out, "Version: %d\n", cloudfiles.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", cloudfiles.Name)
	fmt.Fprintf(out, "User: %s\n", cloudfiles.User)
	fmt.Fprintf(out, "Access key: %s\n", cloudfiles.AccessKey)
	fmt.Fprintf(out, "Bucket: %s\n", cloudfiles.BucketName)
	fmt.Fprintf(out, "Path: %s\n", cloudfiles.Path)
	fmt.Fprintf(out, "Region: %s\n", cloudfiles.Region)
	fmt.Fprintf(out, "Placement: %s\n", cloudfiles.Placement)
	fmt.Fprintf(out, "Period: %d\n", cloudfiles.Period)
	fmt.Fprintf(out, "GZip level: %d\n", cloudfiles.GzipLevel)
	fmt.Fprintf(out, "Format: %s\n", cloudfiles.Format)
	fmt.Fprintf(out, "Format version: %d\n", cloudfiles.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", cloudfiles.ResponseCondition)
	fmt.Fprintf(out, "Message type: %s\n", cloudfiles.MessageType)
	fmt.Fprintf(out, "Timestamp format: %s\n", cloudfiles.TimestampFormat)
	fmt.Fprintf(out, "Public key: %s\n", cloudfiles.PublicKey)

	return nil
}
