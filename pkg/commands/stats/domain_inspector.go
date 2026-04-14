package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/fastly/go-fastly/v14/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DomainInspectorCommand exposes the Domain Inspector API.
type DomainInspectorCommand struct {
	argparser.Base

	cursor      string
	datacenters []string
	domains     []string
	downsample  string
	formatFlag  string
	from        string
	groupBy     []string
	jsonFlag    bool
	limit       int
	metrics     []string
	regions     []string
	serviceName argparser.OptionalServiceNameID
	to          string
}

// NewDomainInspectorCommand is the "stats domain-inspector" subcommand.
func NewDomainInspectorCommand(parent argparser.Registerer, g *global.Data) *DomainInspectorCommand {
	var c DomainInspectorCommand
	c.Globals = g

	c.CmdClause = parent.Command("domain-inspector", "View domain metrics for a Fastly service")

	// Optional.
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
	c.CmdClause.Flag("domain", "Filter by domain (repeatable)").StringsVar(&c.domains)
	c.CmdClause.Flag("datacenter", "Filter by POP (repeatable)").StringsVar(&c.datacenters)
	c.CmdClause.Flag("region", "Filter by region (repeatable)").StringsVar(&c.regions)
	c.CmdClause.Flag("group-by", "Dimensions to group by (repeatable)").StringsVar(&c.groupBy)
	c.CmdClause.Flag("limit", "Max entries to return").IntVar(&c.limit)
	c.CmdClause.Flag("cursor", "Pagination cursor from a previous response").StringVar(&c.cursor)
	c.CmdClause.Flag("format", "Output format (json)").Hidden().EnumVar(&c.formatFlag, "json")
	c.CmdClause.Flag("json", argparser.FlagJSONDesc).Short('j').BoolVar(&c.jsonFlag)

	return &c
}

// Exec implements the command interface.
func (c *DomainInspectorCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := resolveJSONFormat(&c.formatFlag, c.jsonFlag, c.Globals); err != nil {
		return err
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := fastly.GetDomainMetricsInput{
		ServiceID:   serviceID,
		Datacenters: c.datacenters,
		Domains:     c.domains,
		GroupBy:     c.groupBy,
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
