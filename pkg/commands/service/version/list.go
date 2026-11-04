package version

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	fsttime "github.com/fastly/cli/pkg/time"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input       fastly.ListVersionsInput        `kong:"-"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
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

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID

	o, err := c.Globals.APIClient.ListVersions(context.TODO(), &c.Input)
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
		tw.AddHeader("NUMBER", "ACTIVE", "STAGED", "LAST EDITED (UTC)")
		for _, version := range o {
			tw.AddLine(
				fastly.ToValue(version.Number),
				fastly.ToValue(version.Active),
				fastly.ToValue(version.Staging),
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
