package kinesis

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// ListCommand calls the Fastly API to list Amazon Kinesis logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest       manifest.Data
	Input          fastly.ListKinesisInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("list", "List Kinesis endpoints on a Fastly service version")

	// required
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// optional
	c.RegisterFlagBool(c.JSONFlag()) // --json
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
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
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
	c.Input.ServiceVersion = serviceVersion.Number

	o, err := c.Globals.APIClient.ListKinesis(&c.Input)
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
		for _, kinesis := range o {
			tw.AddLine(kinesis.ServiceID, kinesis.ServiceVersion, kinesis.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kinesis := range o {
		fmt.Fprintf(out, "\tKinesis %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", kinesis.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", kinesis.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", kinesis.Name)
		fmt.Fprintf(out, "\t\tStream name: %s\n", kinesis.StreamName)
		fmt.Fprintf(out, "\t\tRegion: %s\n", kinesis.Region)
		if kinesis.AccessKey != "" || kinesis.SecretKey != "" {
			fmt.Fprintf(out, "\t\tAccess key: %s\n", kinesis.AccessKey)
			fmt.Fprintf(out, "\t\tSecret key: %s\n", kinesis.SecretKey)
		}
		if kinesis.IAMRole != "" {
			fmt.Fprintf(out, "\t\tIAM role: %s\n", kinesis.IAMRole)
		}
		fmt.Fprintf(out, "\t\tFormat: %s\n", kinesis.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", kinesis.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", kinesis.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", kinesis.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
