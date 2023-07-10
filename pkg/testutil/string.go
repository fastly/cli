package testutil

import "strings"

func StripNewLines(s string) string {
	return strings.ReplaceAll(s, "\n", "")
}
