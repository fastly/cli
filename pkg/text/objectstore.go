package text

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v6/fastly"
	"github.com/segmentio/textio"
)

// PrintObjectStore pretty prints a fastly.Dictionary structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintObjectStore(out io.Writer, prefix string, d *fastly.ObjectStore) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", d.ID)
	fmt.Fprintf(out, "Name: %s\n", d.Name)
	fmt.Fprintf(out, "Created (UTC): %s\n", d.CreatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Last edited (UTC): %s\n", d.UpdatedAt.UTC().Format(time.Format))
}

// PrintObjectStoreKeys pretty prints a list of object store keys in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintObjectStoreKeys(out io.Writer, prefix string, keys []string) {
	out = textio.NewPrefixWriter(out, prefix)

	for _, k := range keys {
		fmt.Fprintf(out, "Key: %s\n", k)
	}
}

// PrintObjectStoreKeyValue pretty prints a value from an object store to a
// given io.Writer. Consumers can provide a prefix string which will be used as
// a prefix to each line, useful for indentation.
func PrintObjectStoreKeyValue(out io.Writer, prefix string, key, value string) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Key: %s\n", key)
	fmt.Fprintf(out, "Value: %q\n", value)
}
