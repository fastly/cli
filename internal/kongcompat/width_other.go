//go:build appengine || !(linux || freebsd || darwin || dragonfly || netbsd || openbsd)

package kongcompat

import (
	"io"
	"os"
	"strconv"
)

// guessWidth honors COLUMNS and otherwise assumes 80 columns.
func guessWidth(io.Writer) int {
	if cols, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil {
		return cols
	}
	return 80
}
