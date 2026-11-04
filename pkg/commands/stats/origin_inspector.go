package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// OriginInspectorCommand exposes the Origin Inspector API.
type OriginInspectorCommand struct {
	argparser.Base

	// Optional.
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	From        string                          `help:"Start time (RFC3339 or Unix timestamp)"`
	To          string                          `help:"End time (RFC3339 or Unix timestamp)"`
	Downsample  string                          `default:"" enum:"minute,hour,day," help:"Sample window (minute/hour/day)"`
	Metrics     []string                        `name:"metric" sep:"none" help:"Metrics to retrieve (repeatable, up to 10)"`
	Hosts       []string                        `name:"host" sep:"none" help:"Filter by origin host (repeatable)"`
	Datacenters []string                        `name:"datacenter" sep:"none" help:"Filter by POP (repeatable)"`
	Regions     []string                        `name:"region" sep:"none" help:"Filter by region (repeatable)"`
	GroupBy     []string                        `sep:"none" help:"Dimensions to group by (repeatable)"`
	Limit       int                             `help:"Max entries to return"`
	Cursor      string                          `help:"Pagination cursor from a previous response"`
	FormatFlag  string                          `name:"format" default:"" enum:"json," hidden:"" help:"Output format (json)"`
	JSONFlag    bool                            `name:"json" short:"j" help:"Render output as JSON"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *OriginInspectorCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec implements the command interface.
func (c *OriginInspectorCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := fastly.GetOriginMetricsInput{
		ServiceID:   serviceID,
		Datacenters: c.Datacenters,
		GroupBy:     c.GroupBy,
		Hosts:       c.Hosts,
		Metrics:     c.Metrics,
		Regions:     c.Regions,
	}
	if c.Cursor != "" {
		input.Cursor = &c.Cursor
	}
	if c.Downsample != "" {
		input.Downsample = &c.Downsample
	}
	if c.From != "" {
		t, err := parseTime(c.From)
		if err != nil {
			return fmt.Errorf("invalid --from value: %w", err)
		}
		input.Start = &t
	}
	if c.To != "" {
		t, err := parseTime(c.To)
		if err != nil {
			return fmt.Errorf("invalid --to value: %w", err)
		}
		input.End = &t
	}
	if c.Limit > 0 {
		input.Limit = &c.Limit
	}

	switch c.FormatFlag {
	case "json":
		var envelope struct {
			Status *string `json:"status"`
		}
		var raw json.RawMessage
		if err := c.Globals.APIClient.GetOriginMetricsForServiceJSON(context.TODO(), &input, &raw); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{"Service ID": serviceID})
			return err
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			return err
		}
		if fastly.ToValue(envelope.Status) != statusSuccess {
			return fmt.Errorf("non-success response: %s", fastly.ToValue(envelope.Status))
		}
		_, err := out.Write(raw)
		if err != nil {
			return err
		}
		fmt.Fprintln(out)
		return nil

	default:
		resp, err := c.Globals.APIClient.GetOriginMetricsForService(context.TODO(), &input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{"Service ID": serviceID})
			return err
		}
		if fastly.ToValue(resp.Status) != statusSuccess {
			return fmt.Errorf("non-success response: %s", fastly.ToValue(resp.Status))
		}
		text.PrintOriginInspectorTbl(out, resp)
		return nil
	}
}
