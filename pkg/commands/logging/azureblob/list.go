package azureblob

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// ListCommand calls the Fastly API to list Azure Blob Storage logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListBlobStoragesInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Azure Blob Storage logging endpoints on a Fastly service version")
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
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
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

	azureblobs, err := c.Globals.APIClient.ListBlobStorages(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(azureblobs)
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

		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, azureblob := range azureblobs {
			tw.AddLine(azureblob.ServiceID, azureblob.ServiceVersion, azureblob.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, azureblob := range azureblobs {
		fmt.Fprintf(out, "\tBlobStorage %d/%d\n", i+1, len(azureblobs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", azureblob.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", azureblob.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", azureblob.Name)
		fmt.Fprintf(out, "\t\tContainer: %s\n", azureblob.Container)
		fmt.Fprintf(out, "\t\tAccount name: %s\n", azureblob.AccountName)
		fmt.Fprintf(out, "\t\tSAS token: %s\n", azureblob.SASToken)
		fmt.Fprintf(out, "\t\tPath: %s\n", azureblob.Path)
		fmt.Fprintf(out, "\t\tPeriod: %d\n", azureblob.Period)
		fmt.Fprintf(out, "\t\tGZip level: %d\n", azureblob.GzipLevel)
		fmt.Fprintf(out, "\t\tFormat: %s\n", azureblob.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", azureblob.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", azureblob.ResponseCondition)
		fmt.Fprintf(out, "\t\tMessage type: %s\n", azureblob.MessageType)
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", azureblob.TimestampFormat)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", azureblob.Placement)
		fmt.Fprintf(out, "\t\tPublic key: %s\n", azureblob.PublicKey)
		fmt.Fprintf(out, "\t\tFile max bytes: %d\n", azureblob.FileMaxBytes)
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", azureblob.CompressionCodec)
	}
	fmt.Fprintln(out)

	return nil
}
