package ratelimit

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all rate limiters for a particular service and version")
	c.Globals = g
	c.manifest = m

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

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest       manifest.Data
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
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

	input := &fastly.ListERLsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion.Number,
	}

	o, err := c.Globals.APIClient.ListERLs(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, o)
	} else {
		c.printSummary(out, o)
	}
	return nil
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, o []*fastly.ERL) {
	for _, u := range o {
		fmt.Fprintf(out, "\nAction: %+v\n", u.Action)
		fmt.Fprintf(out, "Client Key: %+v\n", u.ClientKey)
		fmt.Fprintf(out, "Feature Revision: %+v\n", u.FeatureRevision)
		fmt.Fprintf(out, "HTTP Methods: %+v\n", u.HTTPMethods)
		fmt.Fprintf(out, "ID: %+v\n", u.ID)
		fmt.Fprintf(out, "Logger Type: %+v\n", u.LoggerType)
		fmt.Fprintf(out, "Name: %+v\n", u.Name)
		fmt.Fprintf(out, "Penalty Box Duration: %+v\n", u.PenaltyBoxDuration)
		fmt.Fprintf(out, "Response: %+v\n", u.Response)
		fmt.Fprintf(out, "Response Object Name: %+v\n", u.ResponseObjectName)
		fmt.Fprintf(out, "RPS Limit: %+v\n", u.RpsLimit)
		fmt.Fprintf(out, "Service ID: %+v\n", u.ServiceID)
		fmt.Fprintf(out, "URI Dictionary Name: %+v\n", u.URIDictionaryName)
		fmt.Fprintf(out, "Version: %+v\n", u.Version)
		fmt.Fprintf(out, "WindowSize: %+v\n", u.WindowSize)

		if u.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", u.CreatedAt)
		}
		if u.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", u.UpdatedAt)
		}
		if u.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", u.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, o []*fastly.ERL) {
	t := text.NewTable(out)
	t.AddHeader("ID", "NAME", "ACTION", "RPS LIMIT", "WINDOW SIZE", "PENALTY BOX DURATION")
	for _, u := range o {
		t.AddLine(u.ID, u.Name, u.Action, u.RpsLimit, u.WindowSize, u.PenaltyBoxDuration)
	}
	t.Print()
}
