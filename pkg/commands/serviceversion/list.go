package serviceversion

import (
	"fmt"
	"io"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	fsttime "github.com/fastly/cli/pkg/time"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input       fastly.ListVersionsInput
	serviceName argparser.OptionalServiceNameID
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Fastly service versions")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID

	o, err := c.Globals.APIClient.ListVersions(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("NUMBER", "ACTIVE", "LAST EDITED (UTC)")
		for _, version := range o {
			tw.AddLine(
				fastly.ToValue(version.Number),
				fastly.ToValue(version.Active),
				parseTime(version.UpdatedAt),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Versions: %d\n", len(o))
	for i, version := range o {
		fmt.Fprintf(out, "\tVersion %d/%d\n", i+1, len(o))
		text.PrintVersion(out, "\t\t", version)
	}
	fmt.Fprintln(out)

	return nil
}

func parseTime(ua *time.Time) string {
	if ua == nil {
		return ""
	}
	return ua.UTC().Format(fsttime.Format)
}
