package testutil

import "strings"

// StripNewLines removes all newline delimiters.
func StripNewLines(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}
