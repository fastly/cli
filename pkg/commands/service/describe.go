package service

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
)

// DescribeCommand calls the Fastly API to describe a service.
type DescribeCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input       fastly.GetServiceDetailsInput   `kong:"-"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	if source == manifest.SourceUndefined && !c.ServiceName.WasSet {
		err := fsterr.ErrNoServiceID
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID

	o, err := c.Globals.APIClient.GetServiceDetails(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(o, out)
}

func (c *DescribeCommand) print(s *fastly.ServiceDetail, out io.Writer) error {
	fmt.Fprintf(out, "ID: %s\n", fastly.ToValue(s.ServiceID))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(s.Name))
	fmt.Fprintf(out, "Type: %s\n", fastly.ToValue(s.Type))
	fmt.Fprintf(out, "Comment: %s\n", fastly.ToValue(s.Comment))
	fmt.Fprintf(out, "Customer ID: %s\n", fastly.ToValue(s.CustomerID))
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(time.Format))
	}
	if s.ActiveVersion != nil {
		fmt.Fprintf(out, "Active version:\n")
		text.PrintVersion(out, "\t", s.ActiveVersion)
	}
	fmt.Fprintf(out, "Versions: %d\n", len(s.Versions))
	for j, version := range s.Versions {
		fmt.Fprintf(out, "\tVersion %d/%d\n", j+1, len(s.Versions))
		text.PrintVersion(out, "\t\t", version)
	}
	return nil
}
