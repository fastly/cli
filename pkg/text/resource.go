package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/segmentio/textio"
)

// PrintResource pretty prints a fastly.Resource structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintResource(out io.Writer, prefix string, r *fastly.Resource) {
	if r == nil {
		return
	}
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", r.ID)
	fmt.Fprintf(out, "Name: %s\n", r.Name)
	fmt.Fprintf(out, "Service ID: %s\n", r.ServiceID)
	fmt.Fprintf(out, "Service Version: %s\n", r.ServiceVersion)
	fmt.Fprintf(out, "Resource ID: %s\n", r.ResourceID)
	fmt.Fprintf(out, "Resource Type: %s\n", r.ResourceType)

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", r.CreatedAt.UTC().Format(time.Format))
	}
	if r.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", r.UpdatedAt.UTC().Format(time.Format))
	}
	if r.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", r.DeletedAt.UTC().Format(time.Format))
	}
}
