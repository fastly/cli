package text

import (
	"fmt"
	"io"

	"github.com/segmentio/textio"

	"github.com/fastly/go-fastly/v15/fastly/dns/v1/tsigkeys"
)

// PrintTSIGKey pretty prints a tsigkeys.TSIGKey in verbose format to a given
// io.Writer. Consumers can provide a prefix string for indentation.
func PrintTSIGKey(out io.Writer, prefix string, k *tsigkeys.TSIGKey) {
	out = textio.NewPrefixWriter(out, prefix)

	fmt.Fprintf(out, "ID: %s\n", strOrEmpty(k.ID))
	fmt.Fprintf(out, "Name: %s\n", strOrEmpty(k.Name))
	fmt.Fprintf(out, "Algorithm: %s\n", strOrEmpty(k.Algorithm))
	fmt.Fprintf(out, "Description: %s\n", strOrEmpty(k.Description))
	fmt.Fprintf(out, "Created at: %s\n", strOrEmpty(k.CreatedAt))
	fmt.Fprintf(out, "Updated at: %s\n", strOrEmpty(k.UpdatedAt))
}

// PrintTSIGKeyTbl prints a slice of tsigkeys.TSIGKey in table format to a given io.Writer.
func PrintTSIGKeyTbl(out io.Writer, keys []tsigkeys.TSIGKey) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Name", "Algorithm", "Description", "Created At", "Updated At")

	if keys == nil {
		tbl.Print()
		return
	}

	for _, k := range keys {
		tbl.AddLine(strOrEmpty(k.ID), strOrEmpty(k.Name), strOrEmpty(k.Algorithm), strOrEmpty(k.Description), strOrEmpty(k.CreatedAt), strOrEmpty(k.UpdatedAt))
	}
	tbl.Print()
}
