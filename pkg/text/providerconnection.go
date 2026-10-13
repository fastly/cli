package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"
)

// PrintProviderConnection displays a provider connection.
func PrintProviderConnection(out io.Writer, pc *providerconnection.ProviderConnection) {
	fmt.Fprintf(out, "ID: %s\n", pc.ID)
	fmt.Fprintf(out, "Name: %s\n", pc.Name)
	fmt.Fprintf(out, "Models: %s\n", strings.Join(pc.Models, ", "))
	fmt.Fprintf(out, "Base URL: %s\n", pc.BaseURL)
	fmt.Fprintf(out, "Created At: %s\n", pc.CreatedAt)
	fmt.Fprintf(out, "Updated At: %s\n", pc.UpdatedAt)
}

// PrintProviderConnectionsTbl displays provider connections in a table format.
func PrintProviderConnectionsTbl(out io.Writer, connectionsToPrint []providerconnection.ProviderConnection) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Models", "Base URL", "Created At")

	if connectionsToPrint == nil {
		tbl.Print()
		return
	}

	for _, pc := range connectionsToPrint {
		tbl.AddLine(pc.ID, pc.Name, strings.Join(pc.Models, ", "), pc.BaseURL, pc.CreatedAt)
	}
	tbl.Print()
}
