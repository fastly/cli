package logshuttle

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

// DescribeCommand calls the Fastly API to describe a Logshuttle logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetLogshuttleInput
	json           bool
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("describe", "Show detailed information about a Logshuttle logging endpoint on a Fastly service version").Alias("get")
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
	c.CmdClause.Flag("name", "The name of the Logshuttle logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	logshuttle, err := c.Globals.APIClient.GetLogshuttle(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(logshuttle)
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
		fmt.Fprintf(out, "\nService ID: %s\n", logshuttle.ServiceID)
	}
	fmt.Fprintf(out, "Version: %d\n", logshuttle.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", logshuttle.Name)
	fmt.Fprintf(out, "URL: %s\n", logshuttle.URL)
	fmt.Fprintf(out, "Token: %s\n", logshuttle.Token)
	fmt.Fprintf(out, "Format: %s\n", logshuttle.Format)
	fmt.Fprintf(out, "Format version: %d\n", logshuttle.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", logshuttle.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", logshuttle.Placement)

	return nil
}
