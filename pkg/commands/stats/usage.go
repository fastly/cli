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

// UsageCommand exposes the Usage Stats API.
type UsageCommand struct {
	argparser.Base

	// Optional.
	From       string `help:"Start time"`
	To         string `help:"End time"`
	By         string `default:"" enum:"minute,hour,day," help:"Aggregation period (minute/hour/day)"`
	Region     string `help:"Filter by region ('stats regions' to list)"`
	ByService  bool   `help:"Break down usage by service"`
	FormatFlag string `name:"format" default:"" enum:"json," hidden:"" help:"Output format (json)"`
	JSONFlag   bool   `name:"json" short:"j" help:"Render output as JSON"`
}

// Exec implements the command interface.
func (c *UsageCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := resolveJSONFormat(&c.FormatFlag, c.JSONFlag, c.Globals); err != nil {
		return err
	}

	input := fastly.GetUsageInput{}
	if c.By != "" {
		input.By = &c.By
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

	if c.ByService {
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

	filterUsageByRegion(resp.Data, c.Region)

	switch c.FormatFlag {
	case "json":
		return writeUsageJSON(out, resp.Data)
	default:
		text.PrintUsageTbl(out, resp.Data)
		return nil
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

	filterUsageByServiceByRegion(resp.Data, c.Region)

	switch c.FormatFlag {
	case "json":
		return writeUsageByServiceJSON(out, resp.Data)
	default:
		text.PrintUsageByServiceTbl(out, resp.Data)
		return nil
	}
}

func writeUsageJSON(out io.Writer, data *fastly.RegionsUsage) error {
	if data == nil {
		return json.NewEncoder(out).Encode(map[string]any{})
	}
	return json.NewEncoder(out).Encode(usageToMap(*data))
}

func writeUsageByServiceJSON(out io.Writer, data *fastly.ServicesByRegionsUsage) error {
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
			regionMap[svcID] = usageEntry(usage)
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

func filterUsageByRegion(data *fastly.RegionsUsage, region string) {
	if region == "" || data == nil {
		return
	}
	for k := range *data {
		if k != region {
			delete(*data, k)
		}
	}
}

func filterUsageByServiceByRegion(data *fastly.ServicesByRegionsUsage, region string) {
	if region == "" || data == nil {
		return
	}
	for k := range *data {
		if k != region {
			delete(*data, k)
		}
	}
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
