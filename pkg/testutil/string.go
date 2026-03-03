package testutil

import "strings"

// StripNewLines removes all newline delimiters.
func StripNewLines(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}

// StrPtr is used to obtain the address of a literal string.
func StrPtr(s string) *string { return &s }
