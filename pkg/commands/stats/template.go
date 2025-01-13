package stats

import (
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"
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
	aggHits := fastly.ToValue(agg.Hits)
	aggMiss := fastly.ToValue(agg.Miss)
	aggErrs := fastly.ToValue(agg.Errors)
	if aggHits > 0 {
		hitRate = float64(aggHits-aggMiss-aggErrs) / float64(aggHits)
	}

	// TODO: parse the JSON more strictly so this doesn't need to be dynamic.
	st, ok := block["start_time"].(float64)
	if !ok {
		return fmt.Errorf("failed to type assert '%v' to a float64", block["start_time"])
	}
	startTime := time.Unix(int64(st), 0).UTC()

	values := map[string]string{
		"ServiceID":   fmt.Sprintf("%30s", service),
		"StartTime":   fmt.Sprintf("%30s", startTime),
		"HitRate":     fmt.Sprintf("%29.2f%%", hitRate*100),
		"AvgHitTime":  fmt.Sprintf("%28.2f\u00b5s", fastly.ToValue(agg.HitsTime)*1000),
		"AvgMissTime": fmt.Sprintf("%28.2f\u00b5s", fastly.ToValue(agg.MissTime)*1000),

		"RequestBytes":        fmt.Sprintf("%30d", fastly.ToValue(agg.RequestHeaderBytes)+fastly.ToValue(agg.RequestBodyBytes)),
		"RequestHeaderBytes":  fmt.Sprintf("%30d", fastly.ToValue(agg.RequestHeaderBytes)),
		"RequestBodyBytes":    fmt.Sprintf("%30d", fastly.ToValue(agg.RequestBodyBytes)),
		"ResponseBytes":       fmt.Sprintf("%30d", fastly.ToValue(agg.ResponseHeaderBytes)+fastly.ToValue(agg.ResponseBodyBytes)),
		"ResponseHeaderBytes": fmt.Sprintf("%30d", fastly.ToValue(agg.ResponseHeaderBytes)),
		"ResponseBodyBytes":   fmt.Sprintf("%30d", fastly.ToValue(agg.ResponseBodyBytes)),

		"RequestCount": fmt.Sprintf("%30d", fastly.ToValue(agg.Requests)),
		"Hits":         fmt.Sprintf("%30d", aggHits),
		"Miss":         fmt.Sprintf("%30d", aggMiss),
		"Pass":         fmt.Sprintf("%30d", fastly.ToValue(agg.Pass)),
		"Synth":        fmt.Sprintf("%30d", fastly.ToValue(agg.Synth)),
		"Errors":       fmt.Sprintf("%30d", aggErrs),
		"Uncacheable":  fmt.Sprintf("%30d", fastly.ToValue(agg.Uncachable)),
	}

	return blockTemplate.Execute(out, values)
}
