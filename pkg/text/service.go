package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v6/fastly"
	"github.com/segmentio/textio"
)

// PrintService pretty prints a fastly.Service structure in verbose format
// to a given io.Writer. Consumers can provide a prefix string which will
// be used as a prefix to each line, useful for indentation.
func PrintService(out io.Writer, prefix string, s *fastly.Service) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", s.ID)
	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "Type: %s\n", s.Type)
	if s.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", s.Comment)
	}
	fmt.Fprintf(out, "Customer ID: %s\n", s.CustomerID)
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(time.Format))
	}
	fmt.Fprintf(out, "Active version: %d\n", s.ActiveVersion)
	fmt.Fprintf(out, "Versions: %d\n", len(s.Versions))
	for j, version := range s.Versions {
		fmt.Fprintf(out, "\tVersion %d/%d\n", j+1, len(s.Versions))
		PrintVersion(out, "\t\t", version)
	}
}

// PrintVersion pretty prints a fastly.Version structure in verbose format to a
// given io.Writer. Consumers can provide a prefix string which will be used
// as a prefix to each line, useful for indentation.
func PrintVersion(out io.Writer, indent string, v *fastly.Version) {
	out = textio.NewPrefixWriter(out, indent)

	fmt.Fprintf(out, "Number: %d\n", v.Number)
	if v.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", v.Comment)
	}
	fmt.Fprintf(out, "Service ID: %s\n", v.ServiceID)
	fmt.Fprintf(out, "Active: %v\n", v.Active)
	fmt.Fprintf(out, "Locked: %v\n", v.Locked)
	fmt.Fprintf(out, "Deployed: %v\n", v.Deployed)
	fmt.Fprintf(out, "Staging: %v\n", v.Staging)
	fmt.Fprintf(out, "Testing: %v\n", v.Testing)
	if v.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", v.CreatedAt.UTC().Format(time.Format))
	}
	if v.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", v.UpdatedAt.UTC().Format(time.Format))
	}
	if v.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", v.DeletedAt.UTC().Format(time.Format))
	}
}
