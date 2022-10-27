package text

import (
	"fmt"
	"io"
)

type Lines map[string]interface{}

func PrintLines(out io.Writer, lines Lines) {
	for k, v := range lines {
		fmt.Fprintf(out, "%s: %+v\n", k, v)
	}
}
