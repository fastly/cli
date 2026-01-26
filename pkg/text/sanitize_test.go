package text

import "testing"

func TestSanitizeTerminalOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text unchanged",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "preserves newlines and tabs",
			input:    "line1\nline2\ttabbed",
			expected: "line1\nline2\ttabbed",
		},
		{
			name:     "escapes color codes",
			input:    "\x1b[31mRED\x1b[0m",
			expected: "\\x1b[31mRED\\x1b[0m",
		},
		{
			name:     "escapes bold and other SGR codes",
			input:    "\x1b[1mbold\x1b[0m \x1b[4munderline\x1b[0m",
			expected: "\\x1b[1mbold\\x1b[0m \\x1b[4munderline\\x1b[0m",
		},
		{
			name:     "escapes cursor movement",
			input:    "before\x1b[2Aafter",
			expected: "before\\x1b[2Aafter",
		},
		{
			name:     "escapes screen clear",
			input:    "\x1b[2Jcleared",
			expected: "\\x1b[2Jcleared",
		},
		{
			name:     "escapes window title manipulation (OSC)",
			input:    "\x1b]0;malicious title\x07content",
			expected: "\\x1b]0;malicious title\\x07content",
		},
		{
			name:     "escapes multiple escape sequences",
			input:    "\x1b[31m\x1b[1mred bold\x1b[0m normal \x1b[32mgreen\x1b[0m",
			expected: "\\x1b[31m\\x1b[1mred bold\\x1b[0m normal \\x1b[32mgreen\\x1b[0m",
		},
		{
			name:     "escapes VCL content with escape sequences",
			input:    "sub vcl_recv { # \x1b[31mRED\x1b[0m }",
			expected: "sub vcl_recv { # \\x1b[31mRED\\x1b[0m }",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "escapes cursor position codes",
			input:    "\x1b[10;20Htext at position",
			expected: "\\x1b[10;20Htext at position",
		},
		{
			name:     "escapes erase codes",
			input:    "\x1b[Kerase line\x1b[Jclear below",
			expected: "\\x1b[Kerase line\\x1b[Jclear below",
		},
		{
			name:     "escapes standalone BEL",
			input:    "before\x07after",
			expected: "before\\x07after",
		},
		{
			name:     "escapes backspace",
			input:    "secret\x08visible",
			expected: "secret\\x08visible",
		},
		{
			name:     "escapes NUL character",
			input:    "before\x00after",
			expected: "before\\x00after",
		},
		{
			name:     "escapes form feed",
			input:    "page1\x0cpage2",
			expected: "page1\\x0cpage2",
		},
		{
			name:     "escapes vertical tab",
			input:    "line1\x0bline2",
			expected: "line1\\x0bline2",
		},
		{
			name:     "escapes DEL character",
			input:    "before\x7fafter",
			expected: "before\\x7fafter",
		},
		{
			name:     "preserves tab newline carriage return",
			input:    "col1\tcol2\nline2\r\nline3",
			expected: "col1\tcol2\nline2\r\nline3",
		},
		{
			name:     "escapes mixed control characters",
			input:    "\x00\x07\x08text\x0b\x0c\x1a\x7f",
			expected: "\\x00\\x07\\x08text\\x0b\\x0c\\x1a\\x7f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeTerminalOutput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeTerminalOutput(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
