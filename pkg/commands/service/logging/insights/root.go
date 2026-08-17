package insights

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CommandName is the string used to invoke the Log Insights command.
const CommandName = "insights"

// visualizationOptions is a string representation of the visualizations
// supported by go-fastly, suitable for use with the enum flag below.
var visualizationOptions = func() (visualizations []string) {
	for _, visualization := range fastly.LogInsightsVisualizations {
		visualizations = append(visualizations, string(visualization))
	}
	return visualizations
}()

// Command exposes the Log Insights API.
type Command struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	start         string
	end           string
	visualization string

	// Optional.
	serviceName      argparser.OptionalServiceNameID
	domain           argparser.OptionalString
	domainExactMatch argparser.OptionalString
	limit            argparser.OptionalInt
	pops             argparser.OptionalString
}

// NewInsightsCommand returns a usable Log Insights command registered under the parent.
func NewInsightsCommand(parent argparser.Registerer, g *global.Data) *Command {
	c := Command{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command(CommandName, "Retrieve statistics from sampled log records")

	// Required.
	c.CmdClause.Flag("start", "Inclusive start time in RFC3339 format").Required().StringVar(&c.start)
	c.CmdClause.Flag("end", "Exclusive end time in RFC3339 format").Required().StringVar(&c.end)
	c.CmdClause.Flag("visualization", "Log Insights visualization to retrieve").
		Required().
		HintOptions(visualizationOptions...).
		EnumVar(&c.visualization, visualizationOptions...)

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
	c.CmdClause.Flag("domain", "Limit data to the specified request domain").Action(c.domain.Set).StringVar(&c.domain.Value)
	c.CmdClause.Flag("domain-exact-match", "Treat --domain as an exact match instead of a suffix match [true, false]").Action(c.domainExactMatch.Set).StringVar(&c.domainExactMatch.Value)
	c.CmdClause.Flag("limit", "Maximum number of rows to return (up to 100)").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("pops", "Comma-separated list of Fastly POP codes").Action(c.pops.Set).StringVar(&c.pops.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the Log Insights API.
func (c *Command) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := &fastly.GetLogInsightsInput{
		ServiceID:     serviceID,
		Start:         c.start,
		End:           c.end,
		Visualization: fastly.LogInsightsVisualization(c.visualization),
	}
	if c.domain.WasSet {
		input.Domain = &c.domain.Value
	}
	if c.domainExactMatch.WasSet {
		domainExactMatch, err := argparser.ConvertBoolFromStringFlag(c.domainExactMatch.Value, "domain-exact-match")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.DomainExactMatch = domainExactMatch
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.pops.WasSet {
		input.POPs = splitPOPs(c.pops.Value)
	}

	result, err := c.Globals.APIClient.GetLogInsights(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{"Service ID": serviceID})
		return err
	}

	if ok, err := c.WriteJSON(out, result); ok {
		return err
	}

	printLogInsights(out, result)
	return nil
}

func splitPOPs(value string) []string {
	var result []string
	for _, pop := range strings.Split(value, ",") {
		if pop = strings.TrimSpace(pop); pop != "" {
			result = append(result, pop)
		}
	}
	return result
}

func printLogInsights(out io.Writer, result *fastly.LogInsightsResponse) {
	if result == nil || len(result.Data) == 0 {
		fmt.Fprintln(out, "No log insights found.")
		return
	}

	table := text.NewTable(out)
	table.AddHeader("DIMENSIONS", "VALUES")

	for _, data := range result.Data {
		if data == nil {
			continue
		}

		dimensions := formatLogInsightsDimensions(data.Dimensions)
		if len(data.Values) == 0 {
			table.AddLine(dimensions, "-")
			continue
		}

		for _, value := range data.Values {
			table.AddLine(dimensions, formatLogInsightsValue(value))
		}
	}

	table.Print()
}

func formatLogInsightsDimensions(d *fastly.LogInsightsDimensions) string {
	if d == nil {
		return "-"
	}

	var values []string
	appendString := func(name string, value *string) {
		if value != nil {
			values = append(values, name+"="+*value)
		}
	}

	appendString("browser", d.Browser)
	appendString("browser_version", d.BrowserVersion)
	appendString("content_type", d.ContentType)
	appendString("country", d.Country)
	appendString("device", d.Device)
	appendString("os", d.OS)
	appendString("region", d.Region)
	appendString("response", d.Response)
	appendString("status-code", d.StatusCode)
	appendString("url", d.URL)

	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func formatLogInsightsValue(v *fastly.LogInsightsValue) string {
	if v == nil {
		return "-"
	}

	var values []string
	appendFloat := func(name string, value *float64) {
		if value != nil {
			values = append(values, name+"="+strconv.FormatFloat(*value, 'g', -1, 64))
		}
	}

	appendFloat("average_bandwidth_bytes", v.AverageBandwidthBytes)
	appendFloat("average_response_time", v.AverageResponseTime)
	appendFloat("bandwidth_percentage", v.BandwidthPercentage)
	appendFloat("cache_hit_ratio", v.CacheHitRatio)
	appendFloat("country_chr", v.CountryCHR)
	appendFloat("country_error_rate", v.CountryErrorRate)
	appendFloat("country_request_rate", v.CountryRequestRate)
	appendFloat("miss_rate", v.MissRate)
	appendFloat("p95_response_time", v.P95ResponseTime)
	appendFloat("rate", v.Rate)
	appendFloat("503_rate_per_url", v.Rate503PerURL)
	appendFloat("rate_per_status", v.RatePerStatus)
	appendFloat("rate_per_url", v.RatePerURL)
	appendFloat("region_chr", v.RegionCHR)
	appendFloat("region_error_rate", v.RegionErrorRate)
	appendFloat("request_percentage", v.RequestPercentage)
	appendFloat("response_time_percentage", v.ResponseTimePercentage)

	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}
