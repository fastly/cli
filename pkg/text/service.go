package text

import (
	"fmt"
	"io"
	"regexp"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/segmentio/textio"

	"github.com/fastly/cli/pkg/time"
)

// PrintService pretty prints a fastly.Service structure in verbose format
// to a given io.Writer. Consumers can provide a prefix string which will
// be used as a prefix to each line, useful for indentation.
func PrintService(out io.Writer, prefix string, s *fastly.Service) {
	out = textio.NewPrefixWriter(out, prefix)

	if s.ServiceID != nil {
		fmt.Fprintf(out, "ID: %s\n", fastly.ToValue(s.ServiceID))
	}
	if s.Name != nil {
		fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(s.Name))
	}
	if s.Type != nil {
		fmt.Fprintf(out, "Type: %s\n", fastly.ToValue(s.Type))
	}
	if s.Comment != nil {
		fmt.Fprintf(out, "Comment: %s\n", fastly.ToValue(s.Comment))
	}
	if s.CustomerID != nil {
		fmt.Fprintf(out, "Customer ID: %s\n", fastly.ToValue(s.CustomerID))
	}
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(time.Format))
	}
	if s.ActiveVersion != nil {
		fmt.Fprintf(out, "Active version: %d\n", fastly.ToValue(s.ActiveVersion))
	}
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

	if v.Number != nil {
		fmt.Fprintf(out, "Number: %d\n", fastly.ToValue(v.Number))
	}
	if v.Comment != nil {
		fmt.Fprintf(out, "Comment: %s\n", fastly.ToValue(v.Comment))
	}
	if v.ServiceID != nil {
		fmt.Fprintf(out, "Service ID: %s\n", fastly.ToValue(v.ServiceID))
	}
	if v.Active != nil {
		fmt.Fprintf(out, "Active: %v\n", fastly.ToValue(v.Active))
	}
	if v.Locked != nil {
		fmt.Fprintf(out, "Locked: %v\n", fastly.ToValue(v.Locked))
	}
	if v.Deployed != nil {
		fmt.Fprintf(out, "Deployed: %v\n", fastly.ToValue(v.Deployed))
	}
	if v.Staging != nil {
		fmt.Fprintf(out, "Staging: %v\n", fastly.ToValue(v.Staging))
	}
	if v.Testing != nil {
		fmt.Fprintf(out, "Testing: %v\n", fastly.ToValue(v.Testing))
	}
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

var fastlyIDRegEx = regexp.MustCompile("^[0-9a-zA-Z]{22}$")

// IsFastlyID determines if a string looks like a Fastly ID.
func IsFastlyID(s string) bool {
	return fastlyIDRegEx.Match([]byte(s))
}
