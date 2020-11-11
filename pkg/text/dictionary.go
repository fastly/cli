package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/go-fastly/fastly"
	"github.com/segmentio/textio"
)

// PrintDictionary pretty prints a fastly.Dictionary structure in verbose
// format to a given io.Writer. Consumers can provider a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintDictionary(out io.Writer, prefix string, d *fastly.Dictionary) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintf(out, "Name: %s\n", d.Name)
	fmt.Fprintf(out, "Write Only: %t\n", d.WriteOnly)
	fmt.Fprintf(out, "Created (UTC): %s\n", d.CreatedAt.UTC().Format(common.TimeFormat))
	fmt.Fprintf(out, "Last edited (UTC): %s\n", d.UpdatedAt.UTC().Format(common.TimeFormat))
	if d.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", d.DeletedAt.UTC().Format(common.TimeFormat))
	}
}
