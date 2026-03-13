package text

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"
)

func PrintUsageTbl(out io.Writer, data *fastly.RegionsUsage) {
	tbl := NewTable(out)
	tbl.AddHeader("REGION", "BANDWIDTH", "REQUESTS", "COMPUTE REQUESTS")
	if data == nil {
		tbl.Print()
		return
	}
	for _, region := range slices.Sorted(maps.Keys(*data)) {
		u := (*data)[region]
		if u == nil {
			continue
		}
		tbl.AddLine(region, fastly.ToValue(u.Bandwidth), fastly.ToValue(u.Requests), fastly.ToValue(u.ComputeRequests))
	}
	tbl.Print()
}

func PrintUsageByServiceTbl(out io.Writer, data *fastly.ServicesByRegionsUsage) {
	tbl := NewTable(out)
	tbl.AddHeader("REGION", "SERVICE", "BANDWIDTH", "REQUESTS", "COMPUTE REQUESTS")
	if data == nil {
		tbl.Print()
		return
	}
	for _, region := range slices.Sorted(maps.Keys(*data)) {
		services := (*data)[region]
		if services == nil {
			continue
		}
		for _, svcID := range slices.Sorted(maps.Keys(*services)) {
			u := (*services)[svcID]
			if u == nil {
				continue
			}
			tbl.AddLine(region, svcID, fastly.ToValue(u.Bandwidth), fastly.ToValue(u.Requests), fastly.ToValue(u.ComputeRequests))
		}
	}
	tbl.Print()
}

func PrintDomainInspectorTbl(out io.Writer, resp *fastly.DomainInspector) {
	if resp.Meta != nil {
		if resp.Meta.Start != nil {
			fmt.Fprintf(out, "Start: %s\n", *resp.Meta.Start)
		}
		if resp.Meta.End != nil {
			fmt.Fprintf(out, "End: %s\n", *resp.Meta.End)
		}
		fmt.Fprintln(out, "---")
	}

	dimKeys := domainDimensionKeys(resp.Data)
	header := make([]any, 0, len(dimKeys)+5)
	for _, k := range dimKeys {
		header = append(header, strings.ToUpper(k))
	}
	header = append(header, "TIMESTAMP", "REQUESTS", "BANDWIDTH", "EDGE REQUESTS", "EDGE HIT RATIO")

	tbl := NewTable(out)
	tbl.AddHeader(header...)
	for _, d := range resp.Data {
		for _, v := range d.Values {
			row := make([]any, 0, len(dimKeys)+5)
			for _, k := range dimKeys {
				row = append(row, fastly.ToValue(d.Dimensions[k]))
			}
			ts := ""
			if v.Timestamp != nil {
				ts = time.Unix(int64(*v.Timestamp), 0).UTC().String() //nolint:gosec
			}
			row = append(row, ts, fastly.ToValue(v.Requests), fastly.ToValue(v.Bandwidth), fastly.ToValue(v.EdgeRequests), fmt.Sprintf("%.4f", fastly.ToValue(v.EdgeHitRatio)))
			tbl.AddLine(row...)
		}
	}
	tbl.Print()

	if resp.Meta != nil && resp.Meta.NextCursor != nil {
		fmt.Fprintf(out, "Next cursor: %s\n", *resp.Meta.NextCursor)
	}
}

func PrintOriginInspectorTbl(out io.Writer, resp *fastly.OriginInspector) {
	if resp.Meta != nil {
		if resp.Meta.Start != nil {
			fmt.Fprintf(out, "Start: %s\n", *resp.Meta.Start)
		}
		if resp.Meta.End != nil {
			fmt.Fprintf(out, "End: %s\n", *resp.Meta.End)
		}
		fmt.Fprintln(out, "---")
	}

	dimKeys := originDimensionKeys(resp.Data)
	header := make([]any, 0, len(dimKeys)+5)
	for _, k := range dimKeys {
		header = append(header, strings.ToUpper(k))
	}
	header = append(header, "TIMESTAMP", "RESPONSES", "STATUS 2XX", "STATUS 4XX", "STATUS 5XX")

	tbl := NewTable(out)
	tbl.AddHeader(header...)
	for _, d := range resp.Data {
		for _, v := range d.Values {
			row := make([]any, 0, len(dimKeys)+5)
			for _, k := range dimKeys {
				row = append(row, d.Dimensions[k])
			}
			ts := ""
			if v.Timestamp != nil {
				ts = time.Unix(int64(*v.Timestamp), 0).UTC().String() //nolint:gosec
			}
			row = append(row, ts, fastly.ToValue(v.Responses), fastly.ToValue(v.Status2xx), fastly.ToValue(v.Status4xx), fastly.ToValue(v.Status5xx))
			tbl.AddLine(row...)
		}
	}
	tbl.Print()

	if resp.Meta != nil && resp.Meta.NextCursor != nil {
		fmt.Fprintf(out, "Next cursor: %s\n", *resp.Meta.NextCursor)
	}
}

func domainDimensionKeys(data []*fastly.DomainData) []string {
	seen := make(map[string]struct{})
	for _, d := range data {
		for k := range d.Dimensions {
			seen[k] = struct{}{}
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

func originDimensionKeys(data []*fastly.OriginData) []string {
	seen := make(map[string]struct{})
	for _, d := range data {
		for k := range d.Dimensions {
			seen[k] = struct{}{}
		}
	}
	return slices.Sorted(maps.Keys(seen))
}
