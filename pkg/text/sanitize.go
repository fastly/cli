package text

import (
	"fmt"
	"strings"
)

// SanitizeTerminalOutput escapes control characters from untrusted content
// to prevent terminal injection attacks.
func SanitizeTerminalOutput(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t', r == '\n', r == '\r':
			b.WriteRune(r)
		case r < 0x20 || r == 0x7F:
			fmt.Fprintf(&b, "\\x%02x", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
