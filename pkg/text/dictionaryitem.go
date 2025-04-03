package text

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/segmentio/textio"

	"github.com/fastly/cli/pkg/time"
)

// PrintDictionaryItem pretty prints a fastly.DictionaryInfo structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintDictionaryItem(out io.Writer, prefix string, d *fastly.DictionaryItem) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Dictionary ID: %s\n", fastly.ToValue(d.DictionaryID))
	fmt.Fprintf(out, "Item Key: %s\n", fastly.ToValue(d.ItemKey))
	fmt.Fprintf(out, "Item Value: %s\n", fastly.ToValue(d.ItemValue))
	if d.CreatedAt != nil {
		fmt.Fprintf(out, "Created (UTC): %s\n", d.CreatedAt.UTC().Format(time.Format))
	}
	if d.UpdatedAt != nil {
		fmt.Fprintf(out, "Last edited (UTC): %s\n", d.UpdatedAt.UTC().Format(time.Format))
	}
	if d.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", d.DeletedAt.UTC().Format(time.Format))
	}
}

// PrintDictionaryItemKV pretty prints only the key/value pairs from a dictionary item.
func PrintDictionaryItemKV(out io.Writer, prefix string, d *fastly.DictionaryItem) {
	out = textio.NewPrefixWriter(out, prefix)
	fmt.Fprintf(out, "Item Key: %s\n", fastly.ToValue(d.ItemKey))
	fmt.Fprintf(out, "Item Value: %s\n", fastly.ToValue(d.ItemValue))
}
