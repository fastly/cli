package routingconfig

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"

	"github.com/fastly/cli/pkg/text"
)

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, data []routingconfigs.Data) {
	t := text.NewTable(out)
	t.AddHeader("NAME", "ROUTING CONFIG ID", "STATE", "CREATED AT", "UPDATED AT")
	for _, d := range data {
		var updatedAt string
		if d.UpdatedAt != nil {
			updatedAt = d.UpdatedAt.String()
		}
		t.AddLine(d.Name, d.RoutingConfigID, d.State, d.CreatedAt, updatedAt)
	}
	t.Print()
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, data []routingconfigs.Data) {
	for _, d := range data {
		fmt.Fprintf(out, "Name: %s\n", d.Name)
		fmt.Fprintf(out, "Routing Config ID: %s\n", d.RoutingConfigID)
		fmt.Fprintf(out, "State: %s\n", d.State)
		fmt.Fprintf(out, "Created at: %s\n", d.CreatedAt)
		if d.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", *d.UpdatedAt)
		}
		if d.ActivatedAt != nil {
			fmt.Fprintf(out, "Activated at: %s\n", *d.ActivatedAt)
		}
		fmt.Fprintf(out, "\n")
	}
}
