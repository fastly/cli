package domainv1

import (
	"fmt"
	"io"

	v1 "github.com/fastly/go-fastly/v9/fastly/domains/v1"

	"github.com/fastly/cli/pkg/text"
)

// printSummary displays the information returned from the API in a summarised
// format.
func printSummary(out io.Writer, data []v1.Data) {
	t := text.NewTable(out)
	t.AddHeader("FQDN", "DOMAIN ID", "SERVICE ID", "CREATED AT", "UPDATED AT")
	for _, d := range data {
		var sid string
		if d.ServiceID != nil {
			sid = *d.ServiceID
		}
		t.AddLine(d.FQDN, d.DomainID, sid, d.CreatedAt, d.UpdatedAt)
	}
	t.Print()
}

// printSummary displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, data []v1.Data) {
	for _, d := range data {
		fmt.Fprintf(out, "FQDN: %s\n", d.FQDN)
		fmt.Fprintf(out, "Domain ID: %s\n", d.DomainID)
		if d.ServiceID != nil {
			fmt.Fprintf(out, "Service ID: %s\n", *d.ServiceID)
		}
		fmt.Fprintf(out, "Created at: %s\n", d.CreatedAt)
		fmt.Fprintf(out, "Updated at: %s\n", d.UpdatedAt)
		fmt.Fprintf(out, "\n")
	}
}
