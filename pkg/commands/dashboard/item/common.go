package item

var (
	sourceTypes        = []string{"stats.domain", "stats.edge", "stats.origin"}
	visualizationTypes = []string{"chart"}
	plotTypes          = []string{"bar", "donut", "line", "single-metric"}
	calculationMethods = []string{"avg", "sum", "min", "max", "latest", "p95"}
	formats            = []string{"number", "bytes", "percent", "requests", "responses", "seconds", "milliseconds", "ratio", "bitrate"}
)
