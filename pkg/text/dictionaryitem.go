package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/go-fastly/fastly"
	"github.com/segmentio/textio"
)

// PrintDictionaryItem pretty prints a fastly.DictionaryInfo structure in verbose
// format to a given io.Writer. Consumers can provider an prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintDictionaryItem(out io.Writer, prefix string, d *fastly.DictionaryItem) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Dictionary ID: %s\n", d.DictionaryID)
	fmt.Fprintf(out, "Item Key: %s\n", d.ItemKey)
	fmt.Fprintf(out, "Item Value: %s\n", d.ItemValue)
	fmt.Fprintf(out, "Created (UTC): %s\n", d.CreatedAt.UTC().Format(common.TimeFormat))
	fmt.Fprintf(out, "Last edited (UTC): %s\n", d.UpdatedAt.UTC().Format(common.TimeFormat))
	if d.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted (UTC): %s\n", d.DeletedAt.UTC().Format(common.TimeFormat))
	}
}
