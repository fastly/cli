package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DomainInspectorCommand exposes the Domain Inspector API.
type DomainInspectorCommand struct {
	argparser.Base

	// Optional.
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	From        string                          `help:"Start time (RFC3339 or Unix timestamp)"`
	To          string                          `help:"End time (RFC3339 or Unix timestamp)"`
	Downsample  string                          `default:"" enum:"minute,hour,day," help:"Sample window (minute/hour/day)"`
	Metrics     []string                        `name:"metric" sep:"none" help:"Metrics to retrieve (repeatable, up to 10)"`
	Domains     []string                        `name:"domain" sep:"none" help:"Filter by domain (repeatable)"`
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
func (c *DomainInspectorCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec implements the command interface.
func (c *DomainInspectorCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := fastly.GetDomainMetricsInput{
		ServiceID:   serviceID,
		Datacenters: c.Datacenters,
		Domains:     c.Domains,
		GroupBy:     c.GroupBy,
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
		if err := c.Globals.APIClient.GetDomainMetricsForServiceJSON(context.TODO(), &input, &raw); err != nil {
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
		resp, err := c.Globals.APIClient.GetDomainMetricsForService(context.TODO(), &input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{"Service ID": serviceID})
			return err
		}
		if fastly.ToValue(resp.Status) != statusSuccess {
			return fmt.Errorf("non-success response: %s", fastly.ToValue(resp.Status))
		}
		text.PrintDomainInspectorTbl(out, resp)
		return nil
	}
}

func parseTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if epoch, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(epoch, 0), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q as RFC3339 or Unix timestamp", s)
}
