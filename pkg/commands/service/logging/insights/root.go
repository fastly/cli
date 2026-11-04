package insights

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
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

	// Required.
	Start         string `required:"" help:"Inclusive start time in RFC3339 format"`
	End           string `required:"" help:"Exclusive end time in RFC3339 format"`
	Visualization string `required:"" help:"Log Insights visualization to retrieve"`

	// Optional.
	ServiceID        *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName      argparser.OptionalServiceNameID `help:"The name of the service"`
	Domain           argparser.OptionalString        `help:"Limit data to the specified request domain"`
	DomainExactMatch bool                            `help:"Treat --domain as an exact match"`
	Limit            argparser.OptionalInt           `help:"Maximum number of rows to return (up to 100)"`
	Pops             argparser.OptionalString        `help:"Comma-separated list of Fastly POP codes"`

	argparser.JSONOutput
}

// AfterApply checks the values kong cannot check by itself, then hands the
// parsed flags over to the fields the command reads them from.
func (c *Command) AfterApply() error {
	if !slices.Contains(visualizationOptions, c.Visualization) {
		return fmt.Errorf("enum value must be one of %s, got '%s'", strings.Join(visualizationOptions, ","), c.Visualization)
	}
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the Log Insights API.
func (c *Command) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := &fastly.GetLogInsightsInput{
		ServiceID:     serviceID,
		Start:         c.Start,
		End:           c.End,
		Visualization: fastly.LogInsightsVisualization(c.Visualization),
	}
	if c.Domain.WasSet {
		input.Domain = &c.Domain.Value
	}
	if c.DomainExactMatch {
		input.DomainExactMatch = new(true)
	}
	if c.Limit.WasSet {
		input.Limit = &c.Limit.Value
	}
	if c.Pops.WasSet {
		input.POPs = splitPOPs(c.Pops.Value)
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

// Hints returns the completion hints of the flags whose values are only
// known at run time, as kingpin's HintOptions gave them.
func (c *Command) Hints() map[string][]string {
	return map[string][]string{
		"visualization": visualizationOptions,
	}
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
