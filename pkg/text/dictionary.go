package text

import (
	"fmt"
	"io"

	"github.com/segmentio/textio"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/time"
)

// PrintDictionary pretty prints a fastly.Dictionary structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintDictionary(out io.Writer, prefix string, d *fastly.Dictionary) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", fastly.ToValue(d.DictionaryID))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(d.Name))
	fmt.Fprintf(out, "Write Only: %t\n", fastly.ToValue(d.WriteOnly))
	fmt.Fprintf(out, "Created (UTC): %s\n", d.CreatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Last edited (UTC): %s\n", d.UpdatedAt.UTC().Format(time.Format))
	if d.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", d.DeletedAt.UTC().Format(time.Format))
	}
}
