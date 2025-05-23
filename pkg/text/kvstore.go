package text

import (
	"fmt"
	"io"

	"github.com/segmentio/textio"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/time"
)

// PrintKVStore pretty prints a fastly.Dictionary structure in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintKVStore(out io.Writer, prefix string, k *fastly.KVStore) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "\nID: %s\n", k.StoreID)
	fmt.Fprintf(out, "Name: %s\n", k.Name)
	fmt.Fprintf(out, "Created (UTC): %s\n", k.CreatedAt.UTC().Format(time.Format))
	fmt.Fprintf(out, "Last edited (UTC): %s\n", k.UpdatedAt.UTC().Format(time.Format))
}

// PrintKVStoreKeys pretty prints a list of kv store keys in verbose
// format to a given io.Writer. Consumers can provide a prefix string which
// will be used as a prefix to each line, useful for indentation.
func PrintKVStoreKeys(out io.Writer, prefix string, keys []string) {
	out = textio.NewPrefixWriter(out, prefix)

	for _, k := range keys {
		fmt.Fprintf(out, "Key: %s\n", k)
	}
}

// PrintKVStoreKeyValue pretty prints a value from an kv store to a
// given io.Writer. Consumers can provide a prefix string which will be used as
// a prefix to each line, useful for indentation.
func PrintKVStoreKeyValue(out io.Writer, prefix string, key, value string) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "Key: %s\n", key)
	fmt.Fprintf(out, "Value: %q\n", value)
}
