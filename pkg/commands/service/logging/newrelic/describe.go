package newrelic

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.serviceVersion.Value = c.Version
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	NameFlag       string `name:"name" required:"" help:"The name for the real-time logging configuration"`
	serviceVersion argparser.OptionalServiceVersion

	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
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

	input := c.constructInput(serviceID, fastly.ToValue(serviceVersion.Number))

	o, err := c.Globals.APIClient.GetNewRelic(context.TODO(), input)
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

	return c.print(out, o)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput(serviceID string, serviceVersion int) *fastly.GetNewRelicInput {
	var input fastly.GetNewRelicInput

	input.Name = c.NameFlag
	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, nr *fastly.NewRelic) error {
	lines := text.Lines{
		"Format Version":     fastly.ToValue(nr.FormatVersion),
		"Format":             fastly.ToValue(nr.Format),
		"Name":               fastly.ToValue(nr.Name),
		"Placement":          fastly.ToValue(nr.Placement),
		"Processing region":  fastly.ToValue(nr.ProcessingRegion),
		"Region":             fastly.ToValue(nr.Region),
		"Response Condition": fastly.ToValue(nr.ResponseCondition),
		"Service Version":    fastly.ToValue(nr.ServiceVersion),
		"Token":              fastly.ToValue(nr.Token),
	}
	if nr.CreatedAt != nil {
		lines["Created at"] = nr.CreatedAt
	}
	if nr.UpdatedAt != nil {
		lines["Updated at"] = nr.UpdatedAt
	}
	if nr.DeletedAt != nil {
		lines["Deleted at"] = nr.DeletedAt
	}

	if !c.Globals.Verbose() {
		lines["Service ID"] = fastly.ToValue(nr.ServiceID)
	}
	text.PrintLines(out, lines)

	return nil
}
