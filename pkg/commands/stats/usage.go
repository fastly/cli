package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UsageCommand exposes the Usage Stats API.
type UsageCommand struct {
	argparser.Base

	by         string
	byService  bool
	formatFlag string
	from       string
	region     string
	to         string
}

// NewUsageCommand is the "stats usage" subcommand.
func NewUsageCommand(parent argparser.Registerer, g *global.Data) *UsageCommand {
	var c UsageCommand
	c.Globals = g

	c.CmdClause = parent.Command("usage", "View usage stats (bandwidth, requests)")

	// Optional.
	c.CmdClause.Flag("from", "Start time").StringVar(&c.from)
	c.CmdClause.Flag("to", "End time").StringVar(&c.to)
	c.CmdClause.Flag("by", "Aggregation period (minute/hour/day)").EnumVar(&c.by, "minute", "hour", "day")
	c.CmdClause.Flag("region", "Filter by region ('stats regions' to list)").StringVar(&c.region)
	c.CmdClause.Flag("by-service", "Break down usage by service").BoolVar(&c.byService)
	c.CmdClause.Flag("format", "Output format (json)").EnumVar(&c.formatFlag, "json")

	return &c
}

// Exec implements the command interface.
func (c *UsageCommand) Exec(_ io.Reader, out io.Writer) error {
	input := fastly.GetUsageInput{}
	if c.by != "" {
		input.By = &c.by
	}
	if c.from != "" {
		input.From = &c.from
	}
	if c.region != "" {
		input.Region = &c.region
	}
	if c.to != "" {
		input.To = &c.to
	}

	if c.byService {
		return c.execByService(out, &input)
	}
	return c.execPlain(out, &input)
}

func (c *UsageCommand) execPlain(out io.Writer, input *fastly.GetUsageInput) error {
	resp, err := c.Globals.APIClient.GetUsage(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if fastly.ToValue(resp.Status) != statusSuccess {
		return fmt.Errorf("non-success response: %s", fastly.ToValue(resp.Message))
	}

	switch c.formatFlag {
	case "json":
		return writeUsageJSON(out, resp.Data)
	default:
		return writeUsageTable(out, resp.Data)
	}
}

func (c *UsageCommand) execByService(out io.Writer, input *fastly.GetUsageInput) error {
	resp, err := c.Globals.APIClient.GetUsageByService(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if fastly.ToValue(resp.Status) != statusSuccess {
		return fmt.Errorf("non-success response: %s", fastly.ToValue(resp.Message))
	}

	namesByID := c.resolveServiceNames(resp.Data)

	switch c.formatFlag {
	case "json":
		return writeUsageByServiceJSON(out, resp.Data, namesByID)
	default:
		return writeUsageByServiceTable(out, resp.Data, namesByID)
	}
}

// resolveServiceNames builds a map of service ID to service name by querying
// the API for each unique service ID found in the usage data.
func (c *UsageCommand) resolveServiceNames(data *fastly.ServicesByRegionsUsage) map[string]string {
	namesByID := make(map[string]string)
	if data == nil {
		return namesByID
	}
	// Collect unique service IDs across all regions.
	seen := make(map[string]bool)
	for _, services := range *data {
		if services == nil {
			continue
		}
		for svcID := range *services {
			seen[svcID] = true
		}
	}
	for svcID := range seen {
		svc, err := c.Globals.APIClient.GetService(context.TODO(), &fastly.GetServiceInput{
			ServiceID: svcID,
		})
		if err == nil && svc != nil {
			namesByID[svcID] = fastly.ToValue(svc.Name)
		}
	}
	return namesByID
}

func writeUsageTable(out io.Writer, data *fastly.RegionsUsage) error {
	if data == nil {
		return nil
	}
	regions := slices.Sorted(maps.Keys(*data))
	for _, region := range regions {
		usage := (*data)[region]
		if usage == nil {
			continue
		}
		text.Output(out, "Region: %s", region)
		text.Output(out, "  Bandwidth:        %d", fastly.ToValue(usage.Bandwidth))
		text.Output(out, "  Requests:         %d", fastly.ToValue(usage.Requests))
		text.Output(out, "  Compute Requests: %d", fastly.ToValue(usage.ComputeRequests))
	}
	return nil
}

func writeUsageJSON(out io.Writer, data *fastly.RegionsUsage) error {
	if data == nil {
		return json.NewEncoder(out).Encode(map[string]any{})
	}
	return json.NewEncoder(out).Encode(usageToMap(*data))
}

func writeUsageByServiceTable(out io.Writer, data *fastly.ServicesByRegionsUsage, namesByID map[string]string) error {
	if data == nil {
		return nil
	}
	regions := slices.Sorted(maps.Keys(*data))
	for _, region := range regions {
		services := (*data)[region]
		if services == nil {
			continue
		}
		text.Output(out, "Region: %s", region)
		serviceIDs := slices.Sorted(maps.Keys(*services))
		for _, svcID := range serviceIDs {
			usage := (*services)[svcID]
			if usage == nil {
				continue
			}
			if name, ok := namesByID[svcID]; ok && name != "" {
				text.Output(out, "  Service: %s (%s)", name, svcID)
			} else {
				text.Output(out, "  Service: %s", svcID)
			}
			text.Output(out, "    Bandwidth:        %d", fastly.ToValue(usage.Bandwidth))
			text.Output(out, "    Requests:         %d", fastly.ToValue(usage.Requests))
			text.Output(out, "    Compute Requests: %d", fastly.ToValue(usage.ComputeRequests))
		}
	}
	return nil
}

func writeUsageByServiceJSON(out io.Writer, data *fastly.ServicesByRegionsUsage, namesByID map[string]string) error {
	if data == nil {
		return json.NewEncoder(out).Encode(map[string]any{})
	}
	result := make(map[string]any)
	for region, services := range *data {
		if services == nil {
			continue
		}
		regionMap := make(map[string]any)
		for svcID, usage := range *services {
			entry := usageEntry(usage)
			entry["service_id"] = svcID
			if name, ok := namesByID[svcID]; ok && name != "" {
				entry["service_name"] = name
			}
			regionMap[svcID] = entry
		}
		result[region] = regionMap
	}
	return json.NewEncoder(out).Encode(result)
}

func usageToMap(data fastly.RegionsUsage) map[string]any {
	result := make(map[string]any)
	for region, usage := range data {
		result[region] = usageEntry(usage)
	}
	return result
}

func usageEntry(u *fastly.Usage) map[string]any {
	if u == nil {
		return map[string]any{
			"bandwidth":        uint64(0),
			"requests":         uint64(0),
			"compute_requests": uint64(0),
		}
	}
	return map[string]any{
		"bandwidth":        fastly.ToValue(u.Bandwidth),
		"requests":         fastly.ToValue(u.Requests),
		"compute_requests": fastly.ToValue(u.ComputeRequests),
	}
}
