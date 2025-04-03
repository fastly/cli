package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/segmentio/textio"
)

// PrintHealthCheck pretty prints a fastly.HealthCheck structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintHealthCheck(out io.Writer, prefix string, h *fastly.HealthCheck) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(h.Name))
	fmt.Fprintf(out, "Comment: %s\n", fastly.ToValue(h.Comment))
	fmt.Fprintf(out, "Method: %s\n", fastly.ToValue(h.Method))
	fmt.Fprintf(out, "Host: %s\n", fastly.ToValue(h.Host))
	fmt.Fprintf(out, "Path: %s\n", fastly.ToValue(h.Path))
	fmt.Fprintf(out, "HTTP version: %s\n", fastly.ToValue(h.HTTPVersion))
	fmt.Fprintf(out, "Timeout: %d\n", fastly.ToValue(h.Timeout))
	fmt.Fprintf(out, "Check interval: %d\n", fastly.ToValue(h.CheckInterval))
	fmt.Fprintf(out, "Expected response: %d\n", fastly.ToValue(h.ExpectedResponse))
	fmt.Fprintf(out, "Window: %d\n", fastly.ToValue(h.Window))
	fmt.Fprintf(out, "Threshold: %d\n", fastly.ToValue(h.Threshold))
	fmt.Fprintf(out, "Initial: %d\n", fastly.ToValue(h.Initial))
}
