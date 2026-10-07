package path

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"

	"github.com/fastly/cli/pkg/text"
)

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, data []paths.Data) {
	t := text.NewTable(out)
	t.AddHeader("PATH", "PATH ID", "CREATED AT", "UPDATED AT")
	for _, d := range data {
		var updatedAt string
		if d.UpdatedAt != nil {
			updatedAt = d.UpdatedAt.String()
		}
		t.AddLine(d.Path, d.PathID, d.CreatedAt, updatedAt)
	}
	t.Print()
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, data []paths.Data) {
	for _, d := range data {
		fmt.Fprintf(out, "Path: %s\n", d.Path)
		fmt.Fprintf(out, "Path ID: %s\n", d.PathID)
		fmt.Fprintf(out, "Created at: %s\n", d.CreatedAt)
		if d.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", *d.UpdatedAt)
		}
		fmt.Fprintf(out, "\n")
	}
}
