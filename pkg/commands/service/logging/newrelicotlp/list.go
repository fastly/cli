package newrelicotlp

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Required.
	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
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

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	o, err := c.Globals.APIClient.ListNewRelicOTLP(context.TODO(), input)
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

	if c.Globals.Verbose() {
		c.printVerbose(out, fastly.ToValue(serviceVersion.Number), o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput(serviceID string, serviceVersion int) *fastly.ListNewRelicOTLPInput {
	var input fastly.ListNewRelicOTLPInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, serviceVersion int, ls []*fastly.NewRelicOTLP) {
	fmt.Fprintf(out, "Service Version: %d\n", serviceVersion)

	for _, l := range ls {
		fmt.Fprintf(out, "\nName: %s\n", fastly.ToValue(l.Name))
		fmt.Fprintf(out, "\nToken: %s\n", fastly.ToValue(l.Token))
		fmt.Fprintf(out, "\nFormat: %s\n", fastly.ToValue(l.Format))
		fmt.Fprintf(out, "\nFormat Version: %d\n", fastly.ToValue(l.FormatVersion))
		fmt.Fprintf(out, "\nPlacement: %s\n", fastly.ToValue(l.Placement))
		fmt.Fprintf(out, "\nRegion: %s\n", fastly.ToValue(l.Region))
		fmt.Fprintf(out, "\nProcessing region: %s\n", fastly.ToValue(l.ProcessingRegion))
		fmt.Fprintf(out, "\nResponse Condition: %s\n\n", fastly.ToValue(l.ResponseCondition))

		if l.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", l.CreatedAt)
		}
		if l.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", l.UpdatedAt)
		}
		if l.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", l.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, nrs []*fastly.NewRelicOTLP) error {
	t := text.NewTable(out)
	t.AddHeader("SERVICE ID", "VERSION", "NAME")
	for _, nr := range nrs {
		t.AddLine(
			fastly.ToValue(nr.ServiceID),
			fastly.ToValue(nr.ServiceVersion),
			fastly.ToValue(nr.Name),
		)
	}
	t.Print()
	return nil
}
