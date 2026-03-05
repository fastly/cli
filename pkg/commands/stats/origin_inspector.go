package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// OriginInspectorCommand exposes the Origin Inspector API.
type OriginInspectorCommand struct {
	argparser.Base

	cursor      string
	datacenters []string
	downsample  string
	formatFlag  string
	from        string
	groupBy     []string
	hosts       []string
	limit       int
	metrics     []string
	regions     []string
	serviceName argparser.OptionalServiceNameID
	to          string
}

// NewOriginInspectorCommand is the "stats origin-inspector" subcommand.
func NewOriginInspectorCommand(parent argparser.Registerer, g *global.Data) *OriginInspectorCommand {
	var c OriginInspectorCommand
	c.Globals = g

	c.CmdClause = parent.Command("origin-inspector", "View origin metrics for a Fastly service")
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})

	c.CmdClause.Flag("from", "Start time (RFC3339 or Unix timestamp)").StringVar(&c.from)
	c.CmdClause.Flag("to", "End time (RFC3339 or Unix timestamp)").StringVar(&c.to)
	c.CmdClause.Flag("downsample", "Sample window (minute/hour/day)").EnumVar(&c.downsample, "minute", "hour", "day")
	c.CmdClause.Flag("metric", "Metrics to retrieve (repeatable, up to 10)").StringsVar(&c.metrics)
	c.CmdClause.Flag("host", "Filter by origin host (repeatable)").StringsVar(&c.hosts)
	c.CmdClause.Flag("datacenter", "Filter by POP (repeatable)").StringsVar(&c.datacenters)
	c.CmdClause.Flag("region", "Filter by region (repeatable)").StringsVar(&c.regions)
	c.CmdClause.Flag("group-by", "Dimensions to group by (repeatable)").StringsVar(&c.groupBy)
	c.CmdClause.Flag("limit", "Max entries to return").IntVar(&c.limit)
	c.CmdClause.Flag("cursor", "Pagination cursor from a previous response").StringVar(&c.cursor)
	c.CmdClause.Flag("format", "Output format (json)").EnumVar(&c.formatFlag, "json")

	return &c
}

// Exec implements the command interface.
func (c *OriginInspectorCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := fastly.GetOriginMetricsInput{
		ServiceID:   serviceID,
		Datacenters: c.datacenters,
		GroupBy:     c.groupBy,
		Hosts:       c.hosts,
		Metrics:     c.metrics,
		Regions:     c.regions,
	}
	if c.cursor != "" {
		input.Cursor = &c.cursor
	}
	if c.downsample != "" {
		input.Downsample = &c.downsample
	}
	if c.from != "" {
		t, err := parseTime(c.from)
		if err != nil {
			return fmt.Errorf("invalid --from value: %w", err)
		}
		input.Start = &t
	}
	if c.to != "" {
		t, err := parseTime(c.to)
		if err != nil {
			return fmt.Errorf("invalid --to value: %w", err)
		}
		input.End = &t
	}
	if c.limit > 0 {
		input.Limit = &c.limit
	}

	switch c.formatFlag {
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
		return writeOriginInspector(out, resp)
	}
}

func writeOriginInspector(out io.Writer, resp *fastly.OriginInspector) error {
	if resp.Meta != nil {
		if resp.Meta.Start != nil {
			text.Output(out, "Start: %s", *resp.Meta.Start)
		}
		if resp.Meta.End != nil {
			text.Output(out, "End: %s", *resp.Meta.End)
		}
		fmt.Fprintln(out, "---")
	}
	for _, d := range resp.Data {
		if d.Dimensions != nil {
			for k, v := range d.Dimensions {
				text.Output(out, "%s: %s", k, v)
			}
		}
		for _, v := range d.Values {
			if v.Timestamp != nil {
				text.Output(out, "  Timestamp:  %s", time.Unix(int64(*v.Timestamp), 0).UTC()) //nolint:gosec // timestamp won't overflow
			}
			text.Output(out, "  Responses:  %d", fastly.ToValue(v.Responses))
			text.Output(out, "  Status 2xx: %d", fastly.ToValue(v.Status2xx))
			text.Output(out, "  Status 4xx: %d", fastly.ToValue(v.Status4xx))
			text.Output(out, "  Status 5xx: %d", fastly.ToValue(v.Status5xx))
		}
	}
	if resp.Meta != nil && resp.Meta.NextCursor != nil {
		text.Output(out, "Next cursor: %s", *resp.Meta.NextCursor)
	}
	return nil
}
