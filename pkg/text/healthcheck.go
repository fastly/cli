package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/segmentio/textio"
)

// PrintHealthCheck pretty prints a fastly.HealthCheck structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintHealthCheck(out io.Writer, prefix string, h *fastly.HealthCheck) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Name: %s\n", h.Name)
	fmt.Fprintf(out, "Comment: %s\n", h.Comment)
	fmt.Fprintf(out, "Method: %s\n", h.Method)
	fmt.Fprintf(out, "Host: %s\n", h.Host)
	fmt.Fprintf(out, "Path: %s\n", h.Path)
	fmt.Fprintf(out, "HTTP version: %s\n", h.HTTPVersion)
	fmt.Fprintf(out, "Timeout: %d\n", h.Timeout)
	fmt.Fprintf(out, "Check interval: %d\n", h.CheckInterval)
	fmt.Fprintf(out, "Expected response: %d\n", h.ExpectedResponse)
	fmt.Fprintf(out, "Window: %d\n", h.Window)
	fmt.Fprintf(out, "Threshold: %d\n", h.Threshold)
	fmt.Fprintf(out, "Initial: %d\n", h.Initial)
}
