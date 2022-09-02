package text

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

var (
	lineStyle   = Reset
	headerStyle = Bold
)

// Table wraps an instance of a tabwriter and provides helper methods to easily
// create a table, add a header, add rows and print to the writer.
type Table struct {
	writer *tabwriter.Writer
}

// NewTable contructs a new Table.
func NewTable(w io.Writer) *Table {
	return &Table{
		writer: tabwriter.NewWriter(w, 0, 2, 2, ' ', 0),
	}
}

// AddLine writes a new row to the table.
func (t *Table) AddLine(args ...any) {
	var b strings.Builder
	for i := range args {
		b.WriteString(lineStyle(`%v`))
		if i+1 != len(args) {
			b.WriteString("\t")
		}
	}
	b.WriteString("\n")
	fmt.Fprintf(t.writer, b.String(), args...)
}

// AddHeader writes a table header line.
func (t *Table) AddHeader(args ...any) {
	var b strings.Builder
	for i := range args {
		b.WriteString(headerStyle(`%s`))
		if i+1 != len(args) {
			b.WriteString("\t")
		}
	}
	b.WriteString("\n")
	fmt.Fprintf(t.writer, b.String(), args...)
}

// Print writes the table to the writer.
func (t *Table) Print() {
	t.writer.Flush()
}
