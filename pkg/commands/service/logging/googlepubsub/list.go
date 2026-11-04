package googlepubsub

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Google Cloud Pub/Sub logging endpoints.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	Input          fastly.ListPubsubsInput         `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
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

	o, err := c.Globals.APIClient.ListPubsubs(context.TODO(), &c.Input)
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
			tw.AddLine(
				fastly.ToValue(googlepubsub.ServiceID),
				fastly.ToValue(googlepubsub.ServiceVersion),
				fastly.ToValue(googlepubsub.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, googlepubsub := range o {
		fmt.Fprintf(out, "\tGoogle Cloud Pub/Sub %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(googlepubsub.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(googlepubsub.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(googlepubsub.Name))
		fmt.Fprintf(out, "\t\tUser: %s\n", fastly.ToValue(googlepubsub.User))
		fmt.Fprintf(out, "\t\tAccount name: %s\n", fastly.ToValue(googlepubsub.AccountName))
		fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(googlepubsub.SecretKey))
		fmt.Fprintf(out, "\t\tProject ID: %s\n", fastly.ToValue(googlepubsub.ProjectID))
		fmt.Fprintf(out, "\t\tTopic: %s\n", fastly.ToValue(googlepubsub.Topic))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(googlepubsub.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(googlepubsub.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(googlepubsub.ResponseCondition))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(googlepubsub.Placement))
		fmt.Fprintf(out, "\t\tProcessing region: %s\n", fastly.ToValue(googlepubsub.ProcessingRegion))
	}
	fmt.Fprintln(out)

	return nil
}
