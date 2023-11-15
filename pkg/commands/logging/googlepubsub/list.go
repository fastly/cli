package googlepubsub

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Google Cloud Pub/Sub logging endpoints.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input          fastly.ListPubsubsInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Google Cloud Pub/Sub endpoints on a Fastly service version")

	// Required.
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
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
		Manifest:           *c.Globals.Manifest,
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

	o, err := c.Globals.APIClient.ListPubsubs(&c.Input)
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
		for _, googlepubsub := range o {
			tw.AddLine(googlepubsub.ServiceID, googlepubsub.ServiceVersion, googlepubsub.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, googlepubsub := range o {
		fmt.Fprintf(out, "\tGoogle Cloud Pub/Sub %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", googlepubsub.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", googlepubsub.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", googlepubsub.Name)
		fmt.Fprintf(out, "\t\tUser: %s\n", googlepubsub.User)
		fmt.Fprintf(out, "\t\tAccount name: %s\n", googlepubsub.AccountName)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", googlepubsub.SecretKey)
		fmt.Fprintf(out, "\t\tProject ID: %s\n", googlepubsub.ProjectID)
		fmt.Fprintf(out, "\t\tTopic: %s\n", googlepubsub.Topic)
		fmt.Fprintf(out, "\t\tFormat: %s\n", googlepubsub.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", googlepubsub.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", googlepubsub.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", googlepubsub.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
