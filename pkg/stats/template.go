package stats

import (
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/fastly/go-fastly/fastly"
	"github.com/mitchellh/mapstructure"
)

var blockTemplate = template.Must(template.New("stats_block").Parse(
	`Service ID:         {{ .ServiceID }}
Start Time:         {{ .StartTime }}
--------------------------------------------------
Hit Rate:           {{ .HitRate }}
Avg Hit Time:       {{ .AvgHitTime }}
Avg Miss Time:      {{ .AvgMissTime }}

Request BW:         {{ .RequestBytes }}
  Headers:          {{ .RequestHeaderBytes }}
  Body:             {{ .RequestBodyBytes }}

Response BW:        {{ .ResponseBytes }}
  Headers:          {{ .ResponseHeaderBytes }}
  Body:             {{ .ResponseBodyBytes }}

Requests:           {{ .RequestCount }}
  Hit:              {{ .Hits }}
  Miss:             {{ .Miss }}
  Pass:             {{ .Pass }}
  Synth:            {{ .Synth }}
  Error:            {{ .Errors }}
  Uncacheable:      {{ .Uncacheable }}

`))

func fmtBlock(out io.Writer, service string, block statsResponseData) error {
	var agg fastly.Stats
	if err := mapstructure.Decode(block, &agg); err != nil {
		return err
	}

	hitRate := 0.0
	if agg.Hits > 0 {
		hitRate = float64((agg.Hits - agg.Miss - agg.Errors)) / float64(agg.Hits)
	}

	// TODO: parse the JSON more strictly so this doesn't need to be dynamic.
	startTime := time.Unix(int64(block["start_time"].(float64)), 0)

	values := map[string]string{
		"ServiceID":   fmt.Sprintf("%30s", service),
		"StartTime":   fmt.Sprintf("%30s", startTime),
		"HitRate":     fmt.Sprintf("%29.2f%%", hitRate*100),
		"AvgHitTime":  fmt.Sprintf("%28.2f\u00b5s", agg.HitsTime*1000),
		"AvgMissTime": fmt.Sprintf("%28.2f\u00b5s", agg.MissTime*1000),

		"RequestBytes":        fmt.Sprintf("%30d", agg.RequestHeaderBytes+agg.RequestBodyBytes),
		"RequestHeaderBytes":  fmt.Sprintf("%30d", agg.RequestHeaderBytes),
		"RequestBodyBytes":    fmt.Sprintf("%30d", agg.RequestBodyBytes),
		"ResponseBytes":       fmt.Sprintf("%30d", agg.ResponseHeaderBytes+agg.ResponseBodyBytes),
		"ResponseHeaderBytes": fmt.Sprintf("%30d", agg.ResponseHeaderBytes),
		"ResponseBodyBytes":   fmt.Sprintf("%30d", agg.ResponseBodyBytes),

		"RequestCount": fmt.Sprintf("%30d", agg.Requests),
		"Hits":         fmt.Sprintf("%30d", agg.Hits),
		"Miss":         fmt.Sprintf("%30d", agg.Miss),
		"Pass":         fmt.Sprintf("%30d", agg.Pass),
		"Synth":        fmt.Sprintf("%30d", agg.Synth),
		"Errors":       fmt.Sprintf("%30d", agg.Errors),
		"Uncacheable":  fmt.Sprintf("%30d", agg.Uncachable)}

	return blockTemplate.Execute(out, values)
}
