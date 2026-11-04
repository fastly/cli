package domain

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe a domain.
type DescribeCommand struct {
	argparser.Base

	Input          fastly.GetDomainInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	NameFlag string `name:"name" short:"n" required:"" help:"Name of domain"`
	Version  string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.Name = c.NameFlag
	c.serviceVersion.Value = c.Version
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

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
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
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.GetDomain(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", fastly.ToValue(o.ServiceID))
	}
	fmt.Fprintf(out, "Version: %d\n", fastly.ToValue(o.ServiceVersion))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(o.Name))
	fmt.Fprintf(out, "Comment: %v\n", fastly.ToValue(o.Comment))

	return nil
}
