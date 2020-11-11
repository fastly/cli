package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/go-fastly/fastly"
	"github.com/segmentio/textio"
)

// ServiceType is a utility function which returns the given type string if
// non-empty otherwise returns the default `vcl`. This should be used unitl the
// API properly returns Service.Type for non-wasm services.
// TODO(phamann): remove once API returns correct type.
func ServiceType(t string) string {
	st := "vcl"
	if t != "" {
		st = t
	}
	return st
}

// PrintService pretty prints a fastly.Service structure in verbose format
// to a given io.Writer. Consumers can provide a prefix string which will
// be used as a prefix to each line, useful for indentation.
func PrintService(out io.Writer, prefix string, s *fastly.Service) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", s.ID)
	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "Type: %s\n", ServiceType(s.Type))
	if s.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", s.Comment)
	}
	fmt.Fprintf(out, "Customer ID: %s\n", s.CustomerID)
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(common.TimeFormat))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(common.TimeFormat))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(common.TimeFormat))
	}
	fmt.Fprintf(out, "Active version: %d\n", s.ActiveVersion)
	fmt.Fprintf(out, "Versions: %d\n", len(s.Versions))
	for j, version := range s.Versions {
		fmt.Fprintf(out, "\tVersion %d/%d\n", j+1, len(s.Versions))
		PrintVersion(out, "\t\t", version)
	}
}

// PrintServiceDetail pretty prints a fastly.ServiceDetail structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintServiceDetail(out io.Writer, indent string, s *fastly.ServiceDetail) {
	out = textio.NewPrefixWriter(out, indent)

	// Initally services have no active version, however go-fastly still
	// returns an empty Version struct with nil values. Which isn't useful for
	// output rendering.
	activeVersion := "none"
	if s.ActiveVersion.Active {
		activeVersion = fmt.Sprintf("%d", s.ActiveVersion.Number)
	}

	fmt.Fprintf(out, "ID: %s\n", s.ID)
	fmt.Fprintf(out, "Name: %s\n", s.Name)
	fmt.Fprintf(out, "Type: %s\n", ServiceType(s.Type))
	if s.Comment != "" {
		fmt.Fprintf(out, "Comment: %s\n", s.Comment)
	}
	fmt.Fprintf(out, "Customer ID: %s\n", s.CustomerID)
	if s.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(common.TimeFormat))
	}
	if s.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(common.TimeFormat))
	}
	if s.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(common.TimeFormat))
	}
	if s.ActiveVersion.Active {
		fmt.Fprintf(out, "Active version:\n")
		PrintVersion(out, "\t", &s.ActiveVersion)
	} else {
		fmt.Fprintf(out, "Active version: %s\n", activeVersion)
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
		fmt.Fprintf(out, "Created (UTC): %s\n", v.CreatedAt.UTC().Format(common.TimeFormat))
	}
	if v.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", v.UpdatedAt.UTC().Format(common.TimeFormat))
	}
	if v.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", v.DeletedAt.UTC().Format(common.TimeFormat))
	}
}
