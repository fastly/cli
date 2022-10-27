package text

import (
	"fmt"
	"io"
	"sort"
)

// Lines is the struct that is used by PrintLines
type Lines map[string]interface{}

// PrintLines pretty prints a Lines struct with one item per line.
// The map is sorted before printing and a newline is added at the beginning
func PrintLines(out io.Writer, lines Lines) {
	keys := make([]string, 0, len(lines))
	for k := range lines {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(out, "\n")
	for _, k := range keys {
		fmt.Fprintf(out, "%s: %+v\n", k, lines[k])
	}
}
