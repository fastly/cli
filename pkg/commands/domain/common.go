package domain

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v13/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/text"
)

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, data []domains.Data) {
	t := text.NewTable(out)
	t.AddHeader("FQDN", "DOMAIN ID", "SERVICE ID", "CREATED AT", "UPDATED AT", "DESCRIPTION")
	for _, d := range data {
		var sid string
		if d.ServiceID != nil {
			sid = *d.ServiceID
		}
		t.AddLine(d.FQDN, d.DomainID, sid, d.CreatedAt, d.UpdatedAt, d.Description)
	}
	t.Print()
}

// printSummary displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, data []domains.Data) {
	for _, d := range data {
		fmt.Fprintf(out, "FQDN: %s\n", d.FQDN)
		fmt.Fprintf(out, "Domain ID: %s\n", d.DomainID)
		if d.ServiceID != nil {
			fmt.Fprintf(out, "Service ID: %s\n", *d.ServiceID)
		}
		fmt.Fprintf(out, "Created at: %s\n", d.CreatedAt)
		fmt.Fprintf(out, "Updated at: %s\n", d.UpdatedAt)
		fmt.Fprintf(out, "Description: %s\n", d.Description)
		fmt.Fprintf(out, "\n")
	}
}
