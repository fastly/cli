package stats

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/fastly/go-fastly/fastly"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/mitchellh/mapstructure"
)

// Maximum number of time slots to remember (1 sec per spot)
const maxSlots = 60

type View struct {
	Events  <-chan ui.Event
	Grid    *ui.Grid
	Service string

	// Header
	Header *widgets.Paragraph

	// Top Row
	Hits     *widgets.Paragraph
	HitTime  *widgets.Paragraph
	Misses   *widgets.Paragraph
	HitRatio *widgets.PieChart
	MissTime *widgets.Paragraph
	Requests *widgets.Paragraph
	Errors   *widgets.Paragraph

	// Second Row
	GlobalLabel *widgets.Paragraph
	Global      *widgets.BarChart
	GlobalData  []float64

	// Third Row
	ReqLabel       *widgets.Paragraph
	ReqChart       *widgets.StackedBarChart
	ReqData        [][]float64
	BandwidthLabel *widgets.Paragraph
	BandwidthGroup *widgets.SparklineGroup
	BandwidthChart *widgets.Sparkline
	BandwidthData  []float64

	// Fourth Row
	ErrLabel *widgets.Paragraph
	ErrGroup *widgets.SparklineGroup
	ErrChart *widgets.Sparkline
	ErrData  []float64
	HitLabel *widgets.Paragraph
	HitGroup *widgets.SparklineGroup
	HitChart *widgets.Sparkline
	HitData  []float64

	// Fifth Row
	IoLabel      *widgets.Paragraph
	IoGroup      *widgets.SparklineGroup
	IoChart      *widgets.Sparkline
	IoData       []float64
	LoggingLabel *widgets.Paragraph
	LoggingGroup *widgets.SparklineGroup
	LoggingChart *widgets.Sparkline
	LoggingData  []float64
}

func NewView(service string) (*View, error) {
	var view = View{}
	view.Service = service

	if err := ui.Init(); err != nil {
		return nil, err
	}

	view.Events = ui.PollEvents()
	view.Grid = ui.NewGrid()

	view.Header = widgets.NewParagraph()
	view.Header.Title = "Fastly Real Time Stats"

	// Top Row
	view.Hits = widgets.NewParagraph()
	view.Hits.Title = "Hits"

	view.HitTime = widgets.NewParagraph()
	view.HitTime.Title = "Hit Time"

	view.Misses = widgets.NewParagraph()
	view.Misses.Title = "Misses"

	view.HitRatio = widgets.NewPieChart()
	view.HitRatio.Title = "Hit Ratio"
	view.HitRatio.AngleOffset = -.5 * math.Pi
	view.HitRatio.LabelFormatter = func(i int, v float64) string {
		if v > 0.0 {
			return fmt.Sprintf("%.01f%%", v*100)
		}
		return ""
	}

	view.MissTime = widgets.NewParagraph()
	view.MissTime.Title = "Miss Time"

	view.Requests = widgets.NewParagraph()
	view.Requests.Title = "Requests"

	view.Errors = widgets.NewParagraph()
	view.Errors.Title = "Errors"

	// Second Row
	view.GlobalLabel = widgets.NewParagraph()
	view.GlobalLabel.Title = "Global POP Traffic"
	view.Global = widgets.NewBarChart()
	view.Global.NumFormatter = func(v float64) string { return "" }

	// Third Row
	view.ReqLabel = widgets.NewParagraph()
	view.ReqLabel.Title = "Requests"
	view.ReqChart = widgets.NewStackedBarChart()
	view.ReqChart.NumFormatter = func(v float64) string { return "" }
	// Stacked bar charts are too wide, so only do the last 30 seconds
	view.ReqData = make([][]float64, maxSlots/2)
	for i := range view.ReqData {
		view.ReqData[i] = make([]float64, 5)
	}

	view.BandwidthLabel = widgets.NewParagraph()
	view.BandwidthLabel.Title = "Bandwidth"
	view.BandwidthLabel.BorderStyle.Fg = ui.ColorMagenta
	view.BandwidthChart = widgets.NewSparkline()
	view.BandwidthChart.LineColor = view.BandwidthLabel.BorderStyle.Fg
	view.BandwidthGroup = widgets.NewSparklineGroup(view.BandwidthChart)
	view.BandwidthGroup.BorderStyle.Fg = view.BandwidthLabel.BorderStyle.Fg
	view.BandwidthData = make([]float64, maxSlots)

	// Fourth Row
	view.ErrLabel = widgets.NewParagraph()
	view.ErrLabel.Title = "Errors"
	view.ErrLabel.BorderStyle.Fg = ui.ColorRed
	view.ErrChart = widgets.NewSparkline()
	view.ErrChart.LineColor = view.ErrLabel.BorderStyle.Fg
	view.ErrGroup = widgets.NewSparklineGroup(view.ErrChart)
	view.ErrGroup.BorderStyle.Fg = view.ErrLabel.BorderStyle.Fg
	view.ErrData = make([]float64, maxSlots)

	view.HitLabel = widgets.NewParagraph()
	view.HitLabel.Title = "Hit Ratio"
	view.HitLabel.BorderStyle.Fg = ui.ColorGreen
	view.HitChart = widgets.NewSparkline()
	view.HitChart.LineColor = view.HitLabel.BorderStyle.Fg
	view.HitGroup = widgets.NewSparklineGroup(view.HitChart)
	view.HitGroup.BorderStyle.Fg = view.HitLabel.BorderStyle.Fg
	view.HitData = make([]float64, maxSlots)

	// Fifth Row
	view.IoLabel = widgets.NewParagraph()
	view.IoLabel.Title = "Image Optimizer"
	view.IoLabel.BorderStyle.Fg = ui.ColorYellow
	view.IoChart = widgets.NewSparkline()
	view.IoChart.LineColor = view.IoLabel.BorderStyle.Fg
	view.IoGroup = widgets.NewSparklineGroup(view.IoChart)
	view.IoGroup.BorderStyle.Fg = view.IoLabel.BorderStyle.Fg
	view.IoData = make([]float64, maxSlots)

	view.LoggingLabel = widgets.NewParagraph()
	view.LoggingLabel.Title = "Logs Sent"
	view.LoggingLabel.BorderStyle.Fg = ui.ColorBlue
	view.LoggingChart = widgets.NewSparkline()
	view.LoggingChart.LineColor = view.LoggingLabel.BorderStyle.Fg
	view.LoggingGroup = widgets.NewSparklineGroup(view.LoggingChart)
	view.LoggingGroup.BorderStyle.Fg = view.LoggingLabel.BorderStyle.Fg
	view.LoggingData = make([]float64, maxSlots)

	return &view, nil
}

func (v *View) SetLayout() {
	v.Grid.Set(
		ui.NewRow(0.05, v.Header),
		ui.NewRow(0.14,
			ui.NewCol(1.0/7, v.Hits),
			ui.NewCol(1.0/7, v.HitTime),
			ui.NewCol(1.0/7, v.Misses),
			ui.NewCol(1.0/7, v.HitRatio),
			ui.NewCol(1.0/7, v.MissTime),
			ui.NewCol(1.0/7, v.Requests),
			ui.NewCol(1.0/7, v.Errors),
		),
		ui.NewRow(0.15,
			ui.NewRow(0.2, v.GlobalLabel), ui.NewRow(0.8, v.Global),
		),
		ui.NewRow(0.22,
			ui.NewCol(1.0/2, ui.NewRow(0.2, v.ReqLabel), ui.NewRow(0.8, v.ReqChart)),
			ui.NewCol(1.0/2, ui.NewRow(0.2, v.BandwidthLabel), ui.NewRow(0.8, v.BandwidthGroup)),
		),
		ui.NewRow(0.22,
			ui.NewCol(1.0/2, ui.NewRow(0.2, v.ErrLabel), ui.NewRow(0.8, v.ErrGroup)),
			ui.NewCol(1.0/2, ui.NewRow(0.2, v.HitLabel), ui.NewRow(0.8, v.HitGroup)),
		),
		ui.NewRow(0.22,
			ui.NewCol(1.0/2, ui.NewRow(0.2, v.IoLabel), ui.NewRow(0.8, v.IoGroup)),
			ui.NewCol(1.0/2, ui.NewRow(0.2, v.LoggingLabel), ui.NewRow(0.8, v.LoggingGroup)),
		),
	)
}

func (v *View) Resize() {
	termWidth, termHeight := ui.TerminalDimensions()
	v.Grid.SetRect(0, 0, termWidth, termHeight)
}

func (v *View) Render() {
	ui.Clear()
	ui.Render(v.Grid)
}

func (v *View) UpdateStats(block realtimeResponseData) error {
	agg := block.Aggregated
	delete(agg, "miss_histogram")

	var s fastly.Stats
	if err := mapstructure.Decode(agg, &s); err != nil {
		return err
	}

	// Header
	startTime := time.Unix(int64(block.Recorded), 0).UTC()
	v.Header.Text = fmt.Sprintf("Stats for %s @ %s", v.Service, startTime)

	// Top Row
	v.Hits.Text = fmt.Sprintf("%30d\nper second", s.Hits)
	v.HitTime.Text = fmt.Sprintf("%28.2f\u00b5s\n(avg)", s.HitsTime*1000)
	v.Misses.Text = fmt.Sprintf("%30d\nper second", s.Miss)

	hitRate := 0.0
	if s.Hits > 0 {
		hitRate = float64((s.Hits - s.Miss - s.Errors)) / float64(s.Hits)
	}
	v.HitRatio.Data = []float64{hitRate, 1.0 - hitRate}

	v.MissTime.Text = fmt.Sprintf("%28.2f\u00b5s\n(avg)", s.MissTime*1000)
	v.Requests.Text = fmt.Sprintf("%30d\nper second", s.Requests)
	v.Errors.Text = fmt.Sprintf("%30d\nper second", s.Errors)

	// Second Row
	var dcLabels []string
	var dcReqs []float64
	for dc := range block.Datacenter {
		if len(dc) > 3 {
			continue
		}
		dcLabels = append(dcLabels, dc)
	}
	sort.Strings(dcLabels)
	for _, dc := range dcLabels {
		dcReqs = append(dcReqs, block.Datacenter[dc]["requests"].(float64))
	}

	v.Global.Labels = dcLabels
	v.Global.Data = dcReqs

	// Third Row
	reqNow := []float64{float64(s.Errors), float64(s.Miss), float64(s.Hits), float64(s.Synth), float64(s.Pass)}
	v.ReqLabel.Text = humanize.SI(float64(s.Requests), "")
	v.ReqData = append(v.ReqData, reqNow)[1:]
	v.ReqChart.Data = v.ReqData
	v.BandwidthLabel.Text = humanize.Bytes(s.ResponseHeaderBytes + s.ResponseBodyBytes)
	v.BandwidthData = append(v.BandwidthData, float64(s.ResponseHeaderBytes+s.ResponseBodyBytes))[1:]
	v.BandwidthChart.Data = v.BandwidthData

	// Fourth Row
	v.ErrLabel.Text = fmt.Sprintf("%30d/second", s.Errors)
	v.ErrData = append(v.ErrData, float64(s.Errors))[1:]
	v.ErrChart.Data = v.ErrData
	v.HitLabel.Text = fmt.Sprintf("%3.1f%%", hitRate*100)
	v.HitData = append(v.HitData, hitRate)[1:]
	v.HitChart.Data = v.HitData

	// Fifth Row
	v.IoLabel.Text = fmt.Sprintf("%30d/second", s.ImageOptimizer)
	v.IoData = append(v.IoData, float64(s.ImageOptimizer))[1:]
	v.IoChart.Data = v.IoData
	v.LoggingLabel.Text = fmt.Sprintf("%30d/second", s.Log)
	v.LoggingData = append(v.LoggingData, float64(s.Log))[1:]
	v.LoggingChart.Data = v.LoggingData

	return nil
}

func (v *View) Close() {
	ui.Close()
}
