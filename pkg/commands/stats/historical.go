package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
)

const statusSuccess = "success"

// HistoricalCommand exposes the Historical Stats API.
type HistoricalCommand struct {
	argparser.Base

	// Optional.
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Field       string                          `help:"Filter to a single stats field (e.g. bandwidth, requests)"`
	From        string                          `help:"From time, accepted formats at https://fastly.dev/reference/api/metrics-stats/historical-stats"`
	To          string                          `help:"To time"`
	By          string                          `default:"" enum:"minute,hour,day," help:"Aggregation period (minute/hour/day)"`
	Region      string                          `help:"Filter by region ('stats regions' to list)"`
	FormatFlag  string                          `name:"format" default:"" enum:"json," hidden:"" help:"Output format (json)"`
	JSONFlag    bool                            `name:"json" short:"j" help:"Render output as JSON"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *HistoricalCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec implements the command interface.
func (c *HistoricalCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := resolveJSONFormat(&c.FormatFlag, c.JSONFlag, c.Globals); err != nil {
		return err
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := fastly.GetStatsInput{
		Service: new(serviceID),
	}
	if c.By != "" {
		input.By = &c.By
	}
	if c.Field != "" {
		input.Field = &c.Field
	}
	if c.From != "" {
		input.From = &c.From
	}
	if c.Region != "" {
		input.Region = &c.Region
	}
	if c.To != "" {
		input.To = &c.To
	}

	var envelope statsResponse
	err = c.Globals.APIClient.GetStatsJSON(context.TODO(), &input, &envelope)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if envelope.Status != statusSuccess {
		return fmt.Errorf("non-success response: %s", envelope.Msg)
	}

	switch c.FormatFlag {
	case "json":
		if err := writeBlocksJSON(out, envelope.Data); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
			return err
		}

	default:
		writeHeader(out, envelope.Meta)
		if c.Field != "" {
			for _, block := range envelope.Data {
				if err := fmtFieldLine(out, c.Field, block); err != nil {
					c.Globals.ErrLog.AddWithContext(err, map[string]any{
						"Service ID": serviceID,
					})
					return err
				}
			}
		} else if err := writeBlocks(out, serviceID, envelope.Data); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
			return err
		}
	}

	return nil
}

func writeHeader(out io.Writer, meta statsResponseMeta) {
	fmt.Fprintf(out, "From: %s\n", meta.From)
	fmt.Fprintf(out, "To: %s\n", meta.To)
	fmt.Fprintf(out, "By: %s\n", meta.By)
	fmt.Fprintf(out, "Region: %s\n", meta.Region)
	fmt.Fprintf(out, "---\n")
}

func writeBlocks(out io.Writer, service string, blocks []statsResponseData) error {
	for _, block := range blocks {
		if err := fmtBlock(out, service, block); err != nil {
			return err
		}
	}

	return nil
}

func writeBlocksJSON(out io.Writer, blocks []statsResponseData) error {
	for _, block := range blocks {
		if err := json.NewEncoder(out).Encode(block); err != nil {
			return err
		}
	}

	return nil
}
